package claude

import "github.com/copyagent/copyagentd/internal/agent"

func init() {
	agent.RegisterAgent("claude", NewFromOptions)
}
