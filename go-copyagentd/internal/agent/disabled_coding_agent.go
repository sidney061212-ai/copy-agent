package agent

import (
	"context"
	"fmt"
)

type DisabledCodingAgent struct {
	name string
}

func NewDisabledCodingAgent(name string) *DisabledCodingAgent {
	return &DisabledCodingAgent{name: name}
}

func (agent *DisabledCodingAgent) Name() string { return agent.name }

func (agent *DisabledCodingAgent) StartSession(context.Context, string) (AgentSession, error) {
	return nil, fmt.Errorf("agent %q is registered but not implemented", agent.name)
}

func (agent *DisabledCodingAgent) Stop() error { return nil }
