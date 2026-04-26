package agent

import "sync"

type AgentSessionStore interface {
	AgentSessionID(sessionKey string) (string, bool)
	SetAgentSessionID(sessionKey string, agentSessionID string)
	ClearAgentSessionID(sessionKey string)
}

type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]string
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{sessions: make(map[string]string)}
}

func (store *MemorySessionStore) AgentSessionID(sessionKey string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	agentSessionID, ok := store.sessions[sessionKey]
	return agentSessionID, ok
}

func (store *MemorySessionStore) SetAgentSessionID(sessionKey string, agentSessionID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.sessions[sessionKey] = agentSessionID
}

func (store *MemorySessionStore) ClearAgentSessionID(sessionKey string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, sessionKey)
}
