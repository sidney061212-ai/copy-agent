# Embedding Notes

[English](./EMBEDDING.md) | [简体中文](./EMBEDDING.zh-CN.md)

copy-agent is being organized so its core ideas can be embedded into other local-first products.

## What Should Be Reusable

- normalized inbound messages
- deterministic action planning
- local action execution
- clear transport boundaries
- clear permission boundaries

## What Should Stay at the Edge

- Feishu/Lark SDK specifics
- macOS LaunchAgent specifics
- UI-specific clipboard history behavior
- target-app automation details

## Practical Embedding Direction

If another product wants copy-agent-style behavior, the reusable shape should look like:

```text
incoming message
  -> normalized event
  -> deterministic policy
  -> action plan
  -> local executor
  -> optional reply
```

The product should not need to embed every transport or every UI layer just to reuse the core workflow.

## Current Status

Today, the production runtime is still centered on `copyagentd`.

The embedding direction is architectural guidance, not a separate stable SDK release yet.
