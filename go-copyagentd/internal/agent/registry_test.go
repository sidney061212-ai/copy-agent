package agent

import (
	"errors"
	"strings"
	"testing"
)

type stubTransport struct{ name string }

func (s stubTransport) Name() string { return s.name }
func (s stubTransport) Start(handler MessageHandler) error { return nil }
func (s stubTransport) Stop() error { return nil }

func TestRegistryCreatesRegisteredTransport(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterTransport("stub", func(opts map[string]any) (Transport, error) {
		if opts["enabled"] != true {
			t.Fatalf("opts not passed through: %#v", opts)
		}
		return stubTransport{name: "stub"}, nil
	})

	transport, err := registry.CreateTransport("stub", map[string]any{"enabled": true})
	if err != nil {
		t.Fatalf("CreateTransport returned error: %v", err)
	}
	if transport.Name() != "stub" {
		t.Fatalf("transport.Name() = %q", transport.Name())
	}
}

func TestRegistryReturnsAvailableNamesForUnknownTransport(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterTransport("feishu", func(map[string]any) (Transport, error) { return stubTransport{name: "feishu"}, nil })

	_, err := registry.CreateTransport("slack", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown transport") || !strings.Contains(err.Error(), "feishu") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegistryPropagatesFactoryError(t *testing.T) {
	expected := errors.New("boom")
	registry := NewRegistry()
	registry.RegisterTransport("bad", func(map[string]any) (Transport, error) { return nil, expected })

	_, err := registry.CreateTransport("bad", nil)
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
