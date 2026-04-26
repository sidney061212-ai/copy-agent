package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTailFileHandlesMissingLog(t *testing.T) {
	text, err := TailFile(filepath.Join(t.TempDir(), "missing.log"), 10)
	if err != nil {
		t.Fatalf("tail missing: %v", err)
	}
	if !strings.Contains(text, "log file not found") {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestTailFileReturnsRecentLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "copyagentd.log")
	if err := os.WriteFile(path, []byte("one\ntwo\nthree\n"), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}
	text, err := TailFile(path, 2)
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	if strings.Contains(text, "one") || !strings.Contains(text, "two") || !strings.Contains(text, "three") {
		t.Fatalf("unexpected tail: %q", text)
	}
}

func TestLogFilesUsesMetaLogPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	logPath := filepath.Join(home, ".copyagent", "logs", "copyagentd.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(logPath, []byte("hello\n"), 0o600); err != nil {
		t.Fatalf("write current log: %v", err)
	}
	if err := os.WriteFile(logPath+".1", []byte("older\n"), 0o600); err != nil {
		t.Fatalf("write backup log: %v", err)
	}
	if err := SaveMeta(&Meta{LogFile: logPath}); err != nil {
		t.Fatalf("save meta: %v", err)
	}

	paths, err := LogFiles()
	if err != nil {
		t.Fatalf("log files: %v", err)
	}
	if len(paths) != 2 || paths[0] != logPath || paths[1] != logPath+".1" {
		t.Fatalf("unexpected log paths: %#v", paths)
	}
}
