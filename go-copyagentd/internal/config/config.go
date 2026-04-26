package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Agent              AgentConfig `json:"agent"`
	Host               string      `json:"host"`
	Port               int         `json:"port"`
	Token              string      `json:"token"`
	Mode               string      `json:"mode"`
	FeishuAppID        string      `json:"feishuAppId"`
	FeishuAppSecret    string      `json:"feishuAppSecret"`
	AllowedActorIDs    []string    `json:"allowedActorIds"`
	DefaultDownloadDir string      `json:"defaultDownloadDir"`
	ImageAction        string      `json:"imageAction"`
	ReplyEnabled       bool        `json:"replyEnabled"`
}

func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".copyagent", "config.json")
}

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	if path == "" {
		return Config{}, errors.New("cannot resolve home directory")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}
	if cfg.Port == 0 {
		cfg.Port = 8765
	}
	if cfg.DefaultDownloadDir == "" {
		cfg.DefaultDownloadDir = "~/Downloads/copyagent"
	}
	if cfg.Agent.Type == "" {
		cfg.Agent.Type = "codex"
	}
	if cfg.Agent.Command == "" {
		cfg.Agent.Command = cfg.Agent.Type
	}
	if cfg.Agent.SessionMode == "" {
		cfg.Agent.SessionMode = "persistent"
	}
	if cfg.Agent.IdleTimeoutMins == 0 {
		cfg.Agent.IdleTimeoutMins = 15
	}
	return cfg, nil
}
