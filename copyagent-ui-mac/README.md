# copyagent-ui-mac

[English](./README.md) | [简体中文](./README.zh-CN.md)

`copyagent-ui-mac` is the optional macOS companion app for copy-agent.

It provides:

- clipboard history
- search and shortcuts
- pins and paste stack
- local desktop UX around clipboard workflows

The daemon logic stays in `copyagentd`. This app does not replace the daemon.

## Build

```bash
cd copyagent-ui-mac
xcodebuild -project Copyagent.xcodeproj -scheme Copyagent -configuration Debug -destination 'platform=macOS,arch=arm64' build
```

Release-style local build:

```bash
xcodebuild -project Copyagent.xcodeproj -scheme Copyagent -configuration Release -destination 'platform=macOS,arch=arm64' build
```

## Install

From the repository root:

```bash
scripts/install.sh --with-ui
open ~/Applications/copyagent.app
```

## Tests

```bash
xcodebuild test -project Copyagent.xcodeproj -scheme Copyagent -destination 'platform=macOS,arch=arm64' -only-testing:CopyagentTests
```

`CopyagentUITests` require an interactive desktop session and the relevant permissions.

## Design Boundaries

- the UI does not run Feishu/Lark long connections
- the UI should remain lightweight
- clipboard-history UX should stay a strength, not an afterthought

## Attribution

This app includes third-party licensed code. See `LICENSE` and the repository root `NOTICE.md` for attribution details.
