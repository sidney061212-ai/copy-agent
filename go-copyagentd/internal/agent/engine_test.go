package agent

import (
	"context"
	"testing"
)

type collectingHandler struct {
	messages []Message
}

func (c *collectingHandler) HandleMessage(ctx context.Context, transport Transport, msg *Message) error {
	c.messages = append(c.messages, *msg)
	return nil
}

func TestEngineStartsTransportsAndHandlesMessages(t *testing.T) {
	transport := &emittingTransport{name: "stub", message: Message{Platform: "stub", UserID: "u1", Content: "copy hi"}}
	handler := &collectingHandler{}
	engine := NewEngine("default", []Transport{transport}, handler.HandleMessage)

	if err := engine.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if len(handler.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(handler.messages))
	}
	if handler.messages[0].Content != "copy hi" {
		t.Fatalf("unexpected message: %#v", handler.messages[0])
	}
}

type emittingTransport struct {
	name    string
	message Message
	started bool
	stopped bool
}

func (e *emittingTransport) Name() string { return e.name }
func (e *emittingTransport) Start(handler MessageHandler) error {
	e.started = true
	handler(e, &e.message)
	return nil
}
func (e *emittingTransport) Stop() error {
	e.stopped = true
	return nil
}
