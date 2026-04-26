package service

import (
	"errors"
	"os"
	"strings"
)

func TailFile(path string, lines int) (string, error) {
	if lines <= 0 {
		lines = 50
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "log file not found: " + path, nil
	}
	if err != nil {
		return "", err
	}
	parts := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(parts) > lines {
		parts = parts[len(parts)-lines:]
	}
	return strings.Join(parts, "\n"), nil
}

func LogFiles() ([]string, error) {
	if meta, err := LoadMeta(); err == nil && strings.TrimSpace(meta.LogFile) != "" {
		paths := []string{meta.LogFile}
		if _, err := os.Stat(meta.LogFile + ".1"); err == nil {
			paths = append(paths, meta.LogFile+".1")
		}
		return paths, nil
	}
	cfg, err := DefaultConfig("")
	if err != nil {
		return nil, err
	}
	paths := []string{cfg.LogFile}
	if _, err := os.Stat(cfg.LogFile + ".1"); err == nil {
		paths = append(paths, cfg.LogFile+".1")
	}
	return paths, nil
}
