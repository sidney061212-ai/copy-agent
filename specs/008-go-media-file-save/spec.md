# Feature Specification: Go Media/File Save

**Feature Branch**: `008-go-media-file-save`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Port the smallest useful Feishu media behavior to the Go daemon without increasing daemon complexity or memory footprint beyond the lightweight target.

## User Story

As a user, I can send an image or file to the Feishu bot and copyagentd saves it to my configured download directory, then replies with a deterministic success message.

## Scope

- Normalize Feishu `image` and `file` messages from `im.message.receive_v1`.
- Enforce `allowedActorIds` before downloading or writing files.
- Download message resources with the official Go SDK/API.
- Save resources under `defaultDownloadDir`.
- Use safe basenames and avoid overwriting existing files.
- Reply with a fixed success message when `replyEnabled` is true.
- Keep image clipboard copy out of this first Go media slice to avoid format-specific AppleScript complexity.

## Non-Goals

- Image-to-clipboard support.
- User-selectable save commands such as `保存到：桌面`.
- Media thumbnails, OCR, previews, or smart routing.
- Temporary E2E/debug CLI commands.

## Acceptance Criteria

1. Text behavior from 007 continues to work.
2. Image/file events produce a normalized resource with message id, key, kind, and safe file name.
3. Missing resource keys do not download or write files.
4. Downloads are written to `defaultDownloadDir`, creating it if needed.
5. Existing files are not overwritten; collisions receive a numeric suffix.
6. Reply failure remains best-effort.
7. `go test ./...` and `go build -o ./copyagentd ./cmd/copyagentd` pass.
8. RSS remains below `20 MB` in idle long-connection smoke tests or the variance is documented.
