package clipboard

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
)

func WriteText(text string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Env = utf8Env()
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		_, writeErr := stdin.Write([]byte(text))
		closeErr := stdin.Close()
		waitErr := cmd.Wait()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		return waitErr
	case "windows":
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		_, writeErr := stdin.Write([]byte(text))
		closeErr := stdin.Close()
		waitErr := cmd.Wait()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		return waitErr
	default:
		return errors.New("clipboard is not implemented for this platform")
	}
}

func ReadText() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("pbpaste")
		cmd.Env = utf8Env()
		out, err := cmd.Output()
		return string(out), err
	case "windows":
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-Clipboard -Raw").Output()
		return string(out), err
	default:
		return "", errors.New("clipboard read is not implemented for this platform")
	}
}

func utf8Env() []string {
	env := os.Environ()
	env = append(env, "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8")
	return env
}
