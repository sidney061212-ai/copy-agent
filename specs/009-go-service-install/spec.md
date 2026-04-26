# Feature Specification: Go Service Install

**Feature Branch**: `009-go-service-install`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Make the Go daemon easy to run persistently on macOS without bloating the daemon or putting secrets into launchd configuration.

## User Story

As a user, I can install, start, stop, and inspect the Go copyagent daemon as a macOS launchd service while keeping credentials in `~/.copyagent/config.json`.

## Scope

- Add macOS launchd plist generation for `copyagentd feishu-serve`.
- Store only executable path and log paths in the plist; do not store tokens or Feishu secrets.
- Add minimal CLI commands for install, uninstall, start, stop, and status.
- Use user LaunchAgents, not a privileged system daemon.
- Keep Node prototype service management untouched.

## Non-Goals

- Windows service installation.
- Auto-updaters or packaging installers.
- Complex health monitors.
- Temporary test commands.

## Acceptance Criteria

1. `copyagentd service install` writes a LaunchAgents plist with no secrets.
2. `copyagentd service start` starts the service via `launchctl`.
3. `copyagentd service stop` stops the service via `launchctl`.
4. `copyagentd service uninstall` removes the service and plist.
5. `copyagentd service status` reports whether launchd has the service loaded/running.
6. `go test ./...` and `go build -o ./copyagentd ./cmd/copyagentd` pass.
