# Feature Specification: cc-connect-Style Pluggable Clipboard Agent

**Created**: 2026-04-25  
**Status**: Draft  
**Reference**: cc-connect 1.3.2 / `https://github.com/chenhg5/cc-connect`

## Goal

Refactor copyagent into a cc-connect-style pluggable local runtime with two clear product modes:

1. **Direct Mode**: `agent.enabled=false`; Feishu/chat messages trigger deterministic clipboard and file actions directly.
2. **Agent Mode**: `agent.enabled=true`; Feishu/chat messages route into a Claude/Codex CLI session, which can continue user work and call copyagent actions.

The current Feishu clipboard behavior remains the stable Direct Mode. Agent Mode is the next product direction.

## Product Vision

copyagent is a bridge from chat/mobile input into the user's Mac workspace:

```text
Direct Mode:
Feishu/chat -> copyagent rules -> clipboard/files/reply

Agent Mode:
Feishu/chat -> copyagent -> Claude/Codex CLI session -> copyagent action API/CLI -> clipboard/files/reply/continued work
```

The user should not need to keep a desktop Claude/Codex app open. Agent Mode should use underlying coding CLI binaries, similar to cc-connect's lightweight approach.

## Reference Architecture

cc-connect patterns to reuse:

- Plugin packages register themselves through a central registry.
- `cmd/.../plugin_*.go` files use blank imports and optional build tags.
- A central engine owns transport startup and message handling.
- A unified message model separates platform parsing from downstream processing.
- A CLI-agent adapter can maintain/resume lightweight sessions without depending on desktop apps.

copyagent should reuse cc-connect's shape, but stay smaller and clipboard-first.

## Current copyagent Problem

`go-copyagentd/internal/transport/feishu/handler.go` currently combines:

- Feishu SDK event parsing
- message normalization
- actor policy
- command planning
- file saving
- clipboard writes
- reply sending

This works for one platform but blocks Agent Mode and future chat transports.

## Target Runtime Flow

```text
copyagentd main
  -> load config
  -> create Engine
  -> create registered transports from config
  -> transport receives platform event
  -> transport normalizes to agent.Message
  -> Engine.HandleMessage
  -> if explicit copy/file command: fast-path direct action
  -> else if agent.enabled: send to AgentSession
  -> else: deterministic planner/executor
  -> reply through transport when needed
```

## Target Packages

```text
go-copyagentd/internal/agent/message.go
go-copyagentd/internal/agent/registry.go
go-copyagentd/internal/agent/engine.go
go-copyagentd/internal/agent/policy.go
go-copyagentd/internal/agent/planner.go
go-copyagentd/internal/agent/executor.go
go-copyagentd/internal/agent/session.go
go-copyagentd/internal/agent/api.go
go-copyagentd/internal/agent/codex/...
go-copyagentd/internal/agent/claude/...
go-copyagentd/internal/transport/feishu/...
cmd/copyagentd/plugin_transport_feishu.go
cmd/copyagentd/plugin_agent_codex.go
cmd/copyagentd/plugin_agent_claude.go
```

## Core Transport Interfaces

```go
type Transport interface {
    Name() string
    Start(handler MessageHandler) error
    Stop() error
}

type MessageHandler func(t Transport, msg *Message)

type ReplyCapable interface {
    Reply(ctx context.Context, replyCtx any, content string) error
}

type ResourceCapable interface {
    Download(ctx context.Context, ref ResourceRef) ([]byte, error)
}

type TransportFactory func(opts map[string]any) (Transport, error)
```

## Core Agent Interfaces

A small cc-connect-inspired subset, not the full cc-connect session system:

```go
type CodingAgent interface {
    Name() string
    StartSession(ctx context.Context, sessionID string) (AgentSession, error)
    Stop() error
}

type AgentSession interface {
    Send(ctx context.Context, msg AgentInput) error
    Events() <-chan AgentEvent
    CurrentSessionID() string
    Alive() bool
    Close() error
}

type AgentFactory func(opts map[string]any) (CodingAgent, error)
```

Agent adapters should support at least:

- `codex` CLI
- `claude` / Claude Code CLI

They should be optional and only started when `agent.enabled=true`.

## Normalized Message

```go
type Message struct {
    SessionKey string
    Platform   string
    MessageID  string
    UserID     string
    UserName   string
    Content    string
    Images     []ImageAttachment
    Files      []FileAttachment
    ReplyCtx   any
}
```

## Product Modes

### Direct Mode

Config:

```json
{
  "agent": { "enabled": false }
}
```

Behavior:

- Explicit text copy commands copy directly.
- Images/files save and copy according to deterministic rules.
- No Claude/Codex CLI session is started.
- This is the current stable behavior.

### Agent Mode

Config:

```json
{
  "agent": {
    "enabled": true,
    "type": "codex",
    "command": "codex",
    "sessionMode": "persistent",
    "idleTimeoutMins": 15
  }
}
```

Behavior:

- Explicit deterministic commands can still fast-path directly.
- Natural-language work messages are forwarded to the configured CLI agent.
- Images/files are saved locally and passed to the CLI agent as local file paths.
- The CLI agent receives a copyagent system prompt explaining how to call copyagent actions.
- Agent output is replied back through the originating transport.
- The desktop app is not required to stay open.

## Copyagent Action API/CLI

Agent Mode needs a narrow action surface for CLI agents:

```text
copyagent action copy-text "..."
copyagent action save-file --from <path>
copyagent action copy-image --from <path>
copyagent action status
```

Internally these should call the daemon or shared executor. The agent should not need arbitrary shell access for clipboard/file actions.

## Config Direction

Short term: keep existing JSON config and add `agent` block.

Medium term: consider TOML if project/workspace profiles become necessary. Do not force migration yet.

## Acceptance Criteria

1. Direct Mode preserves current Feishu text/image/file behavior.
2. Agent Mode can be enabled by config without requiring desktop Claude/Codex apps.
3. Coding agent adapters are registered using cc-connect-style blank-import plugins.
4. Feishu transport emits normalized `agent.Message` and no longer directly owns business execution in the final refactor.
5. Explicit `复制 hello`/`copy hello` fast-paths avoid unnecessary LLM latency.
6. Non-command messages in Agent Mode are sent to the configured CLI agent session.
7. Saved image/file paths can be passed to the CLI agent.
8. The CLI agent can call copyagent's narrow action API/CLI to update clipboard/files.
9. Existing JSON config remains compatible.
10. Node remains reference/prototype only; no Node background service is restored.

## Non-Goals

- Four-mode runtime matrix (`api`, `bot`, `hybrid`, `auto`). The product has only Direct Mode and Agent Mode.
- Dynamic shared-object plugins.
- Full cc-connect cron/relay/TTS/speech/run-as-user/management stack.
- Public network API.
- Forcing an LLM for simple deterministic copy/file commands.
