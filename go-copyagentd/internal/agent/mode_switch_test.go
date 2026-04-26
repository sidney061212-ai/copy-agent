package agent

import (
	"context"
	"testing"
	"time"

	"github.com/copyagent/copyagentd/internal/inject"
)

func TestModeSwitcherAgentAndCopyCommands(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	var changes []bool
	switcher := NewModeSwitcher(ModeSwitcherConfig{
		Direct: NewDirectHandler(DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}}),
		Agent: NewAgentModeHandler(AgentModeHandlerConfig{
			Direct: DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
			Agent:  &scriptedAgent{},
		}),
		OnChange: func(enabled bool) error {
			changes = append(changes, enabled)
			return nil
		},
	})
	if err := switcher.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", MessageID: "m1", Content: "/agent", ReplyCtx: "m1"}); err != nil {
		t.Fatalf("/agent returned error: %v", err)
	}
	if !switcher.Enabled() || len(changes) != 1 || !changes[0] || transport.replies[0] != AgentModeEnabledReply {
		t.Fatalf("after /agent enabled=%v changes=%#v replies=%#v", switcher.Enabled(), changes, transport.replies)
	}
	if err := switcher.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", MessageID: "m2", Content: "/copy", ReplyCtx: "m2"}); err != nil {
		t.Fatalf("/copy returned error: %v", err)
	}
	if switcher.Enabled() || len(changes) != 2 || changes[1] || transport.replies[1] != DirectModeEnabledReply {
		t.Fatalf("after /copy enabled=%v changes=%#v replies=%#v", switcher.Enabled(), changes, transport.replies)
	}
}

func TestModeSwitcherRoutesByCurrentMode(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	clipboard := &mockTextClipboard{}
	codingAgent := &scriptedAgent{}
	switcher := NewModeSwitcher(ModeSwitcherConfig{
		Direct: NewDirectHandler(DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: clipboard}}),
		Agent: NewAgentModeHandler(AgentModeHandlerConfig{
			Direct: DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: clipboard}},
			Agent:  codingAgent,
		}),
	})
	if err := switcher.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", MessageID: "m1", Content: "copy direct"}); err != nil {
		t.Fatalf("direct copy returned error: %v", err)
	}
	if len(clipboard.texts) != 1 || clipboard.texts[0] != "direct" {
		t.Fatalf("clipboard = %#v", clipboard.texts)
	}
	if err := switcher.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", MessageID: "m2", Content: "/agent", ReplyCtx: "m2"}); err != nil {
		t.Fatalf("/agent returned error: %v", err)
	}
	if err := switcher.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "m3", Content: "hello", ReplyCtx: "m3"}); err != nil {
		t.Fatalf("agent message returned error: %v", err)
	}
	waitForReplies(t, transport, 2)
	if len(codingAgent.sessions) != 1 {
		t.Fatalf("agent sessions = %#v", codingAgent.sessions)
	}
}

func TestModeSwitcherRoutesTurnCommandEvenWhenDirectMode(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	clipboard := &mockTextClipboard{}
	injector := &mockInjector{result: inject.Result{Target: inject.Target{AppName: "Codex", BundleID: "com.openai.codex"}}}
	switcher := NewModeSwitcher(ModeSwitcherConfig{
		Direct: NewDirectHandler(DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: clipboard}}),
		Agent: NewAgentModeHandler(AgentModeHandlerConfig{
			Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: clipboard}},
			Agent:    &scriptedAgent{},
			Injector: injector,
		}),
		InitialEnabled: false,
	})
	if err := switcher.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "m1", Content: "/turn codex", ReplyCtx: "m1"}); err != nil {
		t.Fatalf("/turn returned error: %v", err)
	}
	if len(clipboard.texts) != 0 {
		t.Fatalf("/turn should not fall back to direct copy: %#v", clipboard.texts)
	}
	if len(injector.activated) != 1 || injector.activated[0] != "com.openai.codex" {
		t.Fatalf("activated = %#v", injector.activated)
	}
	if len(transport.replies) != 1 || transport.replies[0] != "已切换应用：codex -> Codex (com.openai.codex)" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestNormalizedModeCommand(t *testing.T) {
	for _, input := range []string{"/agent", " ／agent ", "/copy", "／copy"} {
		if normalizedModeCommand(input) == "" {
			t.Fatalf("expected command for %q", input)
		}
	}
	if normalizedModeCommand("/agent now") != "" {
		t.Fatal("command with args should not match")
	}
}

func waitForReplies(t *testing.T, transport *mockReplyTransport, count int) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for len(transport.replies) < count {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for replies: %#v", transport.replies)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
