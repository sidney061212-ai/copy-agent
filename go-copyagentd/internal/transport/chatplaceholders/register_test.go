package chatplaceholders

import (
	"testing"

	"github.com/copyagent/copyagentd/internal/agent"
)

func TestChatPlaceholderTransportsAreRegistered(t *testing.T) {
	registered := map[string]bool{}
	for _, name := range agent.ListRegisteredTransports() {
		registered[name] = true
	}
	for _, name := range Names {
		if !registered[name] {
			t.Fatalf("transport %q is not registered", name)
		}
	}
}

func TestChatPlaceholderTransportIsDisabledNoop(t *testing.T) {
	transport, err := agent.CreateTransport("slack", nil)
	if err != nil {
		t.Fatalf("CreateTransport returned error: %v", err)
	}
	if transport.Name() != "slack" {
		t.Fatalf("Name() = %q", transport.Name())
	}
	if err := transport.Start(func(agent.Transport, *agent.Message) { t.Fatal("disabled transport should not emit messages") }); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}
