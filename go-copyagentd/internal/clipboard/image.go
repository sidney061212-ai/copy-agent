package clipboard

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

func WritePNGFile(ctx context.Context, path string) error {
	if path == "" {
		return errors.New("image path is required")
	}
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf("set the clipboard to (read (POSIX file %q) as «class PNGf»)", path)
		cmd := exec.CommandContext(ctx, "osascript", "-e", script)
		cmd.Env = utf8Env()
		out, err := cmd.CombinedOutput()
		if err != nil {
			if len(out) > 0 {
				return fmt.Errorf("copy image to clipboard: %s", string(out))
			}
			return err
		}
		return nil
	default:
		return errors.New("image clipboard is not implemented for this platform")
	}
}
