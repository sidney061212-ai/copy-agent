# cc-connect Reference Notes for copyagent

Source reviewed: `/tmp/cc-connect-src` cloned from `https://github.com/chenhg5/cc-connect` on 2026-04-25.

## Patterns to Reuse

### 1. Compile-time plugin registration

cc-connect uses small `cmd/cc-connect/plugin_*.go` files with build tags and blank imports:

```go
//go:build !no_feishu
package main
import _ "github.com/chenhg5/cc-connect/platform/feishu"
```

Each plugin package registers itself in `init()` through central registry functions. This keeps the binary simple while allowing build-time feature selection.

copyagent adaptation:

```text
cmd/copyagentd/plugin_transport_feishu.go
cmd/copyagentd/plugin_action_clipboard.go
cmd/copyagentd/plugin_action_files.go
```

Use build tags later, but start with default built-ins.

### 2. Central registry

cc-connect has `core.RegisterPlatform`, `core.RegisterAgent`, `core.CreatePlatform`, `core.CreateAgent`, and listing functions in `core/registry.go`.

copyagent adaptation:

```go
type TransportFactory func(opts map[string]any) (Transport, error)
type ActionFactory func(opts map[string]any) (ActionHandler, error)

func RegisterTransport(name string, factory TransportFactory)
func RegisterAction(name string, factory ActionFactory)
func CreateTransport(name string, opts map[string]any) (Transport, error)
func ListRegisteredTransports() []string
func ListRegisteredActions() []string
```

### 3. Engine owns orchestration

cc-connect `Engine` wires one agent with multiple platforms and exposes a `MessageHandler` callback to platforms.

copyagent should use the same conceptual split:

```text
Engine
  - owns runtime config, policy, planner, executor, registered transports
  - exposes HandleEvent(ctx, event)
  - starts/stops transports
  - exposes status/sessions/capabilities to local API
```

### 4. Unified message/event model

cc-connect uses `core.Message` as the platform-independent inbound message. It carries platform, message id, user id, content, images, files, audio, location, reply context, etc.

copyagent should create a smaller clipboard-focused equivalent:

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

For copyagent, `SessionKey` can initially be `platform:userID` or `platform:chatID:userID`.

### 5. Local Unix socket API

cc-connect exposes local API over a Unix socket at `<data_dir>/run/api.sock` with `0600` permissions. This is better than exposing an HTTP port for local tooling.

copyagent adaptation:

- Prefer Unix socket API on macOS/Linux.
- Keep optional TCP loopback only for explicit debug compatibility.
- Socket path: `~/.copyagent/run/api.sock`.
- Commands/scripts can call this socket in future.

Initial endpoints should be simpler than cc-connect:

```text
GET  /status
GET  /plugins
POST /clipboard/text
POST /events
POST /events/plan
```

### 6. Config shape

cc-connect TOML uses global settings plus `[[projects]]`, each with `agent` and `platforms`.

copyagent does not need projects yet. Reuse the layered shape but keep one default project/runtime:

```toml
data_dir = "~/.copyagent"

[api]
enabled = true
socket = "~/.copyagent/run/api.sock"

[policy]
allowed_actor_ids = []
dedupe_cache_size = 1000
max_text_bytes = 200000

[actions]
default_download_dir = "~/Downloads/copyagent"
image_action = "clipboard"
reply_enabled = true

[[transports]]
type = "feishu"
enabled = true
app_id = "cli_xxx"
app_secret = "replace-with-secret"
```

Backward compatibility: read existing JSON config first, normalize into the new runtime structs, then add TOML support as a later migration if desired.

## Patterns Not to Copy Yet

- Multi-agent persistent sessions: copyagent is an action engine, not a coding-agent bridge.
- Cron, relay, management UI, run-as-user, speech/TTS: useful later but too large for first plugin slice.
- External websocket bridge: defer until local action API is stable.
- Full project/workspace model: unnecessary until users need multiple independent clipboard profiles.

## Copyagent-Specific Mapping

| cc-connect | copyagent |
|---|---|
| `Platform` | `Transport` that receives messages/events |
| `Agent` | Not needed initially; copyagent's planner/executor replaces this |
| `Engine` | Clipboard runtime engine |
| `Message` | Clipboard-focused normalized message |
| `APIServer` over Unix socket | Local copyagent API over Unix socket |
| `RegisterPlatform` | `RegisterTransport` |
| `RegisterAgent` | `RegisterAction` or no equivalent |
| `Reply(ctx, replyCtx, content)` | transport reply capability |
| `Send(ctx, replyCtx, content)` | optional proactive send, later |

## Recommended Architecture Revision

Instead of designing an abstract public REST API first, implement the cc-connect style local runtime:

```text
cmd/copyagentd
  plugin_transport_feishu.go      blank import registers Feishu

internal/agent
  message.go                      normalized Message + attachments
  registry.go                     transport/action registry
  engine.go                       start transports + handle messages
  planner.go                      message -> actions
  executor.go                     actions -> clipboard/files/reply
  api.go                          Unix socket local API

internal/transport/feishu
  init() RegisterTransport("feishu", NewFactory)
  SDK event -> agent.Message
  Reply/Download capabilities
```

This keeps copyagent aligned with cc-connect while staying much smaller.
