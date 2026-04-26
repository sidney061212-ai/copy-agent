# Foreground Injection MVP Design

Status: design accepted for the next implementation slice.
Date: 2026-04-25.
Reference: `/tmp/cc-connect-src` remains the primary architecture reference. This design extends copyagent's existing cc-connect-style transport/engine/action boundary; it does not introduce a separate agent runtime.

## Current Implementation Note

This document started as the initial MVP design. The implemented behavior has since evolved during live Feishu testing. For the latest public-facing state, read the repository `README.md` first and use `docs/DEVELOPMENT.md` for durable architecture history.

Current behavior differs from the initial MVP in important ways:

- Feishu `/inject <task>` now paste-submits by default: Cmd+V followed by Return.
- Local CLI `action inject-text` remains paste-only unless `--submit` is passed.
- Feishu `/turn <name>` now activates and binds a named target app before injection.
- `/inject` can activate the bound target and verify the frontmost bundle before pasting.
- The injected protocol requires three `reply-text`回传 messages: fixed receipt confirmation, user-task plan, and final conclusion.
- The foreground app should retain the final conclusion for local review.
- Generic click-to-focus was tried and rolled back; do not reintroduce it without a new per-target design.
- copyagent does not OCR/read UI output and does not convert remote hosting into a background Codex/Claude CLI subprocess.
- Explicit foreground-hosting commands are now routed before normal Direct-vs-Agent message dispatch, so `/turn` and `/inject` are deterministic commands available regardless of the current copy/agent mode.

## Goal

Make Agent Mode remotely usable before deeper Agent prompt/rule redesign:

```text
Feishu explicit command -> copyagent engine -> local foreground-app injection -> Feishu visible result
```

The first usable behavior is intentionally narrow: paste or paste-submit a Feishu message into the selected foreground coding tool. Do not make ordinary Agent Mode messages inject by default.

## Non-goals for the MVP

- No automatic injection for every Agent Mode message.
- No automatic Enter/Return submit by default.
- No arbitrary shell execution.
- No broad UI automation framework.
- No deep Agent rule redesign yet.
- No attempt to infer the right target app if the frontmost app is not allowed.

## User-facing command shape

Start with one explicit Feishu command handled before normal Agent routing:

```text
/inject <text>
```

Optional aliases can be added later, but the first implementation should keep parsing boring and exact.

Semantics:

- Works only in Agent Mode initially, unless the user explicitly chooses otherwise later.
- Removes the command prefix and injects the remaining text.
- Rejects empty text.
- Does not route the text to Codex/Claude CLI agent.
- Uses the same allowlist/policy path as Direct Mode before performing local side effects.
- Replies visibly in Feishu with success or a precise failure reason.

Possible replies:

- Success: `已提交到前台应用：<app name>，等待回程回复。`
- Empty: `请在 /inject 后提供要粘贴的文本。`
- Wrong app: `未粘贴：当前前台应用 <app name> 不在允许列表。请发送 /turn codex、/turn claude、/turn terminal、/turn iterm、/turn warp、/turn vscode 或 /turn cursor 后重试。`
- Permission missing: `未粘贴：缺少辅助功能或自动化权限，请在 macOS 设置中授权 copyagentd。`
- Clipboard restore warning: `已粘贴，但恢复原剪切板失败：<reason>`

## Local CLI/API shape

Add a narrow local action command so Feishu and future Agent rules share the same side-effect path:

```bash
copyagentd action inject-text [--submit] --text "..."
copyagentd action inject-text [--submit] --stdin
copyagentd action reply-text --session-key <key> --text "..."
copyagentd action reply-text --session-key <key> --stdin
copyagentd action turn status|codex|claude|terminal|iterm|warp|vscode|cursor
copyagentd action status
```

For the first code slice, the Feishu handler may call the same Go package directly instead of spawning the binary. The CLI exists so future Codex/Claude prompt rules can request the same operation without receiving broad control.

## Allowed targets

MVP allowed frontmost targets should be conservative and config-driven later. Hardcoded first-pass candidates are acceptable if documented and easy to change:

- Terminal-style apps: Terminal, iTerm2, Warp.
- IDE/editors that may host Codex/Claude terminals: VS Code, Cursor, JetBrains IDEs.
- Dedicated coding agent apps if detectable by bundle ID/title.

Validation should use frontmost application bundle ID and localized name. Window title is optional supporting evidence, not the only control.

Initial bundle allowlist examples:

```text
com.openai.codex
com.anthropic.claude
com.apple.Terminal
com.googlecode.iterm2
dev.warp.Warp-Stable
com.microsoft.VSCode
com.todesktop.230313mzl4w4u92
com.jetbrains.intellij
com.jetbrains.pycharm
com.jetbrains.goland
```

The Cursor bundle ID has varied across builds; detect and log the real value during status/probe before relying on it.

## macOS injection mechanism

Use the simplest reliable macOS path first:

1. Inspect the frontmost app.
2. Validate it against the allowlist.
3. Save current pasteboard content where possible.
4. Write the requested text to the pasteboard.
5. Send Cmd+V to the frontmost app through AppleScript/System Events or Accessibility.
6. Optionally wait a short bounded delay.
7. Restore the previous pasteboard content where possible.
8. Return structured result to the caller.

Historical note: the initial design avoided Return. Live Feishu testing showed paste-only was not enough for remote hosting, so Feishu `/inject` now submits with Return. The local CLI still requires explicit `--submit`.

## Package boundary

Proposed Go package:

```text
go-copyagentd/internal/inject/
  injector.go       core interface and result types
  macos.go          darwin implementation
  macos_stub.go     non-darwin unsupported implementation
  injector_test.go  parser/validation/unit tests using fakes
```

Core types:

```go
type Request struct {
    Text string
    Submit bool
    TargetBundleID string
    AllowedBundleIDs []string
    MaxBytes int
}

type Target struct {
    AppName string
    BundleID string
    WindowTitle string
}

type Result struct {
    Target Target
    RestoredClipboard bool
    Warning string
}
```

Interfaces for tests:

```go
type FrontmostInspector interface { Frontmost(context.Context) (Target, error) }
type Activator interface { ActivateBundle(context.Context, string) error }
type Pasteboard interface { Snapshot(context.Context) (Snapshot, error); WriteText(context.Context, string) error; Restore(context.Context, Snapshot) error }
type Keystroker interface { Paste(context.Context) error; Submit(context.Context) error }
```

The concrete darwin implementation can use `osascript`/AppleScript first because it is dependency-free. If reliability is poor, replace internals later with Accessibility APIs without changing the action boundary.

## Feishu routing

Add the `/inject` branch before Agent Mode CLI routing, similar to Direct fast-path:

```text
AgentModeHandler.HandleMessage
  -> shouldDirect(copy/images/files)
  -> shouldTurn(/turn ...)
  -> shouldInject(/inject ...)
  -> policy.Allow
  -> inject.Executor
  -> transport.Reply
  -> return, never send to CodingAgent
```

This mirrors cc-connect's principle that the engine owns routing and side effects while agent subprocess details stay hidden.

## Safety checks

MVP checks required before paste:

- User must pass existing actor allowlist policy.
- Text must be non-empty and valid UTF-8 after trimming command prefix.
- Text length must be capped, e.g. 16 KiB initially.
- Frontmost bundle ID must be allowed.
- Pasteboard must not be left containing injected text if restore is possible.
- Operation must return visible success/failure through Feishu.
- Logs must include byte counts and target metadata, not full user text.

## Config direction

Do not block MVP on config plumbing, but design for this shape:

```json
{
  "agent": {
    "inject": {
      "enabled": true,
      "requireAgentMode": true,
      "submitDefault": false,
      "maxBytes": 16384,
      "allowedBundleIDs": ["com.apple.Terminal", "com.microsoft.VSCode"]
    }
  }
}
```

If config is deferred, keep constants in one package and document them.

## Test plan

Unit tests:

- `/inject` parser rejects empty text and accepts multiline text.
- Agent Mode routes `/inject` to injection executor and not to `CodingAgent`.
- Direct fast-path still wins for `copy`/`复制`, images, and files.
- Wrong user policy blocks injection.
- Wrong bundle ID returns a failure reply.
- Pasteboard restore failure returns success with warning, not silent failure.

Manual validation:

1. Start copyagentd foreground daemon.
2. Put Terminal/Codex/Claude/IDE in foreground with an active input box.
3. Send `/agent` in Feishu.
4. Send `/turn codex` or another named target.
5. Send `/inject hello from feishu`.
6. Confirm text appears in foreground input and Feishu replies success.
7. Put an unsupported app in foreground and retry; confirm no paste occurs and Feishu reports wrong app.
8. Confirm ordinary Agent Mode messages still route to Codex/Claude and `/copy` still returns Direct Mode.

## Later evolution

Expected future changes are allowed and anticipated:

- Deeply revise Agent prompt/rules to decide when to call `copyagentd action inject-text`.
- Add a separate explicit托管 mode where ordinary messages inject by default.
- Add `--submit` with stronger confirmation.
- Add per-session authorization windows, e.g. `/inject-auth 10m`.
- Add UI status/menu controls in the macOS app.
- Replace AppleScript internals with Accessibility-native code if needed.
