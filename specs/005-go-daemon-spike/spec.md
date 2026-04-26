# Feature Specification: Go Daemon Spike

**Feature Branch**: `005-go-daemon-spike`  
**Created**: 2026-04-25  
**Status**: Spike

## Goal

Validate whether a Go-based copyagent daemon can deliver the core lightweight value proposition with idle RSS under 20 MB.

## Scope

This spike does not replace the Node prototype. It creates a separate Go prototype under `go-copyagentd/`.

## Minimum Capability

- Load existing `~/.copyagent/config.json` enough to reuse core settings.
- Provide a `copyagentd doctor` command.
- Provide a `copyagentd copy <text>` command using platform clipboard adapters.
- Provide a local HTTP `/copy` endpoint for testability.
- Measure idle memory.

## Out of Scope For Spike

- Full Feishu long connection.
- Image/file download.
- Telegram/WeChat.
- LLM integration.
- Replacing Node launchd service.

## Success Criteria

- Go binary builds and tests pass.
- `copyagentd copy '中文'` works on macOS.
- Local HTTP `/copy` writes clipboard.
- Idle RSS is below 20 MB.
- Migration recommendation is documented.
