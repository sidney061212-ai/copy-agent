package claude

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
	command      string
	workDir      string
	args         []string
	systemPrompt string
}

type Options struct {
	Command      string
	WorkDir      string
	Args         []string
	SystemPrompt string
}

func New(opts Options) (*Agent, error) {
	command := strings.TrimSpace(opts.Command)
	if command == "" {
		command = "claude"
	}
	if filepath.Base(command) == command {
		if _, err := exec.LookPath(command); err != nil {
			return nil, err
		}
	}
	return &Agent{command: command, workDir: opts.WorkDir, args: append([]string(nil), opts.Args...), systemPrompt: opts.SystemPrompt}, nil
}

func NewFromOptions(opts map[string]any) (agent.CodingAgent, error) {
	var parsed Options
	if command, ok := opts["command"].(string); ok {
		parsed.Command = command
	}
	if workDir, ok := opts["workDir"].(string); ok {
		parsed.WorkDir = workDir
	}
	if systemPrompt, ok := opts["systemPrompt"].(string); ok {
		parsed.SystemPrompt = systemPrompt
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

func (claudeAgent *Agent) Name() string { return "claude" }

func (claudeAgent *Agent) StartSession(ctx context.Context, sessionID string) (agent.AgentSession, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return &Session{
		command:      claudeAgent.command,
		workDir:      claudeAgent.workDir,
		args:         append([]string(nil), claudeAgent.args...),
		systemPrompt: claudeAgent.systemPrompt,
		sessionID:    sessionID,
	}, nil
}

func (claudeAgent *Agent) Stop() error { return nil }

type Session struct {
	mu           sync.Mutex
	command      string
	workDir      string
	args         []string
	systemPrompt string
	sessionID    string
	events       chan agent.AgentEvent
	cancel       context.CancelFunc
	running      bool
}

func (session *Session) Send(ctx context.Context, prompt string, attachments agent.AgentAttachments) error {
	session.mu.Lock()
	if session.running {
		session.mu.Unlock()
		return errors.New("claude session turn already running")
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
	return errors.New("claude stream-json adapter does not support permission responses")
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
	args  []string
	input string
}

func (session *Session) buildCommand(prompt string, attachments agent.AgentAttachments) (commandInvocation, func(), error) {
	staged, cleanup, err := stageAttachments(attachments)
	if err != nil {
		return commandInvocation{}, nil, err
	}
	args := []string{"--print", "--input-format=stream-json", "--output-format=stream-json", "--verbose"}
	if session.sessionID != "" {
		args = append(args, "--resume", session.sessionID)
	}
	if strings.TrimSpace(session.systemPrompt) != "" {
		args = append(args, "--append-system-prompt", session.systemPrompt)
	}
	args = append(args, session.args...)
	if len(staged.filePaths) > 0 || len(staged.imagePaths) > 0 {
		prompt += "\n\n附件文件路径:\n"
		for _, path := range append(staged.imagePaths, staged.filePaths...) {
			prompt += "- " + path + "\n"
		}
	}
	input, err := userMessageJSON(prompt)
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return commandInvocation{}, nil, err
	}
	return commandInvocation{args: args, input: input + "\n"}, cleanup, nil
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
		_, _ = strings.NewReader(command.input).WriteTo(stdin)
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

func userMessageJSON(prompt string) (string, error) {
	payload := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": []map[string]string{{"type": "text", "text": prompt}},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type stagedAttachments struct {
	imagePaths []string
	filePaths  []string
}

func stageAttachments(attachments agent.AgentAttachments) (stagedAttachments, func(), error) {
	if len(attachments.Images) == 0 && len(attachments.Files) == 0 {
		return stagedAttachments{}, nil, nil
	}
	dir, err := os.MkdirTemp("", "copyagent-claude-*")
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
	Type      string          `json:"type"`
	Subtype   string          `json:"subtype"`
	SessionID string          `json:"session_id"`
	Result    string          `json:"result"`
	IsError   bool            `json:"is_error"`
	Message   json.RawMessage `json:"message"`
	Error     string          `json:"error"`
}

type rawMessage struct {
	Content []rawContent `json:"content"`
}

type rawContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func parseEvent(line []byte) (agent.AgentEvent, bool) {
	var raw rawEvent
	if err := json.Unmarshal(line, &raw); err != nil {
		return agent.AgentEvent{}, false
	}
	switch raw.Type {
	case "system":
		if raw.SessionID != "" {
			return agent.AgentEvent{Type: agent.AgentEventSessionIDChanged, SessionID: raw.SessionID}, true
		}
	case "assistant":
		text := assistantText(raw.Message)
		if text != "" {
			return agent.AgentEvent{Type: agent.AgentEventText, Text: text, SessionID: raw.SessionID}, true
		}
	case "result":
		if raw.IsError {
			message := raw.Result
			if message == "" {
				message = raw.Error
			}
			if message == "" {
				message = "claude result error"
			}
			return agent.AgentEvent{Type: agent.AgentEventError, Error: errors.New(message), SessionID: raw.SessionID}, true
		}
		return agent.AgentEvent{Type: agent.AgentEventResult, Text: raw.Result, SessionID: raw.SessionID}, true
	case "error":
		message := raw.Error
		if message == "" {
			message = raw.Result
		}
		if message != "" {
			return agent.AgentEvent{Type: agent.AgentEventError, Error: errors.New(message), SessionID: raw.SessionID}, true
		}
	}
	return agent.AgentEvent{}, false
}

func assistantText(raw json.RawMessage) string {
	var message rawMessage
	if err := json.Unmarshal(raw, &message); err != nil {
		return ""
	}
	var parts []string
	for _, content := range message.Content {
		if content.Type == "text" && content.Text != "" {
			parts = append(parts, content.Text)
		}
	}
	return strings.Join(parts, "\n")
}
