# Feature Specification: Transport Plugin Contract

**Feature Branch**: `010-transport-plugin-contract`  
**Created**: 2026-04-25  
**Status**: Draft

## Goal

Stabilize the smallest Go event/action contract so Feishu is no longer the only shape the deterministic handler understands.

## User Story

As a maintainer, I can add another mobile transport later, such as Telegram polling, without duplicating copy/save rules or tying core behavior to Feishu SDK event structs.

## Scope

- Add a small internal normalized event model for text and resource messages.
- Keep transport-specific Feishu normalization at the adapter edge.
- Keep the existing behavior and CLI unchanged.
- Avoid introducing a plugin loader or dynamic module system in this slice.

## Non-Goals

- Runtime plugin loading.
- Telegram implementation.
- New user-facing commands.
- LLM/smart intent routing.

## Acceptance Criteria

1. Text handling can operate on a normalized text event independent of Feishu SDK types.
2. Resource handling can operate on a normalized resource event independent of Feishu SDK types.
3. Feishu adapter remains the only package importing Feishu SDK event structs.
4. Existing text and media tests pass.
5. `go test ./...` and `go build -o ./copyagentd ./cmd/copyagentd` pass.
