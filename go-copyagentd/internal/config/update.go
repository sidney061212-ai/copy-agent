package config

import (
	"encoding/json"
	"os"
)

func SetAgentEnabled(path string, enabled bool) error {
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	agentRaw, _ := raw["agent"].(map[string]any)
	if agentRaw == nil {
		agentRaw = make(map[string]any)
		raw["agent"] = agentRaw
	}
	agentRaw["enabled"] = enabled
	updated, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	updated = append(updated, '\n')
	return os.WriteFile(path, updated, 0o600)
}
