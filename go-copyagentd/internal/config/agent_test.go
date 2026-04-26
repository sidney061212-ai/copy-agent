package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgentConfigDefaultsDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"feishuAppId":"app","feishuAppSecret":"secret"}`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Agent.Enabled {
		t.Fatal("agent should default disabled")
	}
	if cfg.Agent.Type != "codex" || cfg.Agent.Command != "codex" || cfg.Agent.SessionMode != "persistent" || cfg.Agent.IdleTimeoutMins != 15 {
		t.Fatalf("unexpected agent defaults: %#v", cfg.Agent)
	}
}

func TestLoadAgentConfigPreservesExplicitValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"agent":{"enabled":true,"type":"claude","command":"/bin/claude","sessionMode":"exec","idleTimeoutMins":3,"workDir":"/tmp","args":["--model","sonnet"],"systemPrompt":"bridge"}}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if !cfg.Agent.Enabled || cfg.Agent.Type != "claude" || cfg.Agent.Command != "/bin/claude" || cfg.Agent.WorkDir != "/tmp" || len(cfg.Agent.Args) != 2 || cfg.Agent.SystemPrompt != "bridge" {
		t.Fatalf("unexpected agent config: %#v", cfg.Agent)
	}
}
