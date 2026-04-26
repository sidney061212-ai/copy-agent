package agent

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/copyagent/copyagentd/internal/inject"
)

var ErrAgentModeReplyUnsupported = errors.New("agent mode reply requires ReplyCapable transport")

type AgentModeHandler struct {
	policy   *DirectPolicy
	direct   *DirectHandler
	agent    CodingAgent
	store    AgentSessionStore
	queue    *SessionTurnQueue
	prompt   string
	injector inject.Executor
	sessions map[string]AgentSession
	targets  map[string]inject.NamedTarget
	mu       sync.Mutex
}

type AgentModeHandlerConfig struct {
	Policy       DirectPolicyConfig
	Direct       DirectHandlerConfig
	Agent        CodingAgent
	Store        AgentSessionStore
	Queue        *SessionTurnQueue
	SystemPrompt string
	Injector     inject.Executor
}

func NewAgentModeHandler(cfg AgentModeHandlerConfig) *AgentModeHandler {
	store := cfg.Store
	if store == nil {
		store = NewMemorySessionStore()
	}
	queue := cfg.Queue
	if queue == nil {
		queue = NewSessionTurnQueue(8)
	}
	policyCfg := cfg.Policy
	if len(policyCfg.AllowedUserIDs) == 0 && len(cfg.Direct.Policy.AllowedUserIDs) > 0 {
		policyCfg = cfg.Direct.Policy
	}
	return &AgentModeHandler{
		policy:   NewDirectPolicy(policyCfg),
		direct:   NewDirectHandler(cfg.Direct),
		agent:    cfg.Agent,
		store:    store,
		queue:    queue,
		prompt:   cfg.SystemPrompt,
		injector: cfg.Injector,
		sessions: make(map[string]AgentSession),
		targets:  make(map[string]inject.NamedTarget),
	}
}

func (handler *AgentModeHandler) HandleMessage(ctx context.Context, transport Transport, msg *Message) error {
	if handler.shouldDirect(msg) {
		log.Printf("agent mode fast-path direct: transport=%s session=%s message=%s images=%d files=%d", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID, len(msg.Images), len(msg.Files))
		return handler.direct.HandleMessage(ctx, transport, msg)
	}
	if text, ok := inject.ParseCommand(messageContent(msg)); ok {
		return handler.handleInject(ctx, transport, msg, text)
	}
	if legacyArg, ok := parseLegacyTargetCommand(messageContent(msg)); ok {
		return handler.handleLegacyTarget(ctx, transport, msg, legacyArg)
	}
	if targetArg, ok := parseTurnCommand(messageContent(msg)); ok {
		return handler.handleTurn(ctx, transport, msg, targetArg)
	}
	allowed, err := handler.policy.Allow(msg)
	if err != nil || !allowed {
		if err != nil {
			log.Printf("agent mode policy error: transport=%s session=%s message=%s err=%v", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID, err)
		} else {
			log.Printf("agent mode skipped by policy: transport=%s session=%s message=%s", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID)
		}
		return err
	}
	started, err := handler.queue.BeginOrQueue(msg.EffectiveSessionKey(), msg)
	if err != nil || !started {
		if err != nil {
			handler.policy.Complete(msg, false)
			log.Printf("agent mode queue error: transport=%s session=%s message=%s err=%v", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID, err)
		} else {
			log.Printf("agent mode queued: transport=%s session=%s message=%s", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID)
		}
		return err
	}
	log.Printf("agent mode routed to agent: transport=%s session=%s message=%s bytes=%d", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID, len([]byte(msg.Content)))
	go handler.runAgentTurn(ctx, transport, cloneMessage(msg))
	return nil
}

func (handler *AgentModeHandler) handleInject(ctx context.Context, transport Transport, msg *Message, text string) error {
	allowed, err := handler.policy.Allow(msg)
	if err != nil || !allowed {
		if err != nil {
			log.Printf("agent mode inject policy error: transport=%s session=%s message=%s err=%v", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID, err)
		} else {
			log.Printf("agent mode inject skipped by policy: transport=%s session=%s message=%s", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID)
		}
		return err
	}
	success := false
	defer func() { handler.policy.Complete(msg, success) }()
	if handler.injector == nil {
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "未粘贴：本机注入器尚未配置。")
		return nil
	}
	injectText := handler.injectPayload(msg.EffectiveSessionKey(), text)
	targetBundleID := ""
	if target, ok := handler.boundTarget(msg.EffectiveSessionKey()); ok {
		targetBundleID = target.BundleID
	}
	result, err := handler.injector.InjectText(ctx, inject.Request{Text: injectText, Submit: true, TargetBundleID: targetBundleID, AllowedBundleIDs: inject.DefaultAllowedBundleIDs(), MaxBytes: inject.DefaultMaxBytes})
	if err != nil {
		_ = handler.reply(ctx, transport, msg.ReplyCtx, injectFailureReply(result.Target, err))
		return nil
	}
	reply := "已提交到前台应用：" + targetLabel(result.Target) + "，等待回程回复。"
	if result.Warning != "" {
		reply += "\n" + result.Warning
	}
	if err := handler.reply(ctx, transport, msg.ReplyCtx, reply); err != nil {
		return err
	}
	log.Printf("agent mode injected text: transport=%s session=%s message=%s bytes=%d app=%s bundle=%s restored=%t warning=%t", transportName(transport), msg.EffectiveSessionKey(), msg.MessageID, len([]byte(injectText)), result.Target.AppName, result.Target.BundleID, result.RestoredClipboard, result.Warning != "")
	success = true
	return nil
}

func (handler *AgentModeHandler) handleTurn(ctx context.Context, transport Transport, msg *Message, arg string) error {
	if handler.injector == nil {
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "turn unavailable: 本机注入器尚未配置。")
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(arg), "status") || strings.TrimSpace(arg) == "" {
		current, err := handler.injector.Status(ctx)
		if err != nil {
			_ = handler.reply(ctx, transport, msg.ReplyCtx, "turn status failed: "+err.Error())
			return nil
		}
		bound := "未绑定"
		if target, ok := handler.boundTarget(msg.EffectiveSessionKey()); ok {
			bound = target.Name + " (" + target.BundleID + ")"
		}
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "当前前台应用："+targetLabel(current)+" ("+current.BundleID+")\n当前绑定目标："+bound)
		return nil
	}
	target, ok := inject.ResolveNamedTarget(arg)
	if !ok {
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "未知目标："+strings.TrimSpace(arg)+"。可用目标："+availableTargetsText())
		return nil
	}
	current, err := handler.injector.Activate(ctx, target.BundleID)
	if err != nil {
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "应用切换失败："+target.Name+" ("+target.BundleID+")："+err.Error())
		return nil
	}
	handler.setBoundTarget(msg.EffectiveSessionKey(), target)
	_ = handler.reply(ctx, transport, msg.ReplyCtx, "已切换应用："+target.Name+" -> "+targetLabel(current)+" ("+current.BundleID+")")
	return nil
}

func parseTurnCommand(content string) (string, bool) {
	return parseSlashCommand(content, "turn")
}

func parseLegacyTargetCommand(content string) (string, bool) {
	return parseSlashCommand(content, "target")
}

func parseSlashCommand(content string, name string) (string, bool) {
	trimmed := strings.TrimSpace(content)
	for _, prefix := range []string{"/", "／"} {
		command := prefix + name
		if trimmed == command {
			return "status", true
		}
		for _, sep := range []string{" ", "\n", "\t"} {
			fullPrefix := command + sep
			if strings.HasPrefix(trimmed, fullPrefix) {
				return strings.TrimSpace(strings.TrimPrefix(trimmed, command)), true
			}
		}
	}
	return "", false
}

func (handler *AgentModeHandler) handleLegacyTarget(ctx context.Context, transport Transport, msg *Message, arg string) error {
	suggestion := "/turn"
	if clean := strings.TrimSpace(arg); clean != "" && !strings.EqualFold(clean, "status") {
		suggestion += " " + clean
	}
	_ = handler.reply(ctx, transport, msg.ReplyCtx, "命令已更新：请改用 "+suggestion+"。")
	return nil
}

func availableTargetsText() string {
	var names []string
	for _, target := range inject.DefaultNamedTargets() {
		names = append(names, target.Name)
	}
	return strings.Join(names, ", ")
}

func (handler *AgentModeHandler) boundTarget(sessionKey string) (inject.NamedTarget, bool) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	target, ok := handler.targets[sessionKey]
	return target, ok
}

func (handler *AgentModeHandler) setBoundTarget(sessionKey string, target inject.NamedTarget) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.targets[sessionKey] = target
}

func (handler *AgentModeHandler) injectPayload(sessionKey string, userText string) string {
	trimmed := strings.TrimSpace(userText)
	if strings.TrimSpace(sessionKey) == "" {
		return trimmed
	}
	return "你正在被 copyagent 远程托管。\n" +
		"真正需要完成的业务任务，只有下面这个任务块里的内容；本消息中的其他说明都只是回传协议，不是用户任务本身。\n\n" +
		"[BEGIN USER TASK]\n" + trimmed + "\n[END USER TASK]\n\n" +
		"强制要求：\n" +
		"1. 先立即回发第一条固定确认：桌面端 Coding App 已经收到，现在开始执行。\n" +
		"2. 然后在真正开始执行前，回发第二条计划，但计划必须围绕 [BEGIN USER TASK] 和 [END USER TASK] 之间的内容来拆解，不要把“确认收到”“调用 reply-text”“回传结果”、本消息里的标题、强制要求、命令模板、示例文案当成任务步骤。\n" +
		"3. 完成任务后，回发第三条完整版结果。\n" +
		"4. 三次回发都必须使用 shell/terminal 执行下面的 reply-text 命令，不要只在当前窗口回答。\n" +
		"5. 第三条必须和你在桌面端显示给用户的最终回复格式一致，保留真实换行、列表、代码块和缩进；不要把换行写成字面量 \\n 或 /n。\n" +
		"6. 如果最终内容包含多行，优先使用 --stdin 通过 heredoc 传入，避免 shell 参数转义破坏格式。\n" +
		"7. 第二条计划建议先用一句话点明你识别到的用户任务，再说明怎么做；如果用户任务本身很简单，例如“只回复 ok”，第二条计划也要直接说明会如何完成这个任务，例如“用户任务是‘只回复 ok’。这个任务很简单，我会直接回复 ok，然后立即回传最终结果。”，不要硬凑三步。\n" +
		"8. 第三条回发成功后，也要在本窗口显示同一份完整版结果，方便用户回看。\n\n" +
		"第一次回发（立刻执行，固定文案）：\n" +
		replyCommand(sessionKey, "桌面端 Coding App 已经收到，现在开始执行。") + "\n\n" +
		"第二次回发（执行前，替换为围绕用户任务本身的计划，不要描述 reply-text 回传流程；简单任务可以只回一句简短计划）：\n" +
		replyCommand(sessionKey, "我会这样完成这个任务：……") + "\n\n" +
		"例如，如果用户任务是“只回复 ok”，第二次回发应类似：\n" +
		replyCommand(sessionKey, "用户任务是“只回复 ok”。这个任务很简单，我会直接回复 ok，然后立即回传最终结果。") + "\n\n" +
		"第三次回发（完成后，推荐多行结果用 --stdin 保留真实格式）：\n" +
		replyStdinCommand(sessionKey) + " <<'COPYAGENT_REPLY'\n" +
		"<粘贴和桌面端最终回复完全一致的完整版结果，使用真实换行>\n" +
		"COPYAGENT_REPLY"
}

func replyCommand(sessionKey string, text string) string {
	return shellQuote(copyagentCommandPath()) + " action reply-text --session-key " + shellQuote(sessionKey) + " --text " + shellQuote(text)
}

func replyStdinCommand(sessionKey string) string {
	return shellQuote(copyagentCommandPath()) + " action reply-text --session-key " + shellQuote(sessionKey) + " --stdin"
}

func copyagentCommandPath() string {
	path, err := os.Executable()
	if err != nil || strings.TrimSpace(path) == "" {
		return "copyagentd"
	}
	return path
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func messageContent(msg *Message) string {
	if msg == nil {
		return ""
	}
	return msg.Content
}

func injectFailureReply(target inject.Target, err error) string {
	if errors.Is(err, inject.ErrEmptyText) {
		return "请在 /inject 后提供要粘贴的文本。"
	}
	if errors.Is(err, inject.ErrTextTooLarge) {
		return "未粘贴：文本太长。"
	}
	if errors.Is(err, inject.ErrUnsupportedTarget) {
		return "未粘贴：当前前台应用 " + targetLabel(target) + " 不在允许列表。请发送 /turn codex、/turn claude、/turn terminal、/turn iterm、/turn warp、/turn vscode 或 /turn cursor 后重试。"
	}
	return "未粘贴：" + err.Error()
}

func targetLabel(target inject.Target) string {
	if strings.TrimSpace(target.AppName) != "" {
		return target.AppName
	}
	if strings.TrimSpace(target.BundleID) != "" {
		return target.BundleID
	}
	return "未知应用"
}

func (handler *AgentModeHandler) shouldDirect(msg *Message) bool {
	if msg == nil {
		return true
	}
	if len(msg.Images) > 0 || len(msg.Files) > 0 {
		return true
	}
	return IsExplicitCopyCommand(msg.Content)
}

func (handler *AgentModeHandler) runAgentTurn(ctx context.Context, transport Transport, msg Message) {
	sessionKey := msg.EffectiveSessionKey()
	success := false
	stopTyping := startTyping(ctx, transport, msg.ReplyCtx)
	defer func() {
		stopTyping()
		handler.policy.Complete(&msg, success)
		next, ok := handler.queue.CompleteAndDequeue(sessionKey)
		if ok {
			go handler.runAgentTurn(ctx, transport, *next)
		}
	}()
	if handler.agent == nil {
		log.Printf("agent mode missing agent: session=%s message=%s", sessionKey, msg.MessageID)
		return
	}
	session, err := handler.session(ctx, sessionKey)
	if err != nil {
		log.Printf("agent mode session error: session=%s message=%s err=%v", sessionKey, msg.MessageID, err)
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "agent session error: "+err.Error())
		return
	}
	if err := session.Send(ctx, handler.agentPrompt(msg), AgentAttachments{Images: msg.Images, Files: msg.Files}); err != nil {
		log.Printf("agent mode send error: session=%s message=%s err=%v", sessionKey, msg.MessageID, err)
		_ = handler.reply(ctx, transport, msg.ReplyCtx, "agent send error: "+err.Error())
		return
	}
	var parts []string
	for event := range session.Events() {
		switch event.Type {
		case AgentEventSessionIDChanged:
			if event.SessionID != "" {
				handler.store.SetAgentSessionID(sessionKey, event.SessionID)
			}
		case AgentEventText:
			if strings.TrimSpace(event.Text) != "" {
				parts = append(parts, event.Text)
			}
		case AgentEventResult:
			if strings.TrimSpace(event.Text) != "" {
				parts = append(parts, event.Text)
			}
		case AgentEventError:
			if event.Error != nil {
				_ = handler.reply(ctx, transport, msg.ReplyCtx, "agent error: "+event.Error.Error())
				return
			}
		}
	}
	if len(parts) > 0 {
		if err := handler.reply(ctx, transport, msg.ReplyCtx, strings.Join(parts, "\n")); err != nil {
			log.Printf("agent mode reply error: session=%s message=%s err=%v", sessionKey, msg.MessageID, err)
			return
		}
		log.Printf("agent mode replied: session=%s message=%s parts=%d", sessionKey, msg.MessageID, len(parts))
	}
	success = true
}

func startTyping(ctx context.Context, transport Transport, replyCtx any) func() {
	typing, ok := transport.(TypingCapable)
	if !ok {
		return func() {}
	}
	stop := typing.StartTyping(ctx, replyCtx)
	if stop == nil {
		return func() {}
	}
	return stop
}

func (handler *AgentModeHandler) session(ctx context.Context, sessionKey string) (AgentSession, error) {
	handler.mu.Lock()
	cached := handler.sessions[sessionKey]
	handler.mu.Unlock()
	if cached != nil && cached.Alive() {
		return cached, nil
	}
	session, err := StartOrResumeSession(ctx, handler.agent, handler.store, sessionKey)
	if err != nil {
		return nil, err
	}
	handler.mu.Lock()
	handler.sessions[sessionKey] = session
	handler.mu.Unlock()
	return session, nil
}

func transportName(transport Transport) string {
	if transport == nil {
		return ""
	}
	return transport.Name()
}

func (handler *AgentModeHandler) agentPrompt(msg Message) string {
	content := strings.TrimSpace(msg.Content)
	if strings.TrimSpace(handler.prompt) == "" {
		return content
	}
	return strings.TrimSpace(handler.prompt) + "\n\n用户消息:\n" + content
}

func (handler *AgentModeHandler) reply(ctx context.Context, transport Transport, replyCtx any, content string) error {
	return replyToTransport(ctx, transport, replyCtx, content)
}
