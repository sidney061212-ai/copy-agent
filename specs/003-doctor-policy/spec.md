# Feature Specification: Doctor/Profile and Core Policy

**Feature Branch**: `003-doctor-policy`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Add operational diagnostics for open-source users and move allowlist/dedup checks toward the pure core/policy layer.

## User Scenarios

1. As a user, I can run `copyagent doctor` to see whether config, launchd, Feishu credentials, download directory, clipboard, and logs are healthy.
2. As a maintainer, I can run `copyagent profile` to see `copyagentd`, `copyagent-ui-mac`, and `cc-connect` process memory without manually crafting shell commands.
3. As a future platform implementer, I can reuse core policy checks for allowlist and dedup without depending on Feishu/macOS runtime code.

## Requirements

- **FR-001**: Add `copyagent doctor` with redacted, human-readable health checks.
- **FR-002**: Add `copyagent profile` with process RSS/CPU summaries for `copyagentd`, `copyagent-ui-mac`, and `cc-connect` when present.
- **FR-003**: Add `src/core/policy.js` for actor allowlist and event dedup primitives.
- **FR-004**: Avoid printing secrets in diagnostic output.
- **FR-005**: Keep current Feishu bot behavior working.
