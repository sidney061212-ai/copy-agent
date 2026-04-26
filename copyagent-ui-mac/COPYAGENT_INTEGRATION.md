# copyagent-ui-mac Internal Integration Notes

This file is an engineering note for maintainers. Public product-facing usage should rely on `README.md`.

This directory includes imported third-party MIT-licensed macOS clipboard UI source, adapted as the macOS UI foundation for copyagent.

Exact attribution is maintained in the repository root `NOTICE.md` and in `LICENSE`.

Imported on: 2026-04-25 Asia/Shanghai

Integration direction:

- Keep the mature status bar, history, search, settings, shortcuts, pins, paste stack, preview, rich pasteboard, CoreData/SwiftData storage, privacy filters, and appearance UX.
- Rename and bind the app to copyagent instead of treating the imported source as an external dependency.
- Keep `copyagentd` as the lightweight Go daemon for Feishu, files, service, logs, and diagnostics.
- Add UI settings/actions that call local `copyagentd` commands first; daemon HTTP/API can come later.
- Prefer small, reversible integration steps over broad rewrites.

Current copyagent identity:

- Bundle ID: `app.copyagent.ui.mac`
- Product name: `copyagent`
- Local storage: `Application Support/copyagent/Storage.sqlite`
- Daemon lookup order in `CopyagentDaemonClient`: repo-local `go-copyagentd/copyagentd`, then `/usr/local/bin/copyagentd`.

Disabled for now:

- Sparkle update runtime/package/feed.
- App Store review prompt.

Do not commit secrets into this UI project.
