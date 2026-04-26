# cc-connect Architecture Study for copyagent

Source reviewed: `/tmp/cc-connect-src` at commit `3a9bb44` on 2026-04-25.

## Why this matters

copyagent should borrow cc-connect's proven bridge architecture instead of inventing a new agent runtime. The product goal is simpler than cc-connect, but the successful patterns are directly useful:

- compile-time plugin registration
- platform/transport normalization
- central engine orchestration
- persistent CLI agent sessions
- queueing while a session is busy
- reply context preservation
- safe attachment staging
- CLI system prompt/tool contract

## Core cc-connect shapes

### Platform interface

`core/interfaces.go`:

```go
type Platform interface {
    Name() string
    Start(handler MessageHandler) error
    Reply(ctx context.Context, replyCtx any, content string) error
    Send(ctx context.Context, replyCtx any, content string) error
    Stop() error
}

type MessageHandler func(p Platform, msg *Message)
```

copyagent mapping:

- Keep `Transport` equivalent to `Platform`.
- Add `Reply` to the core transport interface instead of a separate Feishu-only reply interface.
- `Send` can be optional later; not required for first copyagent slice.

### Agent interface

`core/interfaces.go`:

```go
type Agent interface {
    Name() string
    StartSession(ctx context.Context, sessionID string) (AgentSession, error)
    ListSessions(ctx context.Context) ([]AgentSessionInfo, error)
    Stop() error
}

type AgentSession interface {
    Send(prompt string, images []ImageAttachment, files []FileAttachment) error
    RespondPermission(requestID string, result PermissionResult) error
    Events() <-chan Event
    CurrentSessionID() string
    Alive() bool
    Close() error
}
```

copyagent mapping:

- Reuse this shape almost directly for Agent Mode.
- `ListSessions` can be optional for first version.
- `RespondPermission` can be deferred, but leave the interface extension point.
- `Events` should stream normalized agent events so engine can reply back to Feishu.

### Registry

`core/registry.go` is intentionally simple:

- `RegisterPlatform(name, factory)`
- `RegisterAgent(name, factory)`
- `CreatePlatform(name, opts)`
- `CreateAgent(name, opts)`
- list registered names

copyagent mapping:

- Extend current `internal/agent/registry.go` to include agent factories.
- Keep registration simple and compile-time.
- Use blank imports in `cmd/copyagentd/plugin_*.go` like cc-connect.

### Config and construction

`cmd/cc-connect/main.go` loops over configured projects:

1. `core.CreateAgent(proj.Agent.Type, buildAgentOptions(...))`
2. `core.CreatePlatform(pc.Type, opts)` for each platform
3. `core.NewEngine(proj.Name, agent, platforms, sessionFile, lang)`
4. configure engine knobs
5. start engines and local API

copyagent mapping:

- No multi-project/workspace first.
- One default runtime:
  - Direct Mode: no coding agent required.
  - Agent Mode: `CreateAgent(agent.type, opts)`.
  - Always create configured transports.
  - `Engine` decides fast-path vs agent routing.

## Message flow lessons

### Platform normalizes early

cc-connect Feishu downloads/normalizes platform payloads into `core.Message` before calling engine:

- session key is platform/chat/user/thread scoped
- reply context is carried as opaque `ReplyCtx`
- images/files become in-memory attachments
- text/post/audio/location become normalized fields

copyagent mapping:

- Feishu transport should normalize to `agent.Message` and stop owning business execution.
- Preserve opaque reply context so engine can reply without Feishu-specific logic.
- For copyagent, images/files may be either in-memory attachments or saved paths; first slice can keep in-memory and executor saves.

### Session key design is central

cc-connect derives stable session keys like:

```text
feishu:<chatID>:<userID>
feishu:<chatID>:root:<rootMessageID>
```

copyagent mapping:

- Use stable session keys for Agent Mode so Claude/Codex can continue work per Feishu chat/thread.
- Current `Message.EffectiveSessionKey()` is too simple; it should support chat/thread isolation.

### Busy session queueing

cc-connect never writes a second message to agent stdin mid-turn. If a session is busy, it queues metadata and sends queued messages only after `EventResult`.

copyagent mapping:

- Must copy this rule for Agent Mode.
- Without it, Codex/Claude CLI sessions can hang or merge turns incorrectly.
- Implement per-session busy lock and small queue before serious Agent Mode rollout.

### Session resume fallback

cc-connect tries to resume saved agent session ID, and if resume fails, clears it and starts fresh.

copyagent mapping:

- Agent Mode should persist the CLI session ID per copyagent session key.
- If resume fails, fall back to new session and keep user workflow alive.

### Event loop responsibilities

cc-connect event loop:

- streams text and tool/progress events
- persists session ID updates
- handles permission requests
- applies idle timeout
- handles final result and queued messages

copyagent mapping:

- First version can be simpler:
  - collect text/result events
  - persist session ID
  - reply final output to Feishu
  - apply idle timeout
  - queue messages while busy
- Permission handling can be later unless Claude/Codex adapter requires it immediately.

## Agent adapter lessons

### Codex adapter

cc-connect's Codex adapter:

- registers with `init()` as `codex`
- checks `codex` binary in PATH
- supports `exec` and `app_server` backends
- launches `codex exec --json` or `codex exec resume ... --json`
- stages images/files into `.cc-connect`
- parses JSON lines into normalized events
- stores Codex thread ID as session ID

copyagent mapping:

- Start with `exec` backend only.
- Use `codex exec --json` and resume by stored thread ID.
- Stage attachments under `~/.copyagent/attachments` or workdir `.copyagent`.
- Parse only text/result/error events first.

### Claude Code adapter

cc-connect's Claude adapter:

- keeps a persistent process with stdin/stdout JSON protocol
- sends messages as JSON lines to stdin
- parses assistant/system/result/permission events from stdout
- appends cc-connect system prompt
- filters environment such as `CLAUDECODE`
- supports permission modes

copyagent mapping:

- Reuse the stdin/stdout JSON process model.
- Inject a copyagent-specific system prompt.
- Filter nested-session env vars if needed.
- Defer advanced permission cards but keep event type for later.

### System prompt contract

cc-connect injects `AgentSystemPrompt()` so the coding agent knows it is inside a bridge and how to call bridge commands.

copyagent needs an equivalent prompt:

```text
You are running inside copyagent.
Use normal text replies for messages to the user.
For local Mac actions, use:
  copyagent action copy-text "..."
  copyagent action copy-image --from /absolute/path.png
  copyagent action save-file --from /absolute/path
Do not use arbitrary shell commands for clipboard/file actions when copyagent actions are available.
```

## Platform lessons

### Feishu implementation

cc-connect Feishu platform handles many message types:

- text
- image
- audio
- post/rich text
- file
- merge_forward
- card actions

copyagent should start with:

- text
- image
- file

But the normalized message structure should not block audio/post later.

### Reply context

cc-connect uses opaque `replyContext` and optional reconstruction from session key.

copyagent should copy:

- opaque `ReplyCtx` for immediate replies
- optional `ReconstructReplyCtx(sessionKey)` later for proactive/session replies

## What copyagent should not copy yet

Do not copy these in first implementation:

- multi-workspace project pool
- cron scheduler
- relay manager
- TTS/STT
- management web UI
- run-as-user isolation
- provider switching UI/cards
- complex slash command registry
- stream preview cards
- full permission UI

They solve cc-connect's broader product, not copyagent's immediate two-mode model.

## Revised copyagent implementation order

1. Extend registry with `CodingAgent` factory support.
2. Add session store keyed by `Message.SessionKey`.
3. Add Direct Mode fast-path planner/executor.
4. Refactor Feishu to normalized `agent.Message`.
5. Add per-session busy lock and queue.
6. Add Codex CLI exec adapter.
7. Add Claude Code persistent adapter.
8. Add copyagent action CLI for agent tool calls.
9. Add final Agent Mode routing and reply handling.

## Non-negotiable behavior

- Direct Mode must preserve current Feishu copy/image/file behavior.
- Agent Mode must not require desktop apps to stay open.
- Explicit copy/file commands must not pay LLM latency.
- Agent Mode must queue while a session is busy; do not write mid-turn.
- Secrets must not be printed in logs or prompts.
