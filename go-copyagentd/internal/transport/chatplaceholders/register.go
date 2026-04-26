package chatplaceholders

import "github.com/copyagent/copyagentd/internal/agent"

var Names = []string{
	"dingtalk",
	"slack",
	"telegram",
	"discord",
	"line",
	"wechat-work",
	"weixin",
	"qq",
	"qqbot",
	"weibo",
}

func init() {
	for _, name := range Names {
		transportName := name
		agent.RegisterTransport(transportName, func(map[string]any) (agent.Transport, error) {
			return agent.NewDisabledTransport(transportName), nil
		})
	}
}
