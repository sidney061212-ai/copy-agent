# Feature Specification: Go Feishu Transport Spike

**Feature Branch**: `006-go-feishu-transport`  
**Created**: 2026-04-25  
**Status**: Spike

## Goal

Determine whether the production Go daemon can support Feishu/Lark long-connection bot events with low memory.

## Questions

1. Does the official Go SDK support long connection?
2. What is the RSS impact of importing/initializing the Go SDK long-connection client?
3. Can we preserve the <20 MB target with official SDK, or do we need a minimal custom transport?

## Scope

- Add optional Go Feishu transport spike files.
- Measure memory for baseline Go daemon vs Go SDK import/client.
- Do not replace Node Feishu runtime yet.
