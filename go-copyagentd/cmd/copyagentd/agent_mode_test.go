package main

import (
	"strings"
	"testing"

	"github.com/copyagent/copyagentd/internal/agent"
	"github.com/copyagent/copyagentd/internal/config"
)

func TestBuildFeishuEngineDefaultsToDirectMode(t *testing.T) {
	transport := stubMainTransport{name: "feishu"}
	engine, err := buildFeishuEngine(config.Config{}, transport, agent.DirectHandlerConfig{})
	if err != nil {
		t.Fatalf("buildFeishuEngine returned error: %v", err)
	}
	if engine.Name() != "feishu-direct" {
		t.Fatalf("engine name = %q", engine.Name())
	}
}

func TestBuildFeishuEngineStaysDirectWhenAgentUnavailable(t *testing.T) {
	transport := stubMainTransport{name: "feishu"}
	engine, err := buildFeishuEngine(config.Config{Agent: config.AgentConfig{Enabled: true, Type: "missing-agent", Command: "missing-agent"}}, transport, agent.DirectHandlerConfig{})
	if err != nil {
		t.Fatalf("buildFeishuEngine returned error: %v", err)
	}
	if engine.Name() != "feishu-direct" {
		t.Fatalf("engine name = %q", engine.Name())
	}
}

func TestBuildFeishuEngineCanStartDirectWithoutAgentCommand(t *testing.T) {
	transport := stubMainTransport{name: "feishu"}
	engine, err := buildFeishuEngine(config.Config{Agent: config.AgentConfig{Enabled: false, Type: "missing-agent", Command: "missing-agent"}}, transport, agent.DirectHandlerConfig{})
	if err != nil {
		t.Fatalf("buildFeishuEngine returned error: %v", err)
	}
	if engine.Name() != "feishu-direct" {
		t.Fatalf("engine name = %q", engine.Name())
	}
}

type stubMainTransport struct{ name string }

func (transport stubMainTransport) Name() string                     { return transport.name }
func (transport stubMainTransport) Start(agent.MessageHandler) error { return nil }
func (transport stubMainTransport) Stop() error                      { return nil }

func TestParseActionReplyTextPreservesFormatting(t *testing.T) {
	want := "第一行\n\n- 第二行\n  缩进\n"
	_, got, err := parseActionReplyText([]string{"--session-key", "feishu:chat:user", "--text", want})
	if err != nil {
		t.Fatalf("parseActionReplyText returned error: %v", err)
	}
	if got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}

func TestHelpTextUsesTurnAction(t *testing.T) {
	got := helpText()
	if !strings.Contains(got, "copyagentd action turn status|codex|claude|terminal|iterm|warp|vscode|cursor") {
		t.Fatalf("help text missing turn action: %q", got)
	}
	if strings.Contains(got, "copyagentd action target ") {
		t.Fatalf("help text should not expose target action: %q", got)
	}
}

func TestNormalizeActionCommandMapsLegacyTargetToTurn(t *testing.T) {
	if got := normalizeActionCommand("turn"); got != "turn" {
		t.Fatalf("normalizeActionCommand(turn) = %q", got)
	}
	if got := normalizeActionCommand("target"); got != "turn" {
		t.Fatalf("normalizeActionCommand(target) = %q", got)
	}
}
