# Results: cc-connect-Style Pluggable Clipboard Agent

Updated: 2026-04-25 21:40 CST

## Summary

Implemented a cc-connect-style bridge for copyagent with two runtime modes:

- Direct Mode: default, deterministic, no AI.
- Agent Mode: experimental, routes natural-language Feishu messages to Codex/Claude CLI agents.

Users can switch modes in Feishu:

- `/agent` enables Agent Mode and persists `agent.enabled=true`.
- `/copy` returns to no-AI Direct Mode and persists `agent.enabled=false`.

## Implemented

- Normalized `agent.Message` and attachment model.
- Transport and coding-agent registries.
- Feishu transport plugin normalization into `agent.Message`.
- cc-connect-style structured Feishu `replyContext` with `messageID`, `chatID`, and `sessionKey`.
- `ReplyCapable`, `ResourceCapable`, and `TypingCapable` transport capabilities.
- Direct Mode policy/planner/executor.
- Session store, busy queue, and resume fallback.
- Codex `exec --json` adapter with thread resume.
- Claude Code stream-json adapter with session resume.
- Agent Mode routing behind runtime mode switcher.
- Feishu typing reaction during Agent turns, including reaction event ignore handlers.

## Live validation

Real Feishu bot validation passed:

- `/agent` switched runtime to Agent Mode.
- Natural-language message routed to Codex and produced a visible Feishu reply.
- `复制 copyagent-direct-ok` hit Direct fast-path and did not route to Codex.
- `/copy` switched runtime back to Direct Mode.
- Agent typing reaction appeared while processing and was removed after completion.

## Critical lessons

- Do not invent Feishu reply architecture. Use cc-connect-style structured `ReplyCtx`; bare `messageID` is insufficient.
- Typing reaction must register reaction created/deleted ignore handlers, matching cc-connect.
- Default must remain Direct Mode; Agent Mode is experimental and user-controlled.
- Agent-related future work should keep referring to `/tmp/cc-connect-src` before implementation.

## Validation

Passed:

```bash
cd go-copyagentd && go test ./...
```

Not run in this final slice:

- `npm test`
- Swift/Xcode build

## Next

1. Discuss and freeze low-token Agent Mode boundary with the user.
2. Design remote托管 behavior: paste/inject user messages into foreground Codex/Claude/programming tools with explicit safety checks.
3. Then implement the narrow `copyagent action ...` CLI/API.
