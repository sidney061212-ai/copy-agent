# Feature Specification: Restore Image Clipboard

**Created**: 2026-04-25  
**Status**: Draft

## Goal

Restore the image clipboard behavior that existed in the Node prototype after the production daemon moved to Go.

## Scope

- Keep Feishu image/file resource saving in Go.
- Respect `imageAction` from config:
  - `clipboard` or empty: save image, then copy PNG image to macOS clipboard.
  - `save`: save image only.
- Keep file events save-only.
- Use built-in macOS tools only; do not add dependencies.
- Keep the feature behind platform-specific clipboard code.

## Non-Goals

- OCR, previews, or smart image routing.
- Windows image clipboard support in this slice.
- Reintroducing the Node daemon.
