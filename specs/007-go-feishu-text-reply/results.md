# Results: Go Feishu Text Reply

## Validation

- `go test ./...` passes.
- `go build -o ./copyagentd ./cmd/copyagentd` passes.
- `./copyagentd doctor` passes with loaded config and available macOS clipboard.
- `./copyagentd feishu-serve` starts and remains alive during a short local smoke test.
- Background `feishu-serve` startup remains alive after shell exit/SIGHUP handling smoke test.
- Real Feishu text message E2E copied `Go E2E 测试成功` to the macOS clipboard.

## Implementation Summary

- Added Go normalization for Feishu `im.message.receive_v1` text payloads.
- Added deterministic message handling: extract copy text, validate non-blank input, write clipboard, then best-effort reply.
- Added `allowedActorIds` enforcement before clipboard writes.
- Added `copyagentd feishu-serve` while leaving the existing Node prototype untouched.

## Memory

Measured on 2026-04-25 Asia/Shanghai during a short `./copyagentd feishu-serve` smoke test:

| Scenario | RSS |
|---|---:|
| Go Feishu text handler long connection | `17,568 KB` (`~17.2 MB`) |

This remains below the `20 MB RSS` acceptance target. The measurement used real local config with secrets redacted from diagnostics.

## E2E Notes

- Clipboard path confirmed with a real Feishu message: `复制：Go E2E 测试成功` resulted in clipboard text `Go E2E 测试成功`.
- Bot reply confirmation still needs user-side confirmation because local diagnostics cannot read the Feishu chat.
- Earlier ad-hoc background launches could exit after shell/session cleanup; `feishu-serve` now ignores `SIGHUP` and stayed alive in an 8-second background smoke test.
- A real post-fix Feishu message should still be sent once to verify copy + reply + continued process survival together.
