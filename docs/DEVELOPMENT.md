# Development Guide

This file is the durable project memory for copy-agent.

## Iron Rule

Every phase that changes architecture, behavior, release scope, or documentation shape must update `docs/DEVELOPMENT.md` before the phase is considered complete.

## Project Definition

copy-agent is a local-first automation bridge.

Its current product core is:

- `copyagentd` as the always-on execution daemon
- `copyagent-ui-mac` as an optional macOS companion
- Feishu/Lark as the current primary remote entrypoint

The product is not defined by foreground desktop control alone. Foreground-hosting is an explicit experimental capability, not the whole identity of the project.

## Public Release Baseline

As of 2026-04-27:

- **Stable**: Direct Mode clipboard, image, file, reply, diagnostics, and service workflows
- **Experimental**: Agent Mode
- **Experimental**: foreground-hosting flows such as `/turn`, `/inject`, and `reply-text`

## Architecture Decisions

### 1. Go is the production daemon runtime

The project keeps the Go daemon as the production direction for:

- lower idle resource usage
- simpler long-running service behavior
- better fit for LaunchAgent and CLI-facing runtime control

The legacy Node implementation remains in the repository for tests, historical reference, and transition safety. It is not the primary runtime direction.

### 2. The UI is optional

`copyagent-ui-mac` is a companion app, not the primary backend.

It should provide local clipboard UX and settings surfaces, while `copyagentd` remains responsible for transports, deterministic actions, diagnostics, and service lifecycle.

### 3. Direct Mode remains the stable baseline

The first public release must keep a narrow promise:

- predictable deterministic behavior
- local-first data handling
- explicit distinction between stable and experimental capabilities

### 4. Foreground hosting is a separate capability layer

Foreground hosting depends on:

- app targeting
- pasteboard injection
- submit keystrokes
- macOS Accessibility and Automation permissions

This layer is useful, but it is intentionally treated as experimental.

### 5. Reference implementations stay in engineering docs only

External projects can inform architecture decisions, but they must not define the public product narrative.

Engineering references belong in internal or development-facing documents, not in public marketing or setup docs.

## Documentation Contract

### Public docs

Public, user-facing documentation should look like a real shipped product:

- clean scope statements
- no chat-session residue
- no internal workflow assumptions
- no “reference implementation” framing in the product narrative

### Language shape

Public docs now use GitHub-style language switching:

- English primary file, such as `README.md`
- Simplified Chinese sibling, such as `README.zh-CN.md`
- top-of-file switch links between the two

Inline mixed bilingual blocks are no longer the preferred public format.

### Internal docs

Internal session notes and engineering archives belong under `docs/internal/`.

The root of the repository should stay clean and product-facing.

### Preservation rule

Do not silently erase prior conclusions, constraints, or historical milestones.

If a previous approach is replaced, record:

- the date
- what changed
- why it changed
- where the new source of truth lives

## Documentation Map

### Public

- `README.md`
- `README.zh-CN.md`
- `docs/README.md`
- `docs/README.zh-CN.md`
- `docs/FEISHU_SETUP.md`
- `docs/FEISHU_SETUP.zh-CN.md`
- `docs/MACOS_PERMISSIONS.md`
- `docs/MACOS_PERMISSIONS.zh-CN.md`
- `docs/DIAGNOSTICS.md`
- `docs/DIAGNOSTICS.zh-CN.md`
- `docs/PRODUCT_ROADMAP.md`
- `docs/PRODUCT_ROADMAP.zh-CN.md`
- `docs/EMBEDDING.md`
- `docs/EMBEDDING.zh-CN.md`
- `go-copyagentd/README.md`
- `go-copyagentd/README.zh-CN.md`
- `copyagent-ui-mac/README.md`
- `copyagent-ui-mac/README.zh-CN.md`
- `CONTRIBUTING.md`
- `CONTRIBUTING.zh-CN.md`
- `SECURITY.md`
- `SECURITY.zh-CN.md`
- `CHANGELOG.md`
- `CHANGELOG.zh-CN.md`

### Internal

- `docs/DEVELOPMENT.md`
- `docs/internal/README.md`
- `docs/internal/HANDOFF.md`
- `docs/internal/HANDOFF.history.md`
- `docs/internal/MEMORY_AUDIT.md`

## Current Working Assumptions

- keep dependencies lightweight
- never place real secrets in repository docs or examples
- clipboard writes must remain behind validation and authorization
- the daemon should remain usable without the optional macOS UI
- public docs should describe the product as copy-agent, not as a derivative of any other tool

## Change Log

### 2026-04-25

- Confirmed the Go daemon as the primary production runtime direction
- Preserved the Node implementation as a legacy prototype and regression reference
- Established the daemon + optional UI split

### 2026-04-26

- Tightened the public release promise around Direct Mode as stable
- Marked Agent Mode and foreground hosting as experimental
- Added first public setup, permissions, diagnostics, contribution, and security docs
- Removed tracked `NEXT_SESSION.md` from the public repository surface and kept local-only restart notes untracked

### 2026-04-27

- Reorganized readable project documentation into a release-oriented structure
- Switched public docs from inline bilingual blocks to GitHub-style language-switch files
- Moved session-style archives out of the repository root into `docs/internal/`
- Rewrote public docs to describe copy-agent as its own product rather than through implementation ancestry
- Standardized the public-facing product name as `copy-agent` while keeping runtime identifiers such as `copyagentd` and `~/.copyagent`
- Refined the bilingual opening product narrative around lightweight execution-agent positioning and low token cost
- Added full user-facing installation and usage documentation in both English and Simplified Chinese:
  - `docs/INSTALL.md`
  - `docs/INSTALL.zh-CN.md`
  - `docs/USAGE.md`
  - `docs/USAGE.zh-CN.md`
- Added dedicated API integration documentation clarifying three separate surfaces:
  - Feishu/Lark bot credentials,
  - Agent Mode local CLI integration,
  - advanced local HTTP/token fields.
- Expanded Feishu/Lark setup docs to include:
  - bot installation context,
  - recommended config example,
  - first-time verification order,
  - practical notes and failure guidance.
- Removed remaining user-visible upstream branding from the macOS UI About panel and related public-facing metadata while preserving required legal attribution in `NOTICE.md` and license files.
- Synced the macOS UI marketing version away from the inherited upstream value to the project-owned `0.1.0`.
- Reduced nonessential upstream naming in `specs/`, internal notes, and noncritical comments while keeping legal attribution and implementation-relevant issue references.
- Stopped surfacing `docs/DEVELOPMENT.md` from the public documentation index; it remains an internal maintainer memory file because the repository workflow still depends on it.
- Ran an open-source first-install validation with isolated temporary `HOME` directories, fake `launchctl`, fresh Go caches, and no dependency on an existing `~/.copyagent` or installed `copyagentd`.
- Verified the documented install, config generation, `doctor`, `copy` / `pbpaste`, service status/log/restart/uninstall, and full uninstall paths under the isolated environment.
- Recorded launchd start/stop as externally constrained for sandboxed validation because the LaunchAgent label is per-user global; tests used a `launchctl` shim to avoid touching any real local service.
- Tightened first-install user-facing text by aligning installer/uninstaller product wording with `copy-agent` while preserving runtime names such as `copyagentd`, `~/.copyagent`, and `copyagent.app`.
- Clarified Feishu/Lark setup docs that the current daemon uses SDK WebSocket / long-connection event delivery, not webhook callback delivery.
- Fixed optional UI source installation to pass an explicit Xcode `-derivedDataPath` under the active `HOME`, so `scripts/install.sh --with-ui` can be validated in an isolated temporary home instead of searching the real user's DerivedData.
- Audited post-cleanup residual configuration and removed two release-blocking local-machine leaks:
  - service fallback `PATH` no longer hardcodes a developer-specific home path and now derives portable defaults from the current `HOME`
  - `copyagent-ui-mac/Copyagent.xcodeproj` no longer embeds a personal Apple development team or personal signing identities
- Normalized historical/internal `specs/*` records so developer-specific absolute paths are no longer required to understand prior validation evidence.
- Prepared the first public GitHub release against a clean-history public repository at `sidney061212-ai/copy-agent`; the pre-release private history was intentionally kept private because it contains local handoff paths and personal signing metadata.
- Added an npm `overrides` pin for `axios` so the retained legacy Node prototype no longer inherits the moderate axios advisories through `@larksuiteoapi/node-sdk`.
- Completed release preflight checks before opening the repository:
  - current-file secret scan found only placeholders and test fixtures
  - git-history scan for common GitHub/OpenAI/Slack token patterns, private keys, and bearer-token literals returned no hits
  - `npm ci --ignore-scripts`
  - `npm audit --audit-level=moderate`
  - `npm test`
  - `go test ./...`
  - `go build -trimpath ./cmd/copyagentd`
  - `bash -n scripts/install.sh scripts/uninstall.sh scripts/build-dev.sh`
  - `git diff --check`
  - `xcodebuild -project Copyagent.xcodeproj -scheme Copyagent -configuration Debug -destination 'platform=macOS,arch=arm64' -derivedDataPath /tmp/copyagent-open-source-release-DerivedData build`
- Recorded remaining release caveats:
  - full LaunchAgent start/stop behavior still requires a real per-user macOS login session and should not be treated as fully proven by sandboxed tests alone
  - real Feishu/Lark delivery still requires user-owned app credentials and bot installation
  - the optional macOS UI builds successfully but still emits Swift 6 future-mode warnings around `AppState: Sendable`
