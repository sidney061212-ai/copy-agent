# Security Policy

[English](./SECURITY.md) | [简体中文](./SECURITY.zh-CN.md)

## Supported Release Shape

The current public release is:

- macOS-first
- source-build friendly
- stable in Direct Mode
- experimental in Agent Mode and foreground-hosting features

## Report a Vulnerability

Please do not open a public issue for:

- credential leaks
- unauthorized clipboard writes
- local privilege escalation
- unsafe local file writes

Preferred path:

1. use GitHub private vulnerability reporting if available
2. otherwise contact the maintainers privately first

Please include:

- affected version or commit
- reproduction steps
- expected versus actual behavior
- whether the issue affects Direct Mode, Agent Mode, or foreground hosting

Do not include real secrets in the report.
