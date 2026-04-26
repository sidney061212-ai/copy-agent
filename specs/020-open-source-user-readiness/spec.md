# Feature Specification: Open Source User Readiness

**Created**: 2026-04-25  
**Status**: In Progress

## Goal

Prepare copyagent for an initial open-source release where a macOS user can clone the repository, build the daemon and companion app, configure Feishu credentials locally, install/uninstall the LaunchAgent, and verify the core workflow without relying on developer-specific local state.

## Users

- A macOS user who wants mobile Feishu messages to copy text/images/files to their Mac.
- A developer who wants to inspect, build, test, and contribute to copyagent.

## In Scope

- Replace local handoff-oriented README content with user-facing installation, configuration, verification, and troubleshooting docs.
- Add install/uninstall scripts for the Go daemon LaunchAgent and optional macOS UI build/install.
- Provide safe example configuration files with placeholder secrets only.
- Add root license and notice/attribution material for the third-party-derived macOS UI.
- Fix known UI test baseline failures caused by icon size changes.
- Exclude generated binaries and local-only artifacts from source releases.

## Out of Scope

- Signed/notarized binary distribution.
- Homebrew formula.
- Automated GitHub Actions release pipeline.
- Windows UI.
- New Feishu OAuth/configuration wizard UI.

## Acceptance Criteria

- `README.md` explains prerequisites, install, configure, start, verify, update, uninstall, privacy/security, and troubleshooting without developer-specific absolute paths.
- `go-copyagentd/README.md` documents daemon-specific commands and config using `$HOME` or relative paths.
- `copyagent-ui-mac/README.md` documents build/install from source and required attribution.
- `scripts/install.sh` builds and installs the daemon, writes a config template if missing, installs/starts LaunchAgent, and optionally builds/installs the UI.
- `scripts/uninstall.sh` stops/removes LaunchAgent and optionally removes installed app/config/logs/downloads.
- Root `LICENSE` and `NOTICE.md` exist.
- Generated binaries are ignored and removed from the working tree.
- Validation commands pass or any intentionally skipped UI automation is clearly documented.

## Security Requirements

- No real Feishu secret, token, or encrypt key in docs, scripts, fixtures, or examples.
- Config templates use placeholders only and are written with user-only permissions.
- LaunchAgent plist must not embed Feishu credentials.
- Troubleshooting examples must avoid printing secrets.
