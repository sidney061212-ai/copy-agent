# Feature Specification: Memory Audit and Runtime Slimming Decision

**Feature Branch**: `004-memory-audit`  
**Created**: 2026-04-25  
**Status**: Draft

## Problem

copyagent's RSS has grown from ~39 MB to ~100 MB while cc-connect Go binary instances are ~16 MB. Since copyagent is intended to be an open-source lightweight always-on agent, high idle memory threatens product viability.

## Goal

Determine whether high memory comes from our code, Node baseline, Feishu SDK, runtime structure, or leaks. Produce an evidence-backed decision: optimize Node prototype, replace Feishu SDK, or rewrite daemon in Go.

## Requirements

- Measure Node baseline memory.
- Measure Node + Feishu SDK import memory.
- Measure current copyagent memory after fresh restart and short idle.
- Compare with cc-connect and the optional macOS UI companion process.
- Inspect whether duplicate WebSocket clients/processes are created.
- Document findings and recommendation.

## Findings

See `docs/internal/MEMORY_AUDIT.md` for measurements and decision notes.
