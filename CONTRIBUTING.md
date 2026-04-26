# Contributing

[English](./CONTRIBUTING.md) | [简体中文](./CONTRIBUTING.zh-CN.md)

Thank you for contributing to copy-agent.

## Contribution Principles

- Keep the product lightweight
- Protect local-first behavior and user data
- Keep Direct Mode predictable
- Treat experimental features as clearly experimental

## Before Opening a Pull Request

Run the relevant checks:

```bash
npm test
cd go-copyagentd && go test ./...
bash -n scripts/install.sh
bash -n scripts/uninstall.sh
```

If you changed public behavior, also update:

- `README.md`
- the relevant user-facing document in `docs/`
- `docs/DEVELOPMENT.md`

## Scope Expectations

For the current public release:

- Direct Mode is the stable baseline
- Agent Mode is experimental
- foreground-hosting commands are experimental

Please do not present experimental behavior as stable in docs, code comments, or release notes.

## Secrets and Private Data

Never commit:

- real `feishuAppSecret`
- tokens or API keys
- private local logs
- local config files

## Engineering Notes

- Prefer small, reversible changes
- Prefer Node.js built-ins when working in the legacy prototype
- Keep clipboard writes behind validation and authorization
- Update `docs/DEVELOPMENT.md` whenever a phase changes behavior or architecture
