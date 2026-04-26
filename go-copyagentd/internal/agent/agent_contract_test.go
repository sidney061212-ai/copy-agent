package agent

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubCodingAgent struct{ name string }

func (s stubCodingAgent) Name() string { return s.name }
func (s stubCodingAgent) StartSession(context.Context, string) (AgentSession, error) {
	return nil, nil
}
func (s stubCodingAgent) Stop() error { return nil }

func TestRegistryCreatesRegisteredAgent(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterAgent("stub", func(opts map[string]any) (CodingAgent, error) {
		if opts["command"] != "stub-agent" {
			t.Fatalf("opts not passed through: %#v", opts)
		}
		return stubCodingAgent{name: "stub"}, nil
	})

	codingAgent, err := registry.CreateAgent("stub", map[string]any{"command": "stub-agent"})
	if err != nil {
		t.Fatalf("CreateAgent returned error: %v", err)
	}
	if codingAgent.Name() != "stub" {
		t.Fatalf("agent.Name() = %q", codingAgent.Name())
	}
}

func TestRegistryListsRegisteredAgentsSorted(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterAgent("codex", func(map[string]any) (CodingAgent, error) { return stubCodingAgent{name: "codex"}, nil })
	registry.RegisterAgent("claude", func(map[string]any) (CodingAgent, error) { return stubCodingAgent{name: "claude"}, nil })

	names := registry.ListRegisteredAgents()
	if len(names) != 2 || names[0] != "claude" || names[1] != "codex" {
		t.Fatalf("unexpected agent names: %#v", names)
	}
}

func TestRegistryReturnsAvailableNamesForUnknownAgent(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterAgent("codex", func(map[string]any) (CodingAgent, error) { return stubCodingAgent{name: "codex"}, nil })

	_, err := registry.CreateAgent("claude", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got == "" || !containsAll(got, "unknown agent", "codex") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistryPropagatesAgentFactoryError(t *testing.T) {
	expected := errors.New("boom")
	registry := NewRegistry()
	registry.RegisterAgent("bad", func(map[string]any) (CodingAgent, error) { return nil, expected })

	_, err := registry.CreateAgent("bad", nil)
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}
