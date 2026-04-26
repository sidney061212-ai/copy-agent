//go:build darwin

package inject

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func NewDefaultService() *Service {
	service := NewService(macosInspector{}, macosPasteboard{}, macosKeystroker{})
	service.Activator = macosActivator{}
	return service
}

type macosInspector struct{}

func (macosInspector) Frontmost(ctx context.Context) (Target, error) {
	script := `tell application "System Events"
set frontApp to first application process whose frontmost is true
set appName to name of frontApp
set bundleID to bundle identifier of frontApp
set windowTitle to ""
try
  set windowTitle to name of front window of frontApp
end try
return appName & linefeed & bundleID & linefeed & windowTitle
end tell`
	out, err := runOSAScript(ctx, script)
	if err != nil {
		return Target{}, err
	}
	parts := strings.SplitN(strings.TrimRight(out, "\n"), "\n", 3)
	for len(parts) < 3 {
		parts = append(parts, "")
	}
	return Target{AppName: parts[0], BundleID: parts[1], WindowTitle: parts[2]}, nil
}

type macosPasteboard struct{}

func (macosPasteboard) Snapshot(ctx context.Context) (Snapshot, error) {
	cmd := exec.CommandContext(ctx, "pbpaste")
	out, err := cmd.Output()
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Text: string(out), OK: true}, nil
}

func (macosPasteboard) WriteText(ctx context.Context, text string) error {
	cmd := exec.CommandContext(ctx, "pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func (pasteboard macosPasteboard) Restore(ctx context.Context, snapshot Snapshot) error {
	if !snapshot.OK {
		return nil
	}
	return pasteboard.WriteText(ctx, snapshot.Text)
}

type macosKeystroker struct{}

func (macosKeystroker) Paste(ctx context.Context) error {
	_, err := runOSAScript(ctx, `tell application "System Events" to keystroke "v" using command down`)
	return err
}

func (macosKeystroker) Submit(ctx context.Context) error {
	_, err := runOSAScript(ctx, `tell application "System Events" to key code 36`)
	return err
}

type macosActivator struct{}

func (macosActivator) ActivateBundle(ctx context.Context, bundleID string) error {
	quoted := strings.ReplaceAll(bundleID, `"`, `\"`)
	_, err := runOSAScript(ctx, `tell application id "`+quoted+`" to activate`)
	return err
}

func runOSAScript(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("osascript: %s", msg)
	}
	return string(out), nil
}
