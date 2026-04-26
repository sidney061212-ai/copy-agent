# Feature Specification: Go Feishu Text Reply

**Feature Branch**: `007-go-feishu-text-reply`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Port the first real production Feishu bot behavior from the Node prototype into the Go daemon while preserving the lightweight daemon target.

## User Story

As a user, I can send a Feishu bot text message such as `复制：内容` or raw text, and copyagentd writes the intended text to the desktop clipboard and replies with a fixed confirmation.

## Scope

- Register the Feishu `im.message.receive_v1` long-connection event handler in Go.
- Normalize text events into a small internal event shape.
- Apply deterministic copy rules only; no LLM or hidden conversation context.
- Write valid text to the macOS clipboard through the existing clipboard adapter.
- Reply to the source Feishu message with `✅ 已复制到剪切板` when replies are enabled.
- Keep the Node prototype unchanged during the Go migration.

## Non-Goals

- Media/image/file handling.
- Launchd/service installation.
- Smart intent extraction.
- Custom Feishu WebSocket protocol implementation.

## Acceptance Criteria

1. Go handler registers `im.message.receive_v1`.
2. Text messages copy the extracted text to clipboard.
3. Bot replies with the fixed success template when copy succeeds and `replyEnabled` is true.
4. Invalid or non-text events do not write to clipboard.
5. Reply failures are best-effort and do not fail the copy action.
6. `go test ./...` passes.
7. RSS after real long-connection start is measured under `20 MB` or documented in `results.md` if not.
