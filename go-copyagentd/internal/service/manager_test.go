package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSetsDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := Config{}
	if err := Resolve(&cfg); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if cfg.BinaryPath == "" || cfg.ConfigPath == "" || cfg.WorkDir == "" || cfg.LogFile == "" || cfg.LogMaxSize <= 0 {
		t.Fatalf("resolve did not fill defaults: %#v", cfg)
	}
}

func TestResolveBuildsPortableFallbackPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")

	cfg := Config{}
	if err := Resolve(&cfg); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if cfg.EnvPATH == "" {
		t.Fatal("expected fallback path to be populated")
	}
	if got, want := cfg.EnvPATH, filepath.Join(home, ".local", "bin"); got == want {
		return
	}
	if want := filepath.Join(home, ".local", "bin"); !containsPathSegment(cfg.EnvPATH, want) {
		t.Fatalf("fallback path %q does not include %q", cfg.EnvPATH, want)
	}
	if strings.Contains(cfg.EnvPATH, "/Users/") {
		t.Fatalf("fallback path should be derived from HOME during the test, got %q", cfg.EnvPATH)
	}
}

func containsPathSegment(pathValue, segment string) bool {
	for _, part := range strings.Split(pathValue, ":") {
		if part == segment {
			return true
		}
	}
	return false
}

func TestSaveAndLoadMeta(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	want := &Meta{
		LogFile:     filepath.Join(home, ".copyagent", "logs", "copyagentd.log"),
		LogMaxSize:  123,
		WorkDir:     "/tmp/work",
		BinaryPath:  "/tmp/copyagentd",
		ConfigPath:  "/tmp/config.json",
		InstalledAt: "2026-04-26T00:00:00Z",
	}
	if err := SaveMeta(want); err != nil {
		t.Fatalf("save meta: %v", err)
	}
	got, err := LoadMeta()
	if err != nil {
		t.Fatalf("load meta: %v", err)
	}
	if *got != *want {
		t.Fatalf("meta mismatch: got=%#v want=%#v", got, want)
	}

	RemoveMeta()
	if _, err := os.Stat(filepath.Join(home, ".copyagent", "daemon.json")); !os.IsNotExist(err) {
		t.Fatalf("expected meta to be removed, err=%v", err)
	}
}
