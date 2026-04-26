package codex

import "github.com/copyagent/copyagentd/internal/agent"

func init() {
	agent.RegisterAgent("codex", NewFromOptions)
}
