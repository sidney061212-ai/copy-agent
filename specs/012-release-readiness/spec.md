# Feature Specification: Release Readiness

**Feature Branch**: `012-release-readiness`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Prepare the Go daemon work for a small open-source release by documenting what works, how to run it, and what remains intentionally out of scope.

## User Story

As a first-time user or maintainer, I can understand the Go daemon's setup, capabilities, safety model, and known gaps without reading handoff notes.

## Scope

- Update Go daemon README with quickstart commands.
- Document current capability matrix for text, image, file, service management, and diagnostics.
- Document security model and secret handling.
- Document known gaps and next steps.
- Avoid adding new runtime features.

## Non-Goals

- Installing service automatically.
- Publishing a package or release artifact.
- Adding new transports or smart/LLM behavior.

## Acceptance Criteria

1. README explains build, doctor, foreground run, service commands, and logs.
2. README includes a capability matrix.
3. README includes security notes and known gaps.
4. Handoff points to release-readiness docs.
5. `go test ./...` and `go build -o ./copyagentd ./cmd/copyagentd` pass after docs changes.
