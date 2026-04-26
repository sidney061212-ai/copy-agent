package diagnostics

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/copyagent/copyagentd/internal/config"
)

type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

func Doctor(cfg config.Config, paths ...string) []Check {
	configPath := config.DefaultPath()
	if len(paths) > 0 && paths[0] != "" {
		configPath = paths[0]
	}
	checks := []Check{
		{Name: "config", OK: true, Detail: "loaded"},
		configPermissionsCheck(configPath),
		downloadDirCheck(cfg.DefaultDownloadDir),
		{Name: "mode", OK: cfg.Mode == "" || cfg.Mode == "feishu-bot", Detail: cfg.Mode},
		{Name: "clipboard", OK: clipboardAvailable(), Detail: runtime.GOOS},
		keystrokeAutomationCheck(),
	}
	return checks
}

func keystrokeAutomationCheck() Check {
	if runtime.GOOS != "darwin" {
		return Check{Name: "keystroke_automation", OK: true, Detail: runtime.GOOS}
	}
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to key code 63`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(out))
		if detail == "" {
			detail = err.Error()
		}
		return Check{Name: "keystroke_automation", OK: false, Detail: detail}
	}
	return Check{Name: "keystroke_automation", OK: true, Detail: "System Events keystroke allowed"}
}

func configPermissionsCheck(path string) Check {
	info, err := os.Stat(path)
	if err != nil {
		return Check{Name: "config_permissions", OK: false, Detail: err.Error()}
	}
	mode := info.Mode().Perm()
	return Check{Name: "config_permissions", OK: mode&0o077 == 0, Detail: fmt.Sprintf("%#o", mode)}
}

func downloadDirCheck(path string) Check {
	if path == "" {
		path = "~/Downloads/copyagent"
	}
	expanded := expandHome(path)
	info, err := os.Stat(expanded)
	if err == nil {
		return Check{Name: "download_dir", OK: info.IsDir() && writable(expanded), Detail: expanded}
	}
	parent := filepath.Dir(expanded)
	return Check{Name: "download_dir", OK: writable(parent), Detail: expanded}
}

func writable(path string) bool {
	file, err := os.CreateTemp(path, ".copyagent-write-check-*")
	if err != nil {
		return false
	}
	name := file.Name()
	_ = file.Close()
	_ = os.Remove(name)
	return true
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	return path
}

func clipboardAvailable() bool {
	name := "pbcopy"
	if runtime.GOOS == "windows" {
		name = "powershell"
	}
	_, err := exec.LookPath(name)
	return err == nil
}

func ProcessRSSKB() int64 {
	var stat os.FileInfo
	_ = stat
	return 0
}
