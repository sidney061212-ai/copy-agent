package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/copyagent/copyagentd/internal/inject"
)

type scriptedAgent struct {
	startIDs []string
	sessions []*scriptedSession
}

func (agent *scriptedAgent) Name() string { return "scripted" }
func (agent *scriptedAgent) StartSession(_ context.Context, sessionID string) (AgentSession, error) {
	agent.startIDs = append(agent.startIDs, sessionID)
	session := &scriptedSession{id: sessionID, events: make(chan AgentEvent, 8)}
	if session.id == "" {
		session.id = "agent-session-1"
	}
	agent.sessions = append(agent.sessions, session)
	return session, nil
}
func (agent *scriptedAgent) Stop() error { return nil }

type scriptedSession struct {
	id      string
	prompts []string
	events  chan AgentEvent
	alive   bool
}

func (session *scriptedSession) Send(_ context.Context, prompt string, _ AgentAttachments) error {
	session.prompts = append(session.prompts, prompt)
	session.alive = true
	session.events <- AgentEvent{Type: AgentEventSessionIDChanged, SessionID: session.id}
	session.events <- AgentEvent{Type: AgentEventText, Text: "agent reply"}
	session.events <- AgentEvent{Type: AgentEventResult}
	close(session.events)
	session.alive = false
	return nil
}
func (session *scriptedSession) RespondPermission(context.Context, string, PermissionResult) error {
	return nil
}
func (session *scriptedSession) Events() <-chan AgentEvent { return session.events }
func (session *scriptedSession) CurrentSessionID() string  { return session.id }
func (session *scriptedSession) Alive() bool               { return session.alive }
func (session *scriptedSession) Close() error              { return nil }

func TestAgentModeHandlerDirectFastPathForCopyCommand(t *testing.T) {
	clipboard := &mockTextClipboard{}
	transport := &mockReplyTransport{name: "feishu"}
	codingAgent := &scriptedAgent{}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct: DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: clipboard}},
		Agent:  codingAgent,
	})
	if err := handler.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", MessageID: "om_1", Content: "copy hi"}); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(clipboard.texts) != 1 || clipboard.texts[0] != "hi" {
		t.Fatalf("clipboard texts = %#v", clipboard.texts)
	}
	if len(codingAgent.sessions) != 0 {
		t.Fatalf("agent should not be used: %#v", codingAgent.sessions)
	}
}

func TestAgentModeHandlerRoutesNaturalLanguageToAgentAndReplies(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	codingAgent := &scriptedAgent{}
	store := NewMemorySessionStore()
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:       DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:        codingAgent,
		Store:        store,
		SystemPrompt: "bridge rules",
	})
	msg := &Message{Platform: "feishu", SessionKey: "feishu:c1:u1", MessageID: "om_1", UserID: "u1", Content: "帮我继续", ReplyCtx: "om_1"}
	if err := handler.HandleMessage(context.Background(), transport, msg); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	waitForReplies(t, transport, 1)
	if transport.replies[0] != "agent reply" {
		t.Fatalf("reply = %#v", transport.replies)
	}
	if len(codingAgent.sessions) != 1 || len(codingAgent.sessions[0].prompts) != 1 || codingAgent.sessions[0].prompts[0] != "bridge rules\n\n用户消息:\n帮我继续" {
		t.Fatalf("sessions = %#v", codingAgent.sessions)
	}
	if got, ok := store.AgentSessionID("feishu:c1:u1"); !ok || got != "agent-session-1" {
		t.Fatalf("stored session = %q %v", got, ok)
	}
}

func TestAgentModeHandlerRoutesInjectCommandWithoutAgent(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	codingAgent := &scriptedAgent{}
	injector := &mockInjector{result: inject.Result{Target: inject.Target{AppName: "Terminal", BundleID: "com.apple.Terminal"}, RestoredClipboard: true}}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    codingAgent,
		Injector: injector,
	})
	msg := &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", UserID: "u1", Content: "/inject hello from feishu", ReplyCtx: "om_1"}
	if err := handler.HandleMessage(context.Background(), transport, msg); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(injector.requests) != 1 || !injector.requests[0].Submit || !strings.Contains(injector.requests[0].Text, "桌面端 Coding App 已经收到，现在开始执行。") || !strings.Contains(injector.requests[0].Text, "真正需要完成的业务任务，只有下面这个任务块里的内容") || !strings.Contains(injector.requests[0].Text, "[BEGIN USER TASK]") || !strings.Contains(injector.requests[0].Text, "[END USER TASK]") || !strings.Contains(injector.requests[0].Text, "计划必须围绕 [BEGIN USER TASK] 和 [END USER TASK] 之间的内容来拆解") || !strings.Contains(injector.requests[0].Text, "不要把“确认收到”“调用 reply-text”“回传结果”、本消息里的标题、强制要求、命令模板、示例文案当成任务步骤") || !strings.Contains(injector.requests[0].Text, "用户任务是“只回复 ok”。这个任务很简单，我会直接回复 ok，然后立即回传最终结果。") || !strings.Contains(injector.requests[0].Text, "第三次回发（完成后") || !strings.Contains(injector.requests[0].Text, "--stdin") || !strings.Contains(injector.requests[0].Text, "hello from feishu") {
		t.Fatalf("inject requests = %#v", injector.requests)
	}
	if len(codingAgent.sessions) != 0 {
		t.Fatalf("agent should not be used: %#v", codingAgent.sessions)
	}
	if len(transport.replies) != 1 || transport.replies[0] != "已提交到前台应用：Terminal，等待回程回复。" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestAgentModeHandlerTurnCommandBindsAndInjectUsesTarget(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{result: inject.Result{Target: inject.Target{AppName: "Codex", BundleID: "com.openai.codex"}, RestoredClipboard: true}}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    &scriptedAgent{},
		Injector: injector,
	})
	ctx := context.Background()
	if err := handler.HandleMessage(ctx, transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "/turn codex", ReplyCtx: "om_1"}); err != nil {
		t.Fatalf("turn HandleMessage returned error: %v", err)
	}
	if len(injector.activated) != 1 || injector.activated[0] != "com.openai.codex" {
		t.Fatalf("activated = %#v", injector.activated)
	}
	if err := handler.HandleMessage(ctx, transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_2", Content: "/inject hello", ReplyCtx: "om_2"}); err != nil {
		t.Fatalf("inject HandleMessage returned error: %v", err)
	}
	if len(injector.requests) != 1 || injector.requests[0].TargetBundleID != "com.openai.codex" {
		t.Fatalf("inject requests = %#v", injector.requests)
	}
}

func TestAgentModeHandlerTurnStatusReplies(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{result: inject.Result{Target: inject.Target{AppName: "Codex", BundleID: "com.openai.codex"}}}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    &scriptedAgent{},
		Injector: injector,
	})
	if err := handler.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "/turn status", ReplyCtx: "om_1"}); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(transport.replies) != 1 || !strings.Contains(transport.replies[0], "当前前台应用：Codex") {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestAgentModeHandlerFullwidthTurnCommandWorks(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{result: inject.Result{Target: inject.Target{AppName: "Codex", BundleID: "com.openai.codex"}}}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    &scriptedAgent{},
		Injector: injector,
	})
	if err := handler.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "／turn codex", ReplyCtx: "om_1"}); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(injector.activated) != 1 || injector.activated[0] != "com.openai.codex" {
		t.Fatalf("activated = %#v", injector.activated)
	}
}

func TestAgentModeHandlerTurnUnknownTargetRepliesAvailableTargets(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    &scriptedAgent{},
		Injector: injector,
	})
	if err := handler.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "/turn notes", ReplyCtx: "om_1"}); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(injector.activated) != 0 {
		t.Fatalf("unknown /turn should not activate target: %#v", injector.activated)
	}
	if len(transport.replies) != 1 || transport.replies[0] != "未知目标：notes。可用目标：codex, claude, terminal, iterm, warp, vscode, cursor" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestAgentModeHandlerLegacyTargetRepliesMigrationHint(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{}
	codingAgent := &scriptedAgent{}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    codingAgent,
		Injector: injector,
	})
	if err := handler.HandleMessage(context.Background(), transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "/target codex", ReplyCtx: "om_1"}); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(injector.activated) != 0 {
		t.Fatalf("legacy /target should not activate target: %#v", injector.activated)
	}
	if len(codingAgent.sessions) != 0 {
		t.Fatalf("legacy /target should not route to agent: %#v", codingAgent.sessions)
	}
	if len(transport.replies) != 1 || transport.replies[0] != "命令已更新：请改用 /turn codex。" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestAgentModeHandlerInjectEmptyTextReplies(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{err: inject.ErrEmptyText}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    &scriptedAgent{},
		Injector: injector,
	})
	msg := &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "/inject", ReplyCtx: "om_1"}
	if err := handler.HandleMessage(context.Background(), transport, msg); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(transport.replies) != 1 || transport.replies[0] != "请在 /inject 后提供要粘贴的文本。" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestAgentModeHandlerInjectWrongTargetReplies(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	injector := &mockInjector{result: inject.Result{Target: inject.Target{AppName: "Notes", BundleID: "com.apple.Notes"}}, err: inject.ErrUnsupportedTarget}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct:   DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:    &scriptedAgent{},
		Injector: injector,
	})
	msg := &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "/inject hello", ReplyCtx: "om_1"}
	if err := handler.HandleMessage(context.Background(), transport, msg); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if len(transport.replies) != 1 || transport.replies[0] != "未粘贴：当前前台应用 Notes 不在允许列表。请发送 /turn codex、/turn claude、/turn terminal、/turn iterm、/turn warp、/turn vscode 或 /turn cursor 后重试。" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

func TestAgentModeHandlerStartsAndStopsTyping(t *testing.T) {
	transport := &mockTypingTransport{mockReplyTransport: mockReplyTransport{name: "feishu"}}
	codingAgent := &scriptedAgent{}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct: DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:  codingAgent,
	})
	msg := &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "帮我继续", ReplyCtx: "om_1"}
	if err := handler.HandleMessage(context.Background(), transport, msg); err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	waitForReplies(t, &transport.mockReplyTransport, 1)
	if transport.typingStarted != 1 || transport.typingStopped != 1 {
		t.Fatalf("typing started=%d stopped=%d", transport.typingStarted, transport.typingStopped)
	}
}

func TestAgentModeHandlerQueuesWhileBusy(t *testing.T) {
	transport := &mockReplyTransport{name: "feishu"}
	session := &blockingSession{id: "agent-session-1", events: make(chan AgentEvent, 8), sent: make(chan string, 2)}
	codingAgent := &blockingAgent{session: session}
	handler := NewAgentModeHandler(AgentModeHandlerConfig{
		Direct: DirectHandlerConfig{Executor: DirectExecutorConfig{Clipboard: &mockTextClipboard{}}},
		Agent:  codingAgent,
		Queue:  NewSessionTurnQueue(2),
	})
	ctx := context.Background()
	if err := handler.HandleMessage(ctx, transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_1", Content: "first", ReplyCtx: "om_1"}); err != nil {
		t.Fatalf("first HandleMessage returned error: %v", err)
	}
	<-session.sent
	if err := handler.HandleMessage(ctx, transport, &Message{Platform: "feishu", SessionKey: "s", MessageID: "om_2", Content: "second", ReplyCtx: "om_2"}); err != nil {
		t.Fatalf("second HandleMessage returned error: %v", err)
	}
	if len(session.sent) != 0 {
		t.Fatal("second prompt should be queued while first is busy")
	}
	session.finish("reply one")
	waitForReplies(t, transport, 1)
	<-session.sent
	session.finish("reply two")
	waitForReplies(t, transport, 2)
	if transport.replies[0] != "reply one" || transport.replies[1] != "reply two" {
		t.Fatalf("replies = %#v", transport.replies)
	}
}

type blockingAgent struct{ session *blockingSession }

func (agent *blockingAgent) Name() string { return "blocking" }
func (agent *blockingAgent) StartSession(context.Context, string) (AgentSession, error) {
	return agent.session, nil
}
func (agent *blockingAgent) Stop() error { return nil }

type blockingSession struct {
	id     string
	events chan AgentEvent
	sent   chan string
	alive  bool
}

func (session *blockingSession) Send(_ context.Context, prompt string, _ AgentAttachments) error {
	session.alive = true
	session.events = make(chan AgentEvent, 8)
	session.sent <- prompt
	return nil
}
func (session *blockingSession) RespondPermission(context.Context, string, PermissionResult) error {
	return nil
}
func (session *blockingSession) Events() <-chan AgentEvent { return session.events }
func (session *blockingSession) CurrentSessionID() string  { return session.id }
func (session *blockingSession) Alive() bool               { return session.alive }
func (session *blockingSession) Close() error              { session.alive = false; return nil }
func (session *blockingSession) finish(text string) {
	session.events <- AgentEvent{Type: AgentEventSessionIDChanged, SessionID: session.id}
	session.events <- AgentEvent{Type: AgentEventText, Text: text}
	session.events <- AgentEvent{Type: AgentEventResult}
	close(session.events)
	session.alive = false
}

type mockTypingTransport struct {
	mockReplyTransport
	typingStarted int
	typingStopped int
}

type mockInjector struct {
	requests  []inject.Request
	activated []string
	result    inject.Result
	err       error
}

func (injector *mockInjector) InjectText(_ context.Context, req inject.Request) (inject.Result, error) {
	injector.requests = append(injector.requests, req)
	return injector.result, injector.err
}

func (injector *mockInjector) Status(context.Context) (inject.Target, error) {
	return injector.result.Target, injector.err
}

func (injector *mockInjector) Activate(_ context.Context, bundleID string) (inject.Target, error) {
	injector.activated = append(injector.activated, bundleID)
	return injector.result.Target, injector.err
}

func (transport *mockTypingTransport) StartTyping(context.Context, any) func() {
	transport.typingStarted++
	return func() { transport.typingStopped++ }
}
