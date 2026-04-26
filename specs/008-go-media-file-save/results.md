# Results: Go Media/File Save

## Validation

- `go test ./...` passes.
- `go build -o ./copyagentd ./cmd/copyagentd` passes.
- `./copyagentd feishu-serve` starts with media path wired.

## Implementation Summary

- Added Go normalization for Feishu `image` and `file` messages.
- Added resource key validation before downloads.
- Added safe basename handling and numeric collision suffixes.
- Added save-only media handling under `defaultDownloadDir`.
- Added Feishu message resource downloader using the official Go SDK.
- Kept image clipboard copy out of this slice to avoid AppleScript/format complexity.

## Memory

Measured on 2026-04-25 Asia/Shanghai during a short `./copyagentd feishu-serve` smoke test:

| Scenario | RSS |
|---|---:|
| Go Feishu text + media-save handler long connection | `17,440 KB` (`~17.0 MB`) |

This remains below the `20 MB RSS` target.

## Remaining E2E

- Send a real Feishu image and file to verify resource download permissions and saved filenames.
- Confirm reply behavior in Feishu chat from the user side or via a future non-production diagnostic flow.
