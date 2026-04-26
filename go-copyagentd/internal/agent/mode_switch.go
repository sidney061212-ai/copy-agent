package agent

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/copyagent/copyagentd/internal/inject"
)

const AgentModeEnabledReply = "✅ 已进入 Agent 模式"
const DirectModeEnabledReply = "✅ 已进入无 AI 复制模式"

type ModeSwitcher struct {
	mu       sync.RWMutex
	direct   *DirectHandler
	agent    *AgentModeHandler
	enabled  bool
	onChange func(enabled bool) error
}

type ModeSwitcherConfig struct {
	Direct         *DirectHandler
	Agent          *AgentModeHandler
	InitialEnabled bool
	OnChange       func(enabled bool) error
}

func NewModeSwitcher(cfg ModeSwitcherConfig) *ModeSwitcher {
	return &ModeSwitcher{direct: cfg.Direct, agent: cfg.Agent, enabled: cfg.InitialEnabled, onChange: cfg.OnChange}
}

func (switcher *ModeSwitcher) HandleMessage(ctx context.Context, transport Transport, msg *Message) error {
	if msg == nil {
		return nil
	}
	switch normalizedModeCommand(msg.Content) {
	case "/agent":
		return switcher.setMode(ctx, transport, msg, true)
	case "/copy":
		return switcher.setMode(ctx, transport, msg, false)
	}
	switcher.mu.RLock()
	enabled := switcher.enabled
	direct := switcher.direct
	agentHandler := switcher.agent
	switcher.mu.RUnlock()
	if isForegroundHostingCommand(msg.Content) {
		if agentHandler != nil {
			return agentHandler.HandleMessage(ctx, transport, msg)
		}
		return replyToTransport(ctx, transport, msg.ReplyCtx, "前台远程托管命令当前不可用：命令处理器尚未配置。")
	}
	if enabled && agentHandler != nil {
		return agentHandler.HandleMessage(ctx, transport, msg)
	}
	if direct == nil {
		return nil
	}
	return direct.HandleMessage(ctx, transport, msg)
}

func (switcher *ModeSwitcher) Enabled() bool {
	switcher.mu.RLock()
	defer switcher.mu.RUnlock()
	return switcher.enabled
}

func (switcher *ModeSwitcher) setMode(ctx context.Context, transport Transport, msg *Message, enabled bool) error {
	switcher.mu.Lock()
	if switcher.onChange != nil {
		if err := switcher.onChange(enabled); err != nil {
			switcher.mu.Unlock()
			return err
		}
	}
	switcher.enabled = enabled
	switcher.mu.Unlock()
	reply := DirectModeEnabledReply
	if enabled {
		reply = AgentModeEnabledReply
	}
	log.Printf("copyagent runtime mode switched: enabled=%v session=%s message=%s", enabled, msg.EffectiveSessionKey(), msg.MessageID)
	if err := replyToTransport(ctx, transport, msg.ReplyCtx, reply); err != nil && !errors.Is(err, ErrAgentModeReplyUnsupported) {
		return err
	}
	return nil
}

func normalizedModeCommand(content string) string {
	trimmed := strings.TrimSpace(content)
	fields := strings.Fields(trimmed)
	if len(fields) != 1 {
		return ""
	}
	switch strings.ToLower(fields[0]) {
	case "/agent", "／agent":
		return "/agent"
	case "/copy", "／copy":
		return "/copy"
	default:
		return ""
	}
}

func isForegroundHostingCommand(content string) bool {
	if _, ok := parseTurnCommand(content); ok {
		return true
	}
	if _, ok := parseLegacyTargetCommand(content); ok {
		return true
	}
	if _, ok := inject.ParseCommand(content); ok {
		return true
	}
	return false
}

func replyToTransport(ctx context.Context, transport Transport, replyCtx any, content string) error {
	if replyCtx == nil || strings.TrimSpace(content) == "" {
		return nil
	}
	replier, ok := transport.(ReplyCapable)
	if !ok {
		return ErrAgentModeReplyUnsupported
	}
	return replier.Reply(ctx, replyCtx, content)
}
