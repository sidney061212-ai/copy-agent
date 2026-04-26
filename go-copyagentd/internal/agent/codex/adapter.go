package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/copyagent/copyagentd/internal/agent"
)

type Agent struct {
	command string
	workDir string
	args    []string
}

type Options struct {
	Command string
	WorkDir string
	Args    []string
}

func New(opts Options) (*Agent, error) {
	command := strings.TrimSpace(opts.Command)
	if command == "" {
		command = "codex"
	}
	if filepath.Base(command) == command {
		if _, err := exec.LookPath(command); err != nil {
			return nil, err
		}
	}
	return &Agent{command: command, workDir: opts.WorkDir, args: append([]string(nil), opts.Args...)}, nil
}

func NewFromOptions(opts map[string]any) (agent.CodingAgent, error) {
	var parsed Options
	if command, ok := opts["command"].(string); ok {
		parsed.Command = command
	}
	if workDir, ok := opts["workDir"].(string); ok {
		parsed.WorkDir = workDir
	}
	if rawArgs, ok := opts["args"].([]string); ok {
		parsed.Args = append(parsed.Args, rawArgs...)
	} else if rawArgs, ok := opts["args"].([]any); ok {
		for _, rawArg := range rawArgs {
			if arg, ok := rawArg.(string); ok {
				parsed.Args = append(parsed.Args, arg)
			}
		}
	}
	return New(parsed)
}

func (codexAgent *Agent) Name() string { return "codex" }

func (codexAgent *Agent) StartSession(ctx context.Context, sessionID string) (agent.AgentSession, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return &Session{command: codexAgent.command, workDir: codexAgent.workDir, args: append([]string(nil), codexAgent.args...), sessionID: sessionID}, nil
}

func (codexAgent *Agent) Stop() error { return nil }

type Session struct {
	mu        sync.Mutex
	command   string
	workDir   string
	args      []string
	sessionID string
	events    chan agent.AgentEvent
	cancel    context.CancelFunc
	running   bool
}

func (session *Session) Send(ctx context.Context, prompt string, attachments agent.AgentAttachments) error {
	session.mu.Lock()
	if session.running {
		session.mu.Unlock()
		return errors.New("codex session turn already running")
	}
	events := make(chan agent.AgentEvent, 32)
	session.events = events
	turnCtx, cancel := context.WithCancel(ctx)
	session.cancel = cancel
	session.running = true
	command, cleanup, err := session.buildCommand(prompt, attachments)
	if err != nil {
		session.running = false
		session.cancel = nil
		session.mu.Unlock()
		close(events)
		return err
	}
	session.mu.Unlock()

	go session.run(turnCtx, command, cleanup, events)
	return nil
}

func (session *Session) RespondPermission(context.Context, string, agent.PermissionResult) error {
	return errors.New("codex exec adapter does not support permission responses")
}

func (session *Session) Events() <-chan agent.AgentEvent {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.events
}

func (session *Session) CurrentSessionID() string {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.sessionID
}

func (session *Session) Alive() bool {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.running
}

func (session *Session) Close() error {
	session.mu.Lock()
	cancel := session.cancel
	session.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

type commandInvocation struct {
	args   []string
	prompt string
}

func (session *Session) buildCommand(prompt string, attachments agent.AgentAttachments) (commandInvocation, func(), error) {
	staged, cleanup, err := stageAttachments(attachments)
	if err != nil {
		return commandInvocation{}, nil, err
	}
	args := []string{"exec"}
	if session.sessionID != "" {
		args = append(args, "resume")
	}
	args = append(args, session.args...)
	for _, imagePath := range staged.imagePaths {
		args = append(args, "--image", imagePath)
	}
	args = append(args, "--json", "--skip-git-repo-check")
	if session.sessionID != "" {
		args = append(args, session.sessionID)
	}
	args = append(args, "-")
	if len(staged.filePaths) > 0 {
		prompt += "\n\n附件文件路径:\n"
		for _, filePath := range staged.filePaths {
			prompt += "- " + filePath + "\n"
		}
	}
	return commandInvocation{args: args, prompt: prompt}, cleanup, nil
}

func (session *Session) run(ctx context.Context, command commandInvocation, cleanup func(), events chan agent.AgentEvent) {
	defer func() {
		if cleanup != nil {
			cleanup()
		}
		session.mu.Lock()
		session.running = false
		session.cancel = nil
		session.mu.Unlock()
		close(events)
	}()
	cmd := exec.CommandContext(ctx, session.command, command.args...)
	cmd.Dir = session.workDir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		events <- agent.AgentEvent{Type: agent.AgentEventError, Error: err}
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		events <- agent.AgentEvent{Type: agent.AgentEventError, Error: err}
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		events <- agent.AgentEvent{Type: agent.AgentEventError, Error: err}
		return
	}
	if err := cmd.Start(); err != nil {
		events <- agent.AgentEvent{Type: agent.AgentEventError, Error: err}
		return
	}
	go func() {
		_, _ = strings.NewReader(command.prompt).WriteTo(stdin)
		_ = stdin.Close()
	}()
	go drainStderr(stderr)
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		event, ok := parseEvent(scanner.Bytes())
		if !ok {
			continue
		}
		if event.SessionID != "" {
			session.mu.Lock()
			session.sessionID = event.SessionID
			session.mu.Unlock()
		}
		events <- event
	}
	if err := scanner.Err(); err != nil {
		events <- agent.AgentEvent{Type: agent.AgentEventError, Error: err}
	}
	if err := cmd.Wait(); err != nil && ctx.Err() == nil {
		events <- agent.AgentEvent{Type: agent.AgentEventError, Error: err}
	}
}

func drainStderr(stderr interface{ Read([]byte) (int, error) }) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
	}
}

type stagedAttachments struct {
	prompt     string
	imagePaths []string
	filePaths  []string
}

func stageAttachments(attachments agent.AgentAttachments) (stagedAttachments, func(), error) {
	if len(attachments.Images) == 0 && len(attachments.Files) == 0 {
		return stagedAttachments{}, nil, nil
	}
	dir, err := os.MkdirTemp("", "copyagent-codex-*")
	if err != nil {
		return stagedAttachments{}, nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	var staged stagedAttachments
	for index, image := range attachments.Images {
		name := safeStageName(image.FileName, fmt.Sprintf("image-%d.png", index+1))
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, image.Data, 0o600); err != nil {
			cleanup()
			return stagedAttachments{}, nil, err
		}
		staged.imagePaths = append(staged.imagePaths, path)
	}
	for index, file := range attachments.Files {
		name := safeStageName(file.FileName, fmt.Sprintf("file-%d", index+1))
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, file.Data, 0o600); err != nil {
			cleanup()
			return stagedAttachments{}, nil, err
		}
		staged.filePaths = append(staged.filePaths, path)
	}
	return staged, cleanup, nil
}

func safeStageName(name string, fallback string) string {
	cleaned := strings.TrimSpace(filepath.Base(name))
	if cleaned == "" || cleaned == "." || cleaned == string(filepath.Separator) {
		return fallback
	}
	return cleaned
}

type rawEvent struct {
	Type     string          `json:"type"`
	ThreadID string          `json:"thread_id"`
	Message  string          `json:"message"`
	Error    string          `json:"error"`
	Item     json.RawMessage `json:"item"`
}

type rawItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func parseEvent(line []byte) (agent.AgentEvent, bool) {
	var raw rawEvent
	if err := json.Unmarshal(line, &raw); err != nil {
		return agent.AgentEvent{}, false
	}
	switch raw.Type {
	case "thread.started":
		return agent.AgentEvent{Type: agent.AgentEventSessionIDChanged, SessionID: raw.ThreadID}, raw.ThreadID != ""
	case "item.completed":
		var item rawItem
		if err := json.Unmarshal(raw.Item, &item); err != nil {
			return agent.AgentEvent{}, false
		}
		if item.Type == "agent_message" && item.Text != "" {
			return agent.AgentEvent{Type: agent.AgentEventText, Text: item.Text}, true
		}
	case "turn.completed":
		return agent.AgentEvent{Type: agent.AgentEventResult}, true
	case "error":
		message := raw.Message
		if message == "" {
			message = raw.Error
		}
		if message != "" {
			return agent.AgentEvent{Type: agent.AgentEventError, Error: errors.New(message)}, true
		}
	}
	return agent.AgentEvent{}, false
}
