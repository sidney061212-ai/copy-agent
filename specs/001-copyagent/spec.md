# Feature Specification: copyagent

**Feature Branch**: `001-copyagent`  
**Created**: 2026-04-24  
**Status**: Draft  
**Input**: User wants a lightweight personal agent named copyagent. Messages sent from Feishu or other chat software to a bot should be copied to the Mac clipboard and remain compatible with standard macOS clipboard observers.

## User Scenarios & Testing

### Primary User Story

As the owner of this Mac, I send text to a chat bot from Feishu or another chat app, and copyagent places that text on my local clipboard so I can immediately paste it anywhere. Any local clipboard history tool can observe the change through normal macOS clipboard monitoring.

### Acceptance Scenarios

1. **Given** copyagent is running locally with a configured token, **When** a valid HTTP request containing text arrives, **Then** the exact text is placed on the macOS clipboard.
2. **Given** a request has no valid token, **When** it reaches copyagent, **Then** copyagent rejects it and does not modify the clipboard.
3. **Given** a Feishu/Lark-style event callback contains message text, **When** it reaches copyagent, **Then** copyagent extracts the text and copies it.
4. **Given** a platform URL verification challenge arrives, **When** it reaches copyagent, **Then** copyagent returns the challenge without touching the clipboard.
5. **Given** text is empty or too large, **When** it reaches copyagent, **Then** copyagent returns a clear validation error and preserves the existing clipboard.

### Edge Cases

- Duplicate delivery retries should not repeatedly rewrite the clipboard when the event id is the same.
- Non-text payloads are ignored with a clear error.
- Missing `pbcopy` on non-macOS systems returns a clear runtime error.
- Bot secrets must come from environment variables or local ignored config, never source code.

## Requirements

### Functional Requirements

- **FR-001**: Provide a local HTTP server with `/health`, `/copy`, and `/feishu` endpoints.
- **FR-002**: `/copy` accepts JSON `{ "text": "..." }` and copies the text to the clipboard.
- **FR-003**: `/feishu` accepts Feishu/Lark URL verification and message callback payloads.
- **FR-004**: All mutating endpoints require a shared token via `Authorization: Bearer`, `X-Copyagent-Token`, or `?token=`.
- **FR-005**: The clipboard write uses macOS `pbcopy`, allowing standard clipboard observers to see clipboard changes naturally.
- **FR-006**: Provide a CLI for starting the server and copying direct stdin/argument text.
- **FR-007**: Provide launchd plist generation/install commands for background running.
- **FR-008**: Log operational events without logging copied text by default.
- **FR-009**: Expose a real agent layer that receives normalized platform events, applies policy, runs clipboard actions, and returns platform responses.
- **FR-010**: Support platform adapters for generic webhooks and Feishu/Lark event callbacks.
- **FR-011**: Support optional Feishu/Lark verification token and encrypt-key signature verification using the raw request body.
- **FR-012**: Provide lifecycle commands for foreground server mode and launchd installation/uninstallation.

### Non-Functional Requirements

- **NFR-001**: No LLM dependency and no database dependency.
- **NFR-002**: Runtime dependencies should be minimal; Node built-ins preferred.
- **NFR-003**: Configuration must be environment-variable driven.
- **NFR-004**: The implementation must be easy to inspect and self-host locally.
- **NFR-005**: Agent actions must be deterministic and bounded; no model call is needed for the copy workflow.

## Out of Scope

- Full Feishu app provisioning automation.
- Rich media/file clipboard support.
- Cloud hosting.
- Multi-user account management.
