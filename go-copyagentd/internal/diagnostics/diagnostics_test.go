package diagnostics

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/copyagent/copyagentd/internal/config"
)

func TestDoctorReportsConfigPermissionsAndDownloadDir(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	checks := Doctor(config.Config{DefaultDownloadDir: filepath.Join(dir, "downloads")}, configPath)

	if !hasCheck(checks, "config_permissions") {
		t.Fatal("missing config permissions check")
	}
	if !hasCheck(checks, "download_dir") {
		t.Fatal("missing download dir check")
	}
}

func hasCheck(checks []Check, name string) bool {
	for _, check := range checks {
		if check.Name == name {
			return true
		}
	}
	return false
}
