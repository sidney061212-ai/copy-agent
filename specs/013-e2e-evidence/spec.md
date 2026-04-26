# Feature Specification: E2E Evidence

**Feature Branch**: `013-e2e-evidence`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Collect release evidence for the current Go daemon without adding new user-facing functionality.

## Scope

- Validate launchd install/start/status/logs/stop/uninstall locally.
- Capture RSS snapshots for foreground or launchd-run daemon.
- Confirm service plist does not contain secrets.
- Update docs with evidence and remaining manual E2E gaps.

## Non-Goals

- New product capabilities.
- New transports.
- LLM/smart intent.
- Permanent local service installation unless cleaned up in the same pass.

## Acceptance Criteria

1. Service install creates a plist with no secrets.
2. Service can start, report status, expose logs, stop, and uninstall.
3. No launchd plist or daemon process is left behind after validation.
4. RSS remains under `20 MB` or variance is documented.
5. `go test ./...` and `go build -o ./copyagentd ./cmd/copyagentd` pass.
