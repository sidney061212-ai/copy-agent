# copy-agent

[English](./README.md) | [简体中文](./README.zh-CN.md)

copy-agent is a lightweight clipboard agent for the AI era. It is more than a clipboard utility: it is a bridge between phone and computer, and a lightweight agent for remotely resuming local AI workflows.

When you have to step away from your desk, copy-agent can use Feishu/Lark to safely send messages, images, files, and tasks back to your machine, so local Claude, Codex, VS Code, Cursor, and CLI-based coding workflows can keep moving instead of stopping with your desktop session.

It also solves a simpler but very common problem: moving text, images, and files from your phone to your computer still often means opening apps, copying, and pasting by hand. In the AI era, the machine running a local agent is often not a place where you want to sign into personal chat apps, which makes device-to-device transfer even more awkward. copy-agent is built to reopen that path in a lightweight, safe, and controllable way.

Some of these workflows can also be handled by more heavily trained systems such as OpenClaw. But not every AI power user has an unlimited token budget. For lower-value, lower-complexity, high-frequency tasks, copy-agent chooses an execution-first path that dramatically reduces unnecessary token burn. The project will keep pushing toward a lighter architecture and extremely low token cost, with the goal of becoming a genuinely practical execution agent.

The public project name is `copy-agent`; the current daemon binary and local paths still use historical runtime names such as `copyagentd` and `~/.copyagent`.

## Release Status

The current public release is intentionally narrow.

| Track | Status | Scope |
|---|---|---|
| Direct Mode | Stable | Deterministic clipboard, image, file, reply, and service workflows |
| Agent Mode | Experimental | Local coding-agent routing |
| Foreground hosting | Experimental | `/turn`, `/inject`, `reply-text`, and any workflow that depends on macOS Accessibility or Automation |

If you are trying copy-agent for the first time, start with Direct Mode.

## Components

- `copyagentd` — lightweight Go daemon for Feishu/Lark transport, deterministic actions, diagnostics, and LaunchAgent lifecycle
- `copyagent-ui-mac` — optional macOS menu bar companion for clipboard history and local desktop UX

The UI is optional. The daemon is the product core.

## Quick Start

For a full step-by-step install, read:

- `docs/INSTALL.md`
- `docs/FEISHU_SETUP.md`
- `docs/USAGE.md`
- `docs/API_INTEGRATION.md`

```bash
git clone https://github.com/sidney061212-ai/copy-agent.git
cd copy-agent
scripts/install.sh
chmod 600 ~/.copyagent/config.json
open -e ~/.copyagent/config.json
```

Set at least:

- `feishuAppId`
- `feishuAppSecret`

Then verify:

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
~/.local/bin/copyagentd service status
```

## Documentation

- Documentation index: `docs/README.md`
- Install guide: `docs/INSTALL.md`
- Usage guide: `docs/USAGE.md`
- API integration guide: `docs/API_INTEGRATION.md`
- Feishu/Lark setup: `docs/FEISHU_SETUP.md`
- macOS permissions: `docs/MACOS_PERMISSIONS.md`
- Diagnostics: `docs/DIAGNOSTICS.md`
- Product roadmap: `docs/PRODUCT_ROADMAP.md`
- Embedding notes: `docs/EMBEDDING.md`

Component-specific docs:

- daemon runtime: `go-copyagentd/README.md`
- macOS UI companion: `copyagent-ui-mac/README.md`

## Common Commands

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service install
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service restart
scripts/uninstall.sh
```

Optional UI install:

```bash
scripts/install.sh --with-ui
open ~/Applications/copyagent.app
```

## Product Model

copy-agent currently has two operating modes:

- **Direct Mode** — chat messages trigger deterministic actions directly
- **Agent Mode** — chat messages can be routed into a configured local coding-agent session

Direct Mode is the release baseline. Agent Mode and foreground-hosting flows are still being hardened.

## Repository Layout

- `go-copyagentd/` — production daemon runtime
- `copyagent-ui-mac/` — optional macOS UI
- `docs/` — user and engineering documentation
- `config/` — safe configuration template
- `scripts/` — install, uninstall, and build helpers
- `specs/` — feature specs and implementation records

## Security and Privacy

- Keep `~/.copyagent/config.json` local and permissioned as `0600`
- Never commit real app secrets, tokens, or local logs
- Review diagnostics before sharing logs publicly

See `SECURITY.md` for reporting guidance.

## License

copy-agent is released under the MIT License.

The macOS UI includes third-party licensed code. See `NOTICE.md` and `copyagent-ui-mac/LICENSE` for attribution details.
