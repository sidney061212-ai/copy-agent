# Feature Specification: LaunchAgent UTF-8 Clipboard

**Created**: 2026-04-25  
**Status**: Draft

## Goal

Ensure the macOS LaunchAgent daemon can copy Chinese and other non-ASCII text from Feishu into the system clipboard.

## Scope

- Set a UTF-8 locale for the LaunchAgent environment.
- Run macOS clipboard commands with an explicit UTF-8 environment.
- Keep secrets out of the plist and logs.
- Add post-write clipboard verification logs that report only byte counts.

## Non-Goals

- Changing Feishu credentials.
- Adding new runtime dependencies.
- Rewriting clipboard handling with AppKit in this slice.
