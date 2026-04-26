# Feature Specification: macOS UI Integration

**Feature Branch**: `015-macos-ui-integration`
**Created**: 2026-04-25
**Status**: Draft

## Goal

Integrate the imported macOS UI foundation into copyagent as a deeply bound companion app, not as an external dependency or untouched clone.

## Product Direction

The combined product is `copyagent`:

- `copyagentd`: lightweight Go daemon for transports, clipboard writes, file saves, service, logs, and diagnostics.
- `copyagent-ui-mac`: macOS status bar/history/settings app derived from imported third-party clipboard UI source, renamed and integrated with copyagent.

Windows UI is deferred. The Go daemon should remain clean enough for future cross-platform work.

## Scope

- Import the upstream UI source into the repo under `copyagent-ui-mac/`.
- Rename visible project/app identity toward copyagent.
- Keep the mature core UI capabilities instead of broad deletion.
- Disable or remove only release-channel modules that conflict with copyagent's current distribution path.
- Add a copyagent integration layer that can call local `copyagentd` commands for doctor/status/logs.
- Document build status and remaining binding work.

## Non-Goals

- Full visual redesign.
- Rewriting the imported UI in Go or Tauri.
- Windows UI support in this slice.
- Removing mature UI features without measurement or product reason.

## Acceptance Criteria

1. `copyagent-ui-mac/` exists in the repo with imported third-party macOS UI source.
2. The imported app is renamed toward copyagent in README/build-facing docs.
3. Sparkle/App Store review paths are identified and disabled or documented if not changed.
4. A copyagent integration layer exists in source and can run `copyagentd doctor` / `service status` style commands.
5. No copyagent secrets are added to UI source.
6. Build attempt is performed or blockers are documented.
