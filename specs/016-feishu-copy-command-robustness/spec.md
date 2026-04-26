# Feature Specification: Feishu Copy Command Robustness

**Created**: 2026-04-25  
**Status**: Draft

## Goal

Make Feishu text copy handling robust for natural command formats users send from mobile chat.

## Scope

- Accept copy commands with either colon separators or whitespace separators.
- Preserve existing behavior for `复制：内容`, `拷贝：内容`, `copy: content`, and `cp: content`.
- Add safe diagnostics that explain why a text message was skipped without logging secrets or full message content.
- Keep the Go daemon lightweight and dependency-free.

## Non-Goals

- Adding LLM intent extraction.
- Adding new runtime dependencies.
- Changing Feishu app credentials or LaunchAgent secret handling.
- Modifying the macOS UI.

## Acceptance Criteria

1. `复制 hello`, `复制：hello`, `copy hello`, and `copy: hello` all extract `hello`.
2. A bare command such as `复制` is rejected as blank instead of copying the command word.
3. Non-command text remains copyable as plain text for backward compatibility.
4. Handler logs safely identify parse/blank failures without printing the full message body.
5. Go tests pass.
