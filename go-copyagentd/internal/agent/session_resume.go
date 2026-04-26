package agent

import "context"

func StartOrResumeSession(ctx context.Context, codingAgent CodingAgent, store AgentSessionStore, sessionKey string) (AgentSession, error) {
	if agentSessionID, ok := store.AgentSessionID(sessionKey); ok && agentSessionID != "" {
		session, err := codingAgent.StartSession(ctx, agentSessionID)
		if err == nil {
			return session, nil
		}
		store.ClearAgentSessionID(sessionKey)
	}
	session, err := codingAgent.StartSession(ctx, "")
	if err != nil {
		return nil, err
	}
	if currentSessionID := session.CurrentSessionID(); currentSessionID != "" {
		store.SetAgentSessionID(sessionKey, currentSessionID)
	}
	return session, nil
}
