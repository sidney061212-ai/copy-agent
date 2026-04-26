package config

type AgentConfig struct {
	Enabled         bool     `json:"enabled"`
	Type            string   `json:"type"`
	Command         string   `json:"command"`
	SessionMode     string   `json:"sessionMode"`
	IdleTimeoutMins int      `json:"idleTimeoutMins"`
	WorkDir         string   `json:"workDir"`
	Args            []string `json:"args"`
	SystemPrompt    string   `json:"systemPrompt"`
}
