# Feature Specification: Core/Adapter Refactor

**Feature Branch**: `002-core-adapter-refactor`  
**Created**: 2026-04-25  
**Status**: Draft  
**Input**: copyagent should become an open-source, pluggable, cross-platform agent. Current Feishu/macOS prototype works, but code must be refactored before adding LLM, Windows, Telegram, WeChat, Claude/Codex connectors, or file modification features.

## Goal

Separate copyagent into a pure, side-effect-free core and platform/transport/action adapters while preserving current behavior.

## User Scenarios & Testing

### Primary User Story

As a project maintainer, I want copyagent's core rules and action planning to be independent from Feishu, macOS, launchd, and Node-specific side effects, so that future Windows/mobile-platform/LLM integrations can be added without rewriting product logic.

### Acceptance Scenarios

1. **Given** a normalized text event, **When** core handles it, **Then** it returns a deterministic action plan to copy text and reply.
2. **Given** a normalized image event, **When** core handles it, **Then** it returns an action plan to download/save/copy image and reply.
3. **Given** a normalized file event, **When** core handles it, **Then** it returns an action plan to download/save file and reply.
4. **Given** the current Feishu bot runtime, **When** text/image/file messages arrive, **Then** behavior remains compatible with the pre-refactor implementation.
5. **Given** tests run, **When** `npm test` completes, **Then** all current and new tests pass.

## Requirements

- **FR-001**: Create `src/core/` modules with no imports from Feishu SDK, clipboard, filesystem, launchd, or HTTP server code.
- **FR-002**: Define normalized event shape and action plan shape.
- **FR-003**: Move command parsing and reply planning into core-compatible modules.
- **FR-004**: Add an executor layer that maps action plans to adapters.
- **FR-005**: Preserve existing CLI commands and launchd behavior.
- **FR-006**: Keep no-LLM behavior as the default.
- **FR-007**: Document the new architecture for future contributors.

## Non-Goals

- No new LLM integration in this phase.
- No new mobile transport in this phase.
- No Go rewrite in this phase.
- No Windows implementation yet, but adapter boundaries must not block it.
