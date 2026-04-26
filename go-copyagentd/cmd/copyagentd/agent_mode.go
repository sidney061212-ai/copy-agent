package main

import (
	"log"

	"github.com/copyagent/copyagentd/internal/agent"
	"github.com/copyagent/copyagentd/internal/config"
	"github.com/copyagent/copyagentd/internal/inject"
)

const copyagentBridgePrompt = `你在 copyagent Agent Mode 中运行。保持回复简洁，优先完成用户明确要求。需要写入剪切板或保存文件时，后续版本会提供 copyagent action CLI；当前阶段请只说明你会如何处理，不要输出真实 token 或密钥。`

func buildFeishuEngine(cfg config.Config, transport agent.Transport, directCfg agent.DirectHandlerConfig) (*agent.Engine, error) {
	directHandler := agent.NewDirectHandler(directCfg)
	agentHandler, err := buildAgentModeHandler(cfg, directCfg)
	agentBuildErr := err
	initialEnabled := cfg.Agent.Enabled && err == nil
	if cfg.Agent.Enabled && err != nil {
		log.Printf("agent mode unavailable, staying in Direct Mode: %v", err)
	}
	switcher := agent.NewModeSwitcher(agent.ModeSwitcherConfig{
		Direct:         directHandler,
		Agent:          agentHandler,
		InitialEnabled: initialEnabled,
		OnChange: func(enabled bool) error {
			if enabled && agentHandler == nil {
				return agentBuildErr
			}
			cfg.Agent.Enabled = enabled
			return config.SetAgentEnabled("", enabled)
		},
	})
	mode := "feishu-direct"
	if initialEnabled {
		mode = "feishu-agent"
	}
	return agent.NewEngine(mode, []agent.Transport{transport}, switcher.HandleMessage), nil
}

func buildAgentModeHandler(cfg config.Config, directCfg agent.DirectHandlerConfig) (*agent.AgentModeHandler, error) {
	agentOptions := map[string]any{
		"command":      cfg.Agent.Command,
		"workDir":      cfg.Agent.WorkDir,
		"args":         cfg.Agent.Args,
		"systemPrompt": cfg.Agent.SystemPrompt,
	}
	codingAgent, err := agent.CreateAgent(cfg.Agent.Type, agentOptions)
	if err != nil {
		return nil, err
	}
	systemPrompt := cfg.Agent.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = copyagentBridgePrompt
	}
	return agent.NewAgentModeHandler(agent.AgentModeHandlerConfig{
		Policy:       agent.DirectPolicyConfig{AllowedUserIDs: cfg.AllowedActorIDs},
		Direct:       directCfg,
		Agent:        codingAgent,
		SystemPrompt: systemPrompt,
		Injector:     inject.NewDefaultService(),
	}), nil
}
