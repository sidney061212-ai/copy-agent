package service

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const Label = "com.copyagent.copyagentd"

func DefaultConfig(executable string) (Config, error) {
	if runtime.GOOS != "darwin" {
		return Config{}, errors.New("launchd service is only supported on macOS")
	}
	cfg := Config{BinaryPath: executable}
	if err := Resolve(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func PlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", Label+".plist"), nil
}

func LaunchdPlist(cfg Config) string {
	var out bytes.Buffer
	_ = plistTemplate.Execute(&out, cfg)
	return out.String()
}

func Install(cfg Config, start bool) error {
	if err := Resolve(&cfg); err != nil {
		return err
	}
	plistPath, err := PlistPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0o755); err != nil {
		return err
	}

	_, _ = runLaunchctl("bootout", guiDomain()+"/"+Label)

	if err := os.WriteFile(plistPath, []byte(LaunchdPlist(cfg)), 0o644); err != nil {
		return err
	}
	if !start {
		return nil
	}
	if _, err := runLaunchctl("bootstrap", guiDomain(), plistPath); err != nil {
		return err
	}
	_, err = runLaunchctl("kickstart", "-kp", guiDomain()+"/"+Label)
	return err
}

func Uninstall() error {
	_, _ = runLaunchctl("bootout", guiDomain()+"/"+Label)
	plistPath, err := PlistPath()
	if err != nil {
		return err
	}
	if err := os.Remove(plistPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func Start() error {
	plistPath, err := PlistPath()
	if err != nil {
		return err
	}
	if _, err := runLaunchctl("bootstrap", guiDomain(), plistPath); err != nil {
		_, err = runLaunchctl("kickstart", "-kp", guiDomain()+"/"+Label)
		if err != nil {
			return err
		}
	}
	return nil
}

func Stop() error {
	_, err := runLaunchctl("bootout", guiDomain()+"/"+Label)
	return err
}

func Restart() error {
	_, _ = runLaunchctl("bootout", guiDomain()+"/"+Label)
	plistPath, err := PlistPath()
	if err != nil {
		return err
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		if _, err := runLaunchctl("bootstrap", guiDomain(), plistPath); err == nil {
			_, err = runLaunchctl("kickstart", "-kp", guiDomain()+"/"+Label)
			return err
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func Status() (*ServiceStatus, error) {
	status := &ServiceStatus{Platform: "launchd"}
	plistPath, err := PlistPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(plistPath); err != nil {
		return status, nil
	}
	status.Installed = true
	out, _ := runLaunchctl("print", guiDomain()+"/"+Label)
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "pid = ") {
			pid, err := strconv.Atoi(strings.TrimPrefix(trimmed, "pid = "))
			if err == nil && pid > 0 {
				status.PID = pid
				status.Running = true
			}
		}
		if strings.Contains(trimmed, "state = running") {
			status.Running = true
		}
	}
	return status, nil
}

func guiDomain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}

func runLaunchctl(args ...string) (string, error) {
	cmd := exec.Command("launchctl", args...)
	out, err := cmd.CombinedOutput()
	message := strings.TrimSpace(string(out))
	if err != nil {
		if message == "" {
			message = err.Error()
		}
		return message, errors.New(message)
	}
	return message, nil
}

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>` + Label + `</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.BinaryPath}}</string>
    <string>feishu-serve</string>
  </array>
  <key>WorkingDirectory</key>
  <string>{{.WorkDir}}</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <true/>
  </dict>
  <key>EnvironmentVariables</key>
  <dict>
    <key>` + ConfigPathEnv + `</key>
    <string>{{.ConfigPath}}</string>
    <key>` + LogFileEnv + `</key>
    <string>{{.LogFile}}</string>
    <key>` + LogMaxSizeEnv + `</key>
    <string>{{.LogMaxSize}}</string>
    <key>LANG</key>
    <string>en_US.UTF-8</string>
    <key>LC_ALL</key>
    <string>en_US.UTF-8</string>
    <key>PATH</key>
    <string>{{.EnvPATH}}</string>
  </dict>
  <key>StandardOutPath</key>
  <string>/dev/null</string>
  <key>StandardErrorPath</key>
  <string>/dev/null</string>
</dict>
</plist>
`))
