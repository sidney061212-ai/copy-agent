package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/copyagent/copyagentd/internal/agent"
)

func TestParseCodexJSONEvents(t *testing.T) {
	tests := []struct {
		name string
		line string
		want agent.AgentEvent
		ok   bool
	}{
		{name: "thread", line: `{"type":"thread.started","thread_id":"thread-1"}`, want: agent.AgentEvent{Type: agent.AgentEventSessionIDChanged, SessionID: "thread-1"}, ok: true},
		{name: "text", line: `{"type":"item.completed","item":{"type":"agent_message","text":"hello"}}`, want: agent.AgentEvent{Type: agent.AgentEventText, Text: "hello"}, ok: true},
		{name: "result", line: `{"type":"turn.completed"}`, want: agent.AgentEvent{Type: agent.AgentEventResult}, ok: true},
		{name: "ignored", line: `{"type":"turn.started"}`, ok: false},
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

func TestParseCodexErrorEvent(t *testing.T) {
	event, ok := parseEvent([]byte(`{"type":"error","message":"boom"}`))
	if !ok || event.Type != agent.AgentEventError || event.Error == nil || event.Error.Error() != "boom" {
		t.Fatalf("event = %#v ok=%v", event, ok)
	}
}

func TestCodexSessionRunsExecJSONAndStoresThreadID(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is unix-only")
	}
	dir := t.TempDir()
	argsPath := filepath.Join(dir, "args.txt")
	stdinPath := filepath.Join(dir, "stdin.txt")
	command := writeFakeCodex(t, dir, `#!/bin/sh
printf '%s\n' "$@" > "`+argsPath+`"
cat > "`+stdinPath+`"
printf '%s\n' '{"type":"thread.started","thread_id":"thread-1"}'
printf '%s\n' '{"type":"item.completed","item":{"type":"agent_message","text":"OK"}}'
printf '%s\n' '{"type":"turn.completed"}'
`)
	codexAgent, err := New(Options{Command: command})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	session, err := codexAgent.StartSession(context.Background(), "")
	if err != nil {
		t.Fatalf("StartSession returned error: %v", err)
	}
	if err := session.Send(context.Background(), "hello codex", agent.AgentAttachments{}); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	events := collectEvents(t, session.Events())
	if len(events) != 3 || events[0].SessionID != "thread-1" || events[1].Text != "OK" || events[2].Type != agent.AgentEventResult {
		t.Fatalf("events = %#v", events)
	}
	if got := session.CurrentSessionID(); got != "thread-1" {
		t.Fatalf("CurrentSessionID = %q", got)
	}
	args := readLines(t, argsPath)
	wantArgs := []string{"exec", "--json", "--skip-git-repo-check", "-"}
	if strings.Join(args, "\x00") != strings.Join(wantArgs, "\x00") {
		t.Fatalf("args = %#v, want %#v", args, wantArgs)
	}
	stdin, err := os.ReadFile(stdinPath)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if string(stdin) != "hello codex" {
		t.Fatalf("stdin = %q", stdin)
	}
}

func TestCodexSessionResumeArgs(t *testing.T) {
	codexAgent := &Agent{command: "/bin/echo", args: []string{"--model", "gpt-test"}}
	sessionAny, err := codexAgent.StartSession(context.Background(), "thread-1")
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
	want := []string{"exec", "resume", "--model", "gpt-test", "--json", "--skip-git-repo-check", "thread-1", "-"}
	if strings.Join(command.args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("args = %#v, want %#v", command.args, want)
	}
}

func TestCodexSessionStagesAttachments(t *testing.T) {
	codexAgent := &Agent{command: "/bin/echo"}
	sessionAny, err := codexAgent.StartSession(context.Background(), "")
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
	if !containsArgPair(command.args, "--image") {
		t.Fatalf("expected --image arg, got %#v", command.args)
	}
	if !strings.Contains(command.prompt, "附件文件路径:") || !strings.Contains(command.prompt, "report.txt") {
		t.Fatalf("prompt = %q", command.prompt)
	}
}

func TestCodexSessionRejectsConcurrentTurns(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is unix-only")
	}
	dir := t.TempDir()
	command := writeFakeCodex(t, dir, "#!/bin/sh\ncat >/dev/null\nsleep 2\n")
	codexAgent := &Agent{command: command}
	sessionAny, err := codexAgent.StartSession(context.Background(), "")
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

func TestCodexSessionCloseCancelsProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell helper is unix-only")
	}
	dir := t.TempDir()
	command := writeFakeCodex(t, dir, "#!/bin/sh\ncat >/dev/null\nsleep 5\n")
	codexAgent := &Agent{command: command}
	sessionAny, err := codexAgent.StartSession(context.Background(), "")
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

func writeFakeCodex(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "fake-codex.sh")
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatalf("write fake codex: %v", err)
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

func containsArgPair(args []string, flag string) bool {
	for index := 0; index < len(args)-1; index++ {
		if args[index] == flag && strings.TrimSpace(args[index+1]) != "" {
			return true
		}
	}
	return false
}
