# Implementation Plan: copyagent

## Architecture

copyagent is a lightweight Node.js CLI and local HTTP service. It writes to the macOS clipboard with `pbcopy`; integration with local clipboard observers is passive because standard macOS clipboard monitoring sees normal system clipboard changes.

## Components

- `src/config.js`: environment parsing and validation.
- `src/clipboard.js`: safe `pbcopy` wrapper.
- `src/auth.js`: shared-token validation.
- `src/payloads.js`: generic and Feishu payload extraction.
- `src/server.js`: built-in HTTP server endpoints.
- `src/agent.js`: deterministic copy agent policy and action executor.
- `src/platforms/feishu.js`: Feishu/Lark event adapter and signature verifier.
- `src/platforms/generic.js`: generic JSON webhook adapter.
- `src/cli.js`: command-line interface.

## Endpoints

- `GET /health`: returns status and platform metadata.
- `POST /copy`: copies `{ text }` after token validation.
- `POST /feishu`: handles Feishu challenge and text message events after token validation for event callbacks.

## Security

- Do not hardcode app secrets or tokens.
- Require token for clipboard mutation.
- Verify Feishu/Lark verification token when configured.
- Verify Feishu/Lark `X-Lark-Signature` when encrypt key is configured and signature headers are present.
- Limit payload and copied text size.
- Avoid logging copied text unless explicit debug mode is enabled.

## Verification

- Unit tests for auth and payload extraction.
- Unit tests for server request handling with mocked clipboard writer.
- Manual smoke test using `pbpaste` on macOS.
