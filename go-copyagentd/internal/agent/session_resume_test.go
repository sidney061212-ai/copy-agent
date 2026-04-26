package agent

import (
	"context"
	"errors"
	"testing"
)

type startSessionCall struct {
	sessionID string
}

type resumeAgent struct {
	calls       []startSessionCall
	failSession map[string]error
	newSession  AgentSession
}

func (agent *resumeAgent) Name() string { return "resume" }
func (agent *resumeAgent) StartSession(_ context.Context, sessionID string) (AgentSession, error) {
	agent.calls = append(agent.calls, startSessionCall{sessionID: sessionID})
	if err := agent.failSession[sessionID]; err != nil {
		return nil, err
	}
	if sessionID != "" {
		return fixedAgentSession{id: sessionID}, nil
	}
	return agent.newSession, nil
}
func (agent *resumeAgent) Stop() error { return nil }

type fixedAgentSession struct{ id string }

func (session fixedAgentSession) Send(context.Context, string, AgentAttachments) error { return nil }
func (session fixedAgentSession) RespondPermission(context.Context, string, PermissionResult) error {
	return nil
}
func (session fixedAgentSession) Events() <-chan AgentEvent { return nil }
func (session fixedAgentSession) CurrentSessionID() string  { return session.id }
func (session fixedAgentSession) Alive() bool               { return true }
func (session fixedAgentSession) Close() error              { return nil }

func TestStartOrResumeSessionUsesStoredAgentSessionID(t *testing.T) {
	store := NewMemorySessionStore()
	store.SetAgentSessionID("feishu:c1:u1", "thread-1")
	agent := &resumeAgent{failSession: map[string]error{}, newSession: fixedAgentSession{id: "thread-new"}}

	session, err := StartOrResumeSession(context.Background(), agent, store, "feishu:c1:u1")
	if err != nil {
		t.Fatalf("StartOrResumeSession returned error: %v", err)
	}
	if session.CurrentSessionID() != "thread-1" {
		t.Fatalf("session id = %q", session.CurrentSessionID())
	}
	if len(agent.calls) != 1 || agent.calls[0].sessionID != "thread-1" {
		t.Fatalf("unexpected calls: %#v", agent.calls)
	}
}

func TestStartOrResumeSessionClearsStoredIDAndStartsFreshOnResumeFailure(t *testing.T) {
	store := NewMemorySessionStore()
	store.SetAgentSessionID("feishu:c1:u1", "stale-thread")
	agent := &resumeAgent{
		failSession: map[string]error{"stale-thread": errors.New("not found")},
		newSession:  fixedAgentSession{id: "fresh-thread"},
	}

	session, err := StartOrResumeSession(context.Background(), agent, store, "feishu:c1:u1")
	if err != nil {
		t.Fatalf("StartOrResumeSession returned error: %v", err)
	}
	if session.CurrentSessionID() != "fresh-thread" {
		t.Fatalf("session id = %q", session.CurrentSessionID())
	}
	if len(agent.calls) != 2 || agent.calls[0].sessionID != "stale-thread" || agent.calls[1].sessionID != "" {
		t.Fatalf("unexpected calls: %#v", agent.calls)
	}
	got, ok := store.AgentSessionID("feishu:c1:u1")
	if !ok || got != "fresh-thread" {
		t.Fatalf("stored session id = %q, %v", got, ok)
	}
}

func TestStartOrResumeSessionReturnsFreshStartError(t *testing.T) {
	store := NewMemorySessionStore()
	expected := errors.New("cannot start")
	agent := &resumeAgent{failSession: map[string]error{"": expected}}

	_, err := StartOrResumeSession(context.Background(), agent, store, "feishu:c1:u1")
	if !errors.Is(err, expected) {
		t.Fatalf("expected %v, got %v", expected, err)
	}
}
