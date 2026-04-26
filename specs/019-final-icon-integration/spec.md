# Feature Specification: Final Icon Integration

**Created**: 2026-04-25  
**Status**: Draft

## Goal

Integrate the user-selected copyagent icon into the macOS UI app resources.

## Scope

- Use the provided selected 1024 PNG as the app icon source.
- Generate all required macOS AppIcon asset sizes.
- Replace the status bar image assets with scaled versions of the selected icon.
- Build the macOS UI and update the installed local app.

## Non-Goals

- Redesigning the icon.
- Changing UI layout or app behavior.
- Adding image-generation tooling or persistent temp assets.
