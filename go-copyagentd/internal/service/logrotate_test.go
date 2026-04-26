package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRotatingWriterCreatesBackup(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "copyagentd.log")
	writer, err := NewRotatingWriter(logPath, 24)
	if err != nil {
		t.Fatalf("new rotating writer: %v", err)
	}
	defer writer.Close()

	for _, line := range []string{"1234567890\n", "abcdefghij\n", "rotate-now\n", "after-rotate\n"} {
		if _, err := writer.Write([]byte(line)); err != nil {
			t.Fatalf("write log: %v", err)
		}
	}

	current, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	backup, err := os.ReadFile(logPath + ".1")
	if err != nil {
		t.Fatalf("read backup log: %v", err)
	}
	if len(current) == 0 || len(backup) == 0 {
		t.Fatalf("expected both current and backup logs, got current=%q backup=%q", string(current), string(backup))
	}
	if !strings.Contains(string(current), "after-rotate") {
		t.Fatalf("expected post-rotation line in current log, got current=%q backup=%q", string(current), string(backup))
	}
	if !strings.Contains(string(backup), "rotate-now") {
		t.Fatalf("expected rotated line in backup log, got current=%q backup=%q", string(current), string(backup))
	}
}
