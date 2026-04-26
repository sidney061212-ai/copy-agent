# Code Review: Current Runtime Logic

Reviewed on 2026-04-25.

## Runtime Flow Today

```text
copyagentd main
  -> config.Load(v1 flat config)
  -> feishu.NewTransportWithHandler
  -> Feishu SDK websocket callback
  -> feishu.MessageHandler.Handle
  -> NormalizeResourceMessage or NormalizeTextMessage
  -> actor allowlist check
  -> core.ExtractCopyText / saveResourceFile
  -> clipboard.WriteText / clipboard.WritePNGFile
  -> feishu reply
```

## What Is Solid

- `internal/clipboard` is cleanly isolated and can become an action adapter.
- `internal/service` is isolated from agent logic and can remain mostly unchanged.
- `internal/core/rules.go` has deterministic text command parsing and validation.
- Feishu resource saving has safe filename and unique path behavior.
- Reply failures are best-effort and do not roll back local actions.
- LaunchAgent has UTF-8 environment and no embedded secrets.
- Node prototype contains a good planner/executor concept that should be ported to Go.

## What Blocks Pluggability

- `MessageHandler.Handle` both normalizes events and executes business actions.
- Event model is transport-shaped (`TextMessage`, `ResourceMessage`) instead of agent-shaped.
- Config is flat and Feishu-specific.
- HTTP `serve` mode is a small debug endpoint, not a versioned agent API.
- There is no runtime registry for transports or actions.
- Actor policy and dedupe are not first-class in Go runtime.
- Reply and resource download interfaces are local to Feishu instead of runtime capabilities.

## Recommended Refactor Boundary

Create `go-copyagentd/internal/agent` as the central package:

- `types.go`: `Event`, `Actor`, `ResourceRef`, `ReplyTarget`, `Plan`, `Action`, `ActionResult`.
- `policy.go`: actor allowlist, dedupe, max text size checks.
- `planner.go`: text/image/file event to action list.
- `executor.go`: action execution with injected clipboard, file store, image clipboard, reply router, resource fetcher.
- `runtime.go`: `HandleEvent(ctx, event)` orchestration.

Then make Feishu a plugin-like adapter:

```text
internal/transport/feishu
  -> normalize SDK event into agent.Event
  -> register capabilities: reply, resource fetch
  -> sink.HandleEvent(event)
```

## Implementation Rule

Do not add dynamic plugin loading first. Start with compile-time built-in plugin interfaces and registry. Dynamic external plugins can come after the runtime contract is stable.
