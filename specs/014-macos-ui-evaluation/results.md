# Results: macOS UI Foundation Evaluation

## Local Runtime Measurement

Measured on 2026-04-25 Asia/Shanghai with the installed upstream reference app `/Applications/Maccy.app` version `2.6.1`.

| Metric | Value |
|---|---:|
| Process RSS (`ps`) | `56,288 KB` (`~55 MB`) |
| CPU idle | `0.0%` |
| App bundle size | `8.5 MB` |
| Main binary | `3.2 MB` |
| Frameworks folder | `3.0 MB` |
| Resources folder | `2.1 MB` |
| Container data | `8.1 MB` |
| CoreData SQLite + WAL | `~8.0 MB` |

`sample` reported physical footprint around `210 MB`, which includes broader macOS framework/runtime accounting and is not directly comparable to RSS. RSS is the better lightweight comparison metric for our current process budgeting.

## Bundle Size Drivers

Largest bundled files:

- `Contents/MacOS/Maccy`: `3.2 MB`
- `Sparkle.framework` binary: `844 KB`
- `Sparkle Autoupdate`: `644 KB`
- `Assets.car`: `628 KB`
- Sound files: `Write.caf` `144 KB`, `Knock.caf` `108 KB`
- Localized `.strings` files: many small files, collectively meaningful but not dominant

Conclusion: package size is already small. Removing Sparkle, sounds, extra localizations can reduce a few MB, but bundle size is not the main problem.

## Runtime Memory Drivers

Likely high-level contributors:

- SwiftUI/AppKit baseline for status bar app and floating panel.
- CoreData + SQLite page cache.
- History item content retained in memory.
- SwiftUI view graph / AttributeGraph.
- Image/rich clipboard support and preview paths when active.
- Sparkle is bundled but unlikely to dominate idle RSS.

Observed `vmmap` highlights:

- `MALLOC` zones reserve/dirty non-trivial memory.
- SQLite page cache present but small in RSS terms.
- Large mapped system framework regions from SwiftUI/AppKit/CoreData are expected for a native macOS UI app.

## Keep + Slim Strategy

The best path is not heavy deletion. Preserve the mature clipboard UX while slimming high-risk or low-value areas:

### Keep by Default

- Status bar icon and popup.
- Search and keyboard navigation.
- Settings framework.
- Keyboard shortcuts.
- Pins and paste stack.
- Preview/slideout source initially, even if not modified.
- Rich pasteboard type support.
- CoreData storage initially.
- Ignore/privacy filters.
- Appearance options.
- App metadata and optional notifications.

### Disable or Remove First

- Sparkle updater until release/signing pipeline exists.
- App Store review prompts.
- Update appcast packaging.
- Any App Store-only entitlement that blocks local/open-source distribution.

### Slim Without Feature Loss

- Keep settings framework but add copyagent settings pane instead of replacing it.
- Keep localization files initially; only stop expanding translations for new copyagent strings until release process is clear.
- Keep CoreData initially; cap history count and size to control memory/data growth.
- Add preferences that can lower memory without deleting features:
  - max history items
  - max item bytes
  - disable image preview cache
  - disable rich types if user wants text-only
  - disable sounds/notifications
  - pause full-system monitoring and show only copyagent-delivered records if desired

## Recommended Integration Plan

1. Import the upstream macOS UI source as `copyagent-ui-mac` with minimal rename and build changes.
2. Disable Sparkle/App Store review first, not broad features.
3. Add a `Copyagent` settings pane for daemon status, Feishu config, logs, service controls, and download directory.
4. Keep the mature clipboard-manager UX intact.
5. Add memory budget checks after each change.
6. Only remove a module when measurement shows clear value or it blocks copyagent product direction.

## Decision

The measured upstream app does not support Windows. If we choose this UI base, copyagent UI becomes macOS-first. This is acceptable if the Go daemon remains clean and the UI is treated as a companion app rather than part of the daemon core.
