package agent

import "testing"

func TestMemorySessionStoreStoresAndClearsAgentSessionID(t *testing.T) {
	store := NewMemorySessionStore()
	if _, ok := store.AgentSessionID("feishu:c1:u1"); ok {
		t.Fatal("expected empty store")
	}

	store.SetAgentSessionID("feishu:c1:u1", "codex-thread-1")
	got, ok := store.AgentSessionID("feishu:c1:u1")
	if !ok || got != "codex-thread-1" {
		t.Fatalf("AgentSessionID() = %q, %v", got, ok)
	}

	store.ClearAgentSessionID("feishu:c1:u1")
	if _, ok := store.AgentSessionID("feishu:c1:u1"); ok {
		t.Fatal("expected cleared session id")
	}
}

func TestMemorySessionStoreKeepsSessionKeysIsolated(t *testing.T) {
	store := NewMemorySessionStore()
	store.SetAgentSessionID("feishu:c1:u1", "thread-1")
	store.SetAgentSessionID("feishu:c1:u2", "thread-2")

	got, _ := store.AgentSessionID("feishu:c1:u2")
	if got != "thread-2" {
		t.Fatalf("wrong session id for second key: %q", got)
	}
}
