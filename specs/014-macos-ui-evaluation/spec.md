# Feature Specification: macOS UI Foundation Evaluation

**Feature Branch**: `014-macos-ui-evaluation`
**Created**: 2026-04-25
**Status**: Analysis

## Goal

Decide whether copyagent should depend on, import/refactor, or reimplement a lightweight tray/history UI foundation while preserving cross-platform and low-memory goals.

## Constraints

- Status bar / tray UI is required.
- Avoid Swift-only architecture because Windows support matters.
- Preserve Go daemon lightweight principle.
- Do not vendor large UI code before the integration path is chosen.

## Findings Summary

- The candidate upstream UI is MIT licensed.
- The candidate upstream UI is a native macOS app built heavily on SwiftUI, AppKit, NSPasteboard, CoreData, NSStatusItem, NSPanel, Carbon, and macOS-specific packages.
- The candidate upstream UI includes over 100 Swift files, ~29 SwiftUI view files, many settings/localization files, and CoreData models.
- Direct dependency/fork would pull in macOS-only architecture and conflict with cross-platform goals.
- The valuable product patterns are: tray icon, searchable recent history, privacy filters, pinned items, keyboard-first popup.
- The reusable code is mostly conceptual, not directly portable to Go/Windows.

## Recommendation

Do not treat the upstream UI as a direct live dependency or untouched fork. Build a copyagent-owned tray/history layer inspired by the imported product patterns.

Preferred implementation direction:

1. Keep Go daemon as the source of truth for transport, clipboard, file save, and history records.
2. Add a minimal history API/store in Go only after agreeing on privacy defaults.
3. Add a separate cross-platform tray UI process that talks to the daemon over local HTTP or a small local socket.
4. Evaluate Tauri or Wails for tray UI; avoid SwiftUI as the main UI path.

## Decision Needed

Choose the UI stack before implementing:

- Wails: Go-native backend + web UI, good fit with current daemon language, but v3 tray docs are alpha/current.
- Tauri: mature cross-platform tray/window story, Rust shell + web UI, but introduces Rust alongside Go.
- Go systray: very light tray menu, but not enough for searchable popup UI by itself.
