package feishu

import (
	"testing"

	"github.com/copyagent/copyagentd/internal/agent"
)

func TestFeishuTransportFactoryValidatesCredentials(t *testing.T) {
	_, err := agent.CreateTransport("feishu", map[string]any{})
	if err == nil {
		t.Fatal("expected missing credentials error")
	}
}

func TestFeishuTransportFactoryCreatesRegisteredTransport(t *testing.T) {
	transport, err := agent.CreateTransport("feishu", map[string]any{"appId": "cli_xxx", "appSecret": "secret"})
	if err != nil {
		t.Fatalf("CreateTransport returned error: %v", err)
	}
	if transport.Name() != "feishu" {
		t.Fatalf("Name() = %q", transport.Name())
	}
}
