# Diagnostics

[English](./DIAGNOSTICS.md) | [简体中文](./DIAGNOSTICS.zh-CN.md)

Use this page to verify an install or collect a useful bug report.

## Quick Health Check

```bash
~/.local/bin/copyagentd doctor
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
```

## Direct Mode Sanity Check

```bash
~/.local/bin/copyagentd copy 'hello from copy-agent'
pbpaste
```

Then send:

```text
copy hello
```

## Useful Service Commands

```bash
~/.local/bin/copyagentd service install
~/.local/bin/copyagentd service restart
~/.local/bin/copyagentd service status
~/.local/bin/copyagentd service logs
~/.local/bin/copyagentd service uninstall
```

## Useful Bug Report Data

Include:

- macOS version
- whether you tested Direct Mode or experimental foreground hosting
- output of `doctor`, `service status`, and `service logs`
- the exact user-visible failure

Do not include:

- real secrets
- full local config files
- private logs you have not reviewed

## Foreground-Hosting Issues

If the issue involves `/turn` or `/inject`, also record:

- target app name
- whether the app was already open
- whether the input field was focused
- whether the failure happened during activation, paste, or submit
