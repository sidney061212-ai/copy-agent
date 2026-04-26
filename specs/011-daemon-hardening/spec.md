# Feature Specification: Daemon Hardening

**Feature Branch**: `011-daemon-hardening`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Harden the Go daemon for day-to-day use before adding new product complexity.

## User Story

As a user, I can inspect daemon health and logs without exposing secrets or manually hunting launchd files.

## Scope

- Add config file permission diagnostics.
- Add download directory diagnostics.
- Add service log path discovery and a minimal log tail command.
- Keep checks read-only unless the user runs explicit install/start commands.

## Non-Goals

- Log rotation daemon.
- Real service installation without user confirmation.
- Smart/LLM intent extraction.
- Temporary E2E commands.

## Acceptance Criteria

1. `copyagentd doctor` reports config file permission status.
2. `copyagentd doctor` reports whether `defaultDownloadDir` is writable or creatable.
3. `copyagentd service logs` prints recent service logs from known launchd log paths.
4. Missing logs are handled with a friendly message.
5. `go test ./...` and `go build -o ./copyagentd ./cmd/copyagentd` pass.
