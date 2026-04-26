package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetAgentEnabledPreservesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"feishuAppSecret":"secret","agent":{"type":"codex","enabled":false}}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := SetAgentEnabled(path, true); err != nil {
		t.Fatalf("SetAgentEnabled returned error: %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(mustRead(t, path), &raw); err != nil {
		t.Fatalf("json: %v", err)
	}
	agent := raw["agent"].(map[string]any)
	if agent["enabled"] != true || agent["type"] != "codex" || raw["feishuAppSecret"] != "secret" {
		t.Fatalf("unexpected config: %#v", raw)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return data
}
