package feishu

import (
	"fmt"

	"github.com/copyagent/copyagentd/internal/agent"
)

func init() {
	agent.RegisterTransport("feishu", NewRegisteredTransport)
}

func NewRegisteredTransport(opts map[string]any) (agent.Transport, error) {
	appID, _ := opts["appId"].(string)
	if appID == "" {
		appID, _ = opts["feishuAppId"].(string)
	}
	appSecret, _ := opts["appSecret"].(string)
	if appSecret == "" {
		appSecret, _ = opts["feishuAppSecret"].(string)
	}
	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("feishu transport requires appId and appSecret")
	}
	return NewAgentTransport(appID, appSecret), nil
}
