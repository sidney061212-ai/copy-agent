package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"

	"github.com/copyagent/copyagentd/internal/agent"
	"github.com/copyagent/copyagentd/internal/clipboard"
	"github.com/copyagent/copyagentd/internal/config"
	"github.com/copyagent/copyagentd/internal/core"
	"github.com/copyagent/copyagentd/internal/diagnostics"
	"github.com/copyagent/copyagentd/internal/inject"
	"github.com/copyagent/copyagentd/internal/service"
	"github.com/copyagent/copyagentd/internal/transport/feishu"
)

type copyRequest struct {
	Text string `json:"text"`
}

func main() {
	logCloser, daemonLogEnabled, err := initDaemonLogging()
	if err != nil {
		fmt.Fprintln(os.Stderr, "copyagentd:", err)
		os.Exit(1)
	}
	if logCloser != nil {
		defer logCloser.Close()
	}
	defer func() {
		if err := recover(); err != nil {
			log.Printf("copyagentd panic: %v\n%s", err, string(debug.Stack()))
			os.Exit(2)
		}
	}()
	if err := run(os.Args[1:]); err != nil {
		if daemonLogEnabled {
			log.Printf("copyagentd fatal: %v", err)
		} else {
			fmt.Fprintln(os.Stderr, "copyagentd:", err)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	configPath := service.ConfigPath()
	command := "help"
	if len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "help", "--help", "-h":
		printHelp()
		return nil
	case "doctor":
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		for _, check := range diagnostics.Doctor(cfg) {
			mark := "✅"
			if !check.OK {
				mark = "❌"
			}
			fmt.Printf("%s %s: %s\n", mark, check.Name, check.Detail)
		}
		return nil
	case "copy":
		if len(args) < 2 {
			return errors.New("copy requires text")
		}
		text := core.ExtractCopyText(args[1])
		if !core.ValidText(text) {
			return errors.New("text is required")
		}
		return clipboard.WriteText(text)
	case "action":
		cfg, _ := config.Load(configPath)
		return runAction(context.Background(), cfg, args[1:])
	case "serve":
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		lock, err := service.AcquireInstanceLock(configPath)
		if err != nil {
			return err
		}
		defer lock.Release()
		return serve(cfg)
	case "service":
		return runService(configPath, args[1:])
	case "feishu-serve":
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		lock, err := service.AcquireInstanceLock(configPath)
		if err != nil {
			return err
		}
		defer lock.Release()
		ignoreHangupSignal()
		transport := feishu.NewAgentTransport(cfg.FeishuAppID, cfg.FeishuAppSecret)
		directCfg := agent.DirectHandlerConfig{
			Policy: agent.DirectPolicyConfig{
				AllowedUserIDs: cfg.AllowedActorIDs,
			},
			Planner: agent.DirectPlannerConfig{
				ImageAction: cfg.ImageAction,
			},
			Executor: agent.DirectExecutorConfig{
				ReplyEnabled:       cfg.ReplyEnabled,
				DefaultDownloadDir: cfg.DefaultDownloadDir,
				Clipboard: agent.TextClipboardWriterFunc(func(_ context.Context, text string) error {
					if err := clipboard.WriteText(text); err != nil {
						return err
					}
					current, err := clipboard.ReadText()
					if err != nil {
						log.Printf("feishu clipboard verify skipped: err=%v", err)
						return nil
					}
					if current != text {
						log.Printf("feishu clipboard verify mismatch: wrote_bytes=%d read_bytes=%d", len([]byte(text)), len([]byte(current)))
						return nil
					}
					log.Printf("feishu clipboard verified: bytes=%d", len([]byte(text)))
					return nil
				}),
				ImageClipboard: agent.ImageClipboardWriterFunc(func(ctx context.Context, path string) error {
					return clipboard.WritePNGFile(ctx, path)
				}),
			},
		}
		engine, err := buildFeishuEngine(cfg, transport, directCfg)
		if err != nil {
			return err
		}
		log.Printf("copyagentd feishu-serve starting mode=%s", engine.Name())
		err = engine.Start(context.Background())
		log.Printf("copyagentd feishu-serve stopped: %v", err)
		return err
	case "feishu-probe":
		cfg, err := config.Load("")
		if err != nil {
			return err
		}
		transport := feishu.NewTransport(cfg.FeishuAppID, cfg.FeishuAppSecret)
		fmt.Println("feishu transport initialized")
		if len(args) > 1 && args[1] == "--start" {
			return transport.Start(context.Background())
		}
		select {}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func initDaemonLogging() (io.Closer, bool, error) {
	logPath := os.Getenv(service.LogFileEnv)
	if strings.TrimSpace(logPath) == "" {
		return nil, false, nil
	}
	maxSize := int64(service.DefaultLogMaxSize)
	if value := strings.TrimSpace(os.Getenv(service.LogMaxSizeEnv)); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, false, fmt.Errorf("invalid %s: %w", service.LogMaxSizeEnv, err)
		}
		if parsed > 0 {
			maxSize = parsed
		}
	}
	writer, err := service.NewRotatingWriter(logPath, maxSize)
	if err != nil {
		return nil, false, err
	}
	log.SetOutput(writer)
	return writer, true, nil
}

func runAction(ctx context.Context, cfg config.Config, args []string) error {
	if len(args) == 0 {
		return errors.New("action requires inject-text, reply-text, status, or turn")
	}
	service := inject.NewDefaultService()
	switch normalizeActionCommand(args[0]) {
	case "status":
		target, err := service.Status(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("frontmost app=%q bundle=%q window=%q allowed=%t\n", target.AppName, target.BundleID, target.WindowTitle, inject.BundleAllowed(target.BundleID, inject.DefaultAllowedBundleIDs()))
		return nil
	case "turn":
		if len(args) < 2 {
			return errors.New("turn requires status or a target name")
		}
		if args[1] == "status" {
			target, err := service.Status(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("frontmost app=%q bundle=%q window=%q allowed=%t\n", target.AppName, target.BundleID, target.WindowTitle, inject.BundleAllowed(target.BundleID, inject.DefaultAllowedBundleIDs()))
			return nil
		}
		named, ok := inject.ResolveNamedTarget(args[1])
		if !ok {
			return fmt.Errorf("unknown target %q", args[1])
		}
		target, err := service.Activate(ctx, named.BundleID)
		if err != nil {
			return err
		}
		fmt.Printf("activated target=%q app=%q bundle=%q window=%q\n", named.Name, target.AppName, target.BundleID, target.WindowTitle)
		return nil
	case "inject-text":
		text, submit, err := parseActionInjectText(args[1:])
		if err != nil {
			return err
		}
		result, err := service.InjectText(ctx, inject.Request{Text: text, Submit: submit, AllowedBundleIDs: inject.DefaultAllowedBundleIDs(), MaxBytes: inject.DefaultMaxBytes})
		if err != nil {
			return err
		}
		fmt.Printf("injected app=%q bundle=%q restored_clipboard=%t\n", result.Target.AppName, result.Target.BundleID, result.RestoredClipboard)
		if result.Warning != "" {
			fmt.Println(result.Warning)
		}
		return nil
	case "reply-text":
		sessionKey, text, err := parseActionReplyText(args[1:])
		if err != nil {
			return err
		}
		if strings.TrimSpace(cfg.FeishuAppID) == "" || strings.TrimSpace(cfg.FeishuAppSecret) == "" {
			return errors.New("feishu credentials are required for reply-text")
		}
		replyCtx, err := feishu.ReconstructReplyContext(sessionKey)
		if err != nil {
			return err
		}
		transport := feishu.NewAgentTransport(cfg.FeishuAppID, cfg.FeishuAppSecret)
		if err := transport.Send(ctx, replyCtx, text); err != nil {
			return err
		}
		fmt.Println("sent")
		return nil
	default:
		return fmt.Errorf("unknown action command: %s", args[0])
	}
}

func normalizeActionCommand(name string) string {
	clean := strings.ToLower(strings.TrimSpace(name))
	if clean == "target" {
		return "turn"
	}
	return clean
}

func parseActionInjectText(args []string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, errors.New("inject-text requires --text or --stdin")
	}
	submit := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--submit":
			submit = true
		case "--stdin":
			data, err := os.ReadFile("/dev/stdin")
			if err != nil {
				return "", false, err
			}
			return string(data), submit, nil
		case "--text":
			if i+1 >= len(args) {
				return "", false, errors.New("--text requires a value")
			}
			return args[i+1], submit, nil
		}
	}
	return strings.Join(args, " "), submit, nil
}

func parseActionReplyText(args []string) (string, string, error) {
	var sessionKey string
	var text string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--session-key":
			if i+1 >= len(args) {
				return "", "", errors.New("--session-key requires a value")
			}
			sessionKey = args[i+1]
			i++
		case "--text":
			if i+1 >= len(args) {
				return "", "", errors.New("--text requires a value")
			}
			text = args[i+1]
			i++
		case "--stdin":
			data, err := os.ReadFile("/dev/stdin")
			if err != nil {
				return "", "", err
			}
			text = string(data)
		default:
			if text == "" {
				text = strings.Join(args[i:], " ")
				break
			}
		}
	}
	if strings.TrimSpace(sessionKey) == "" {
		return "", "", errors.New("reply-text requires --session-key")
	}
	if strings.TrimSpace(text) == "" {
		return "", "", errors.New("reply-text requires --text or --stdin")
	}
	return sessionKey, text, nil
}

func runService(configPath string, args []string) error {
	if len(args) == 0 {
		return errors.New("service requires install, uninstall, start, stop, restart, status, or logs")
	}
	switch args[0] {
	case "install":
		startNow := true
		for _, arg := range args[1:] {
			if arg == "--no-start" {
				startNow = false
				continue
			}
			return fmt.Errorf("unknown service install flag: %s", arg)
		}
		cfg, err := service.DefaultConfig("")
		if err != nil {
			return err
		}
		cfg.ConfigPath = configPath
		if err := service.Resolve(&cfg); err != nil {
			return err
		}
		if err := service.Install(cfg, startNow); err != nil {
			return err
		}
		if err := service.SaveMeta(&service.Meta{
			LogFile:     cfg.LogFile,
			LogMaxSize:  cfg.LogMaxSize,
			WorkDir:     cfg.WorkDir,
			BinaryPath:  cfg.BinaryPath,
			ConfigPath:  cfg.ConfigPath,
			InstalledAt: service.NowISO(),
		}); err != nil {
			log.Printf("service metadata save failed: %v", err)
		}
		if startNow {
			fmt.Println("copyagentd service installed and started.")
		} else {
			fmt.Println("copyagentd service installed.")
		}
		fmt.Println()
		fmt.Printf("  Platform:  %s\n", "launchd")
		fmt.Printf("  Binary:    %s\n", cfg.BinaryPath)
		fmt.Printf("  Config:    %s\n", cfg.ConfigPath)
		fmt.Printf("  WorkDir:   %s\n", cfg.WorkDir)
		fmt.Printf("  Log:       %s\n", cfg.LogFile)
		fmt.Printf("  LogMax:    %d MB\n", cfg.LogMaxSize/1024/1024)
		if !startNow {
			fmt.Println("  Start:     pending (use `copyagentd service start`)")
		}
		return nil
	case "uninstall":
		if err := service.Uninstall(); err != nil {
			return err
		}
		service.RemoveMeta()
		fmt.Println("copyagentd service uninstalled.")
		return nil
	case "start":
		if err := service.Start(); err != nil {
			return err
		}
		fmt.Println("copyagentd service started.")
		return nil
	case "stop":
		if err := service.Stop(); err != nil {
			return err
		}
		fmt.Println("copyagentd service stopped.")
		return nil
	case "restart":
		if err := service.Restart(); err != nil {
			return err
		}
		fmt.Println("copyagentd service restarted.")
		return nil
	case "status":
		status, err := service.Status()
		if err != nil {
			return err
		}
		fmt.Println("copyagentd service status")
		fmt.Println()
		if !status.Installed {
			fmt.Println("  Status:    Not installed")
			fmt.Printf("  Platform:  %s\n", status.Platform)
			return nil
		}
		state := "Installed (not running)"
		if status.Running {
			state = "Running"
		}
		fmt.Printf("  Status:    %s\n", state)
		fmt.Printf("  Platform:  %s\n", status.Platform)
		if status.PID > 0 {
			fmt.Printf("  PID:       %d\n", status.PID)
		}
		if meta, err := service.LoadMeta(); err == nil {
			if strings.TrimSpace(meta.BinaryPath) != "" {
				fmt.Printf("  Binary:    %s\n", meta.BinaryPath)
			}
			if strings.TrimSpace(meta.ConfigPath) != "" {
				fmt.Printf("  Config:    %s\n", meta.ConfigPath)
			}
			if strings.TrimSpace(meta.LogFile) != "" {
				fmt.Printf("  Log:       %s\n", meta.LogFile)
			}
		}
		return nil
	case "logs":
		return printServiceLogs()
	default:
		return fmt.Errorf("unknown service command: %s", args[0])
	}
}

func printServiceLogs() error {
	paths, err := service.LogFiles()
	if err != nil {
		return err
	}
	for _, path := range paths {
		text, err := service.TailFile(path, 50)
		if err != nil {
			return err
		}
		fmt.Printf("==> %s <==\n%s\n", path, text)
	}
	return nil
}

func ignoreHangupSignal() {
	signal.Ignore(syscall.SIGHUP)
}

func printHelp() {
	fmt.Println(helpText())
}

func helpText() string {
	return `copyagentd spike

Usage:
  copyagentd doctor
  copyagentd copy <text>
  copyagentd action inject-text [--submit] --text <text>
  copyagentd action inject-text [--submit] --stdin
  copyagentd action reply-text --session-key <key> --text <text>
  copyagentd action reply-text --session-key <key> --stdin
  copyagentd action status
  copyagentd action turn status|codex|claude|terminal|iterm|warp|vscode|cursor
  copyagentd serve
  copyagentd feishu-serve
  copyagentd service install|uninstall|start|stop|restart|status|logs`
}

func serve(cfg config.Config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "service": "copyagentd"})
	})
	mux.HandleFunc("/copy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Token != "" && r.Header.Get("Authorization") != "Bearer "+cfg.Token && r.URL.Query().Get("token") != cfg.Token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req copyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		text := core.ExtractCopyText(req.Text)
		if !core.ValidText(text) {
			http.Error(w, "text is required", http.StatusBadRequest)
			return
		}
		if err := clipboard.WriteText(text); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	addr := cfg.Host + ":" + strconv.Itoa(cfg.Port+1000)
	log.Printf("copyagentd spike listening on http://%s", addr)
	return http.ListenAndServe(addr, mux)
}
