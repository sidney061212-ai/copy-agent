# Product Roadmap

[English](./PRODUCT_ROADMAP.md) | [简体中文](./PRODUCT_ROADMAP.zh-CN.md)

This roadmap describes the intended public direction for copy-agent.

## Product Goal

copy-agent aims to be a lightweight, local-first automation bridge between chat tools and desktop actions.

The core principle is simple:

- deterministic actions first
- local execution first
- optional agent workflows second

## Current Release Shape

### Stable

- Direct Mode text copy
- image and file handling
- fixed replies
- LaunchAgent install, status, restart, and logs

### Experimental

- Agent Mode
- foreground-hosting commands such as `/turn` and `/inject`
- workflows that depend on macOS Accessibility or Automation

## Next

- clean public release packaging
- release-build automation
- clearer user-facing install and troubleshooting flow
- continued hardening of the daemon lifecycle

## Later

- signed and notarized macOS distribution
- easier packaging and update paths
- stronger foreground-app targeting and input-focus recovery
- broader transport support

## Non-Goals for the First Public Release

- claiming that all agent workflows are production-ready
- requiring the optional UI for normal operation
- hiding the distinction between stable and experimental features
