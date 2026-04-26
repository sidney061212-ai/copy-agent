# Agent Mode Design

## Product Model

copyagent has two modes:

```text
Direct Mode: agent.enabled=false
Feishu/chat -> deterministic copy/file/image actions

Agent Mode: agent.enabled=true
Feishu/chat -> Claude/Codex CLI session -> copyagent actions -> desktop/CLI continuation
```

The distinction is intentionally simple for users:

- Off: copyagent is a lightweight clipboard bot.
- On: copyagent is a remote entry point into a coding CLI agent plus desktop clipboard actions.

## Why use the underlying CLI binary

Agent Mode should follow cc-connect's lightweight principle: use the CLI backend (`codex`, `claude`) and do not require desktop apps to stay open.

Benefits:

- Lower baseline than desktop apps.
- Works headlessly/backgrounded.
- Reuses the user's existing Claude/Codex credentials and CLI setup.
- Keeps copyagent focused on bridge + actions instead of becoming a full model provider.

## Routing Rules

Agent Mode does not mean every message must pay LLM latency.

1. Explicit deterministic commands fast-path:
   - `复制 hello`
   - `复制：hello`
   - `copy hello`
   - image copy
   - file save
2. Non-command natural language routes to the configured CLI agent.
3. Attachments are saved locally and referenced in the agent prompt.
4. Agent output is replied to the original chat transport.
5. Agent actions use copyagent's narrow action API/CLI.

## Agent Prompt Contract

copyagent should inject a short system prompt fragment, similar to cc-connect's `AgentSystemPrompt`, adapted for clipboard actions:

```text
You are running inside copyagent.
The user's message came from a chat/mobile entry point.
Use normal text replies for messages that should be sent back to the user.
For local Mac actions, use the copyagent action CLI:

  copyagent action copy-text "text to copy"
  copyagent action copy-image --from /absolute/path/image.png
  copyagent action save-file --from /absolute/path/file.pdf
  copyagent action status

Do not use arbitrary shell commands for clipboard/file actions when copyagent actions are available.
```

## Minimal Agent Interfaces

```go
type CodingAgent interface {
    Name() string
    StartSession(ctx context.Context, sessionID string) (AgentSession, error)
    Stop() error
}

type AgentSession interface {
    Send(ctx context.Context, input AgentInput) error
    Events() <-chan AgentEvent
    CurrentSessionID() string
    Alive() bool
    Close() error
}
```

## Initial Adapters

- `codex`: invokes the Codex CLI binary.
- `claude`: invokes Claude Code CLI.

The first implementation can be conservative:

- Start one persistent session per chat session key.
- Add idle timeout cleanup.
- Stream or collect output depending on what the CLI supports reliably.
- Defer advanced permission handling until the basic path works.

## Safety Boundaries

- Keep explicit commands deterministic and dependency-free.
- Keep clipboard/file actions behind copyagent executor.
- Redact secrets from prompts/logs/status.
- Do not grant broad shell freedom as the default action path.
- Limit attachment size and saved file locations.
