package claude

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/copyagent/copyagentd/internal/agent"
)

func TestParseClaudeStreamJSONEvents(t *testing.T) {
	tests := []struct {
		name string
		line string
		want agent.AgentEvent
		ok   bool
	}{
		{name: "system", line: `{"type":"system","subtype":"init","session_id":"session-1"}`, want: agent.AgentEvent{Type: agent.AgentEventSessionIDChanged, SessionID: "session-1"}, ok: true},
		{name: "assistant", line: `{"type":"assistant","session_id":"session-1","message":{"content":[{"type":"text","text":"hello"}]}}`, want: agent.AgentEvent{Type: agent.AgentEventText, Text: "hello", SessionID: "session-1"}, ok: true},
		{name: "result", line: `{"type":"result","session_id":"session-1","is_error":false,"result":"done"}`, want: agent.AgentEvent{Type: agent.AgentEventResult, Text: "done", SessionID: "session-1"}, ok: true},
		{name: "ignored", line: `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`, ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseEvent([]byte(tt.line))
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if !ok {
				return
			}
			if got.Type != tt.want.Type || got.Text != tt.want.Text || got.SessionID != tt.want.SessionID {
				t.Fatalf("event = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseClaudeErrorResult(t *testing.T) {
	event, ok := parseEvent([]byte(`{"type":"result","is_error":true,"result":"boom"}`))
	if !ok || event.Type != agent.AgentEventError || event.Error == nil || event.Error.Error() != "boom" {
		t.Fatalf("event = %#v ok=%v", event, ok)
	}
}

func TestUserMessageJSON(t *testing.T) {
	line, err := userMessageJSON("hello")
	if err != nil {
		t.Fatalf("userMessageJSON returned error: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if payload["type"] != "user" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestClaudeSessionRunsStreamJSONAndStoresSessionID(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is unix-only")
	}
	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args.txt")
	stdinPath := filepath.Join(dir, "stdin.txt")
	command := writeFakeClaude(t, dir, `#!/bin/sh
printf '%s\n' "$@" > "`+argsPath+`"
cat > "`+stdinPath+`"
printf '%s\n' '{"type":"system","subtype":"init","session_id":"session-1"}'
printf '%s\n' '{"type":"assistant","session_id":"session-1","message":{"content":[{"type":"text","text":"OK"}]}}'
printf '%s\n' '{"type":"result","session_id":"session-1","is_error":false,"result":"OK"}'
`)
	claudeAgent, err := New(Options{Command: command})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	session, err := claudeAgent.StartSession(context.Background(), "")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	if err := session.Send(context.Background(), "hello claude", agent.AgentAttachments{}); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	events := collectEvents(t, session.Events())
	if len(events) != 3 || events[0].SessionID != "session-1" || events[1].Text != "OK" || events[2].Type != agent.AgentEventResult {
		t.Fatalf("events = %#v", events)
	}
	if got := session.CurrentSessionID(); got != "session-1" {
		t.Fatalf("CurrentSessionID = %q", got)
	}
	args := readLines(t, argsPath)
	wantArgs := []string{"--print", "--input-format=stream-json", "--output-format=stream-json", "--verbose"}
	if strings.Join(args, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %#v, want %#v", args, wantArgs)
	}
	stdin, err := os.ReadFile(stdinPath)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if !strings.Contains(string(stdin), "hello claude") || !strings.Contains(string(stdin), `"type":"user"`) {
		t.Fatalf("stdin = %q", stdin)
	}
}

func TestClaudeSessionResumeArgsAndSystemPrompt(t *testing.T) {
	claudeAgent := &Agent{command: "/bin/echo", args: []string{"--model", "sonnet"}, systemPrompt: "copyagent bridge"}
	sessionAny, err := claudeAgent.StartSession(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	session := sessionAny.(*Session)
	command, cleanup, err := session.buildCommand("resume", agent.AgentAttachments{})
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		t.Fatalf("buildCommand returned error: %v", err)
	}
	want := []string{"--print", "--input-format=stream-json", "--output-format=stream-json", "--verbose", "--resume", "session-1", "--append-system-prompt", "copyagent bridge", "--model", "sonnet"}
	if strings.Join(command.args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestClaudeSessionStagesAttachments(t *testing.T) {
	claudeAgent := &Agent{command: "/bin/echo"}
	sessionAny, err := claudeAgent.StartSession(context.Background(), "")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	session := sessionAny.(*Session)
	command, cleanup, err := session.buildCommand("inspect", agent.AgentAttachments{
		Images: []agent.ImageAttachment{{FileName: "photo.png", Data: []byte("png")}},
		Files:  []agent.FileAttachment{{FileName: "report.txt", Data: []byte("file")}},
	})
	if err != nil {
		t.Fatalf("buildCommand returned error: %v", err)
	}
	defer cleanup()
	if !strings.Contains(command.input, "附件文件路径:") || !strings.Contains(command.input, "photo.png") || !strings.Contains(command.input, "report.txt") {
		t.Fatalf("input = %q", command.input)
	}
}

func TestClaudeSessionRejectsConcurrentTurns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is unix-only")
	}
	dir := t.TempDir()
	command := writeFakeClaude(t, dir, "#!/bin/sh\ncat >/dev/null\nsleep 2\n")
	claudeAgent := &Agent{command: command}
	sessionAny, err := claudeAgent.StartSession(context.Background(), "")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	if err := sessionAny.Send(context.Background(), "first", agent.AgentAttachments{}); err != nil {
		t.Fatalf("first Send returned error: %v", err)
	}
	if err := sessionAny.Send(context.Background(), "second", agent.AgentAttachments{}); err == nil {
		t.Fatal("expected concurrent turn error")
	}
	_ = sessionAny.Close()
}

func TestClaudeSessionCloseCancelsProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is unix-only")
	}
	dir := t.TempDir()
	command := writeFakeClaude(t, dir, "#!/bin/sh\ncat >/dev/null\nsleep 5\n")
	claudeAgent := &Agent{command: command}
	sessionAny, err := claudeAgent.StartSession(context.Background(), "")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	if err := sessionAny.Send(context.Background(), "first", agent.AgentAttachments{}); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if err := sessionAny.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	deadline := time.After(3 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("session did not stop after Close")
		default:
			if !sessionAny.Alive() {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func collectEvents(t *testing.T, ch <-chan agent.AgentEvent) []agent.AgentEvent {
	t.Helper()
	var events []agent.AgentEvent
	for event := range ch {
		if event.Type == agent.AgentEventError && event.Error != nil && !errors.Is(event.Error, context.Canceled) {
			t.Fatalf("unexpected error event: %v", event.Error)
		}
		events = append(events, event)
	}
	return events
}

func writeFakeClaude(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "fake-claude.sh")
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatalf("write fake claude: %v", err)
	}
	return path
}

func readLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read lines: %v", err)
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}
