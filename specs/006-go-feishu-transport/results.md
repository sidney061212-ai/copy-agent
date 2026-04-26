# Results: Go Feishu Transport Spike

## Findings

- Official Go SDK includes `github.com/larksuite/oapi-sdk-go/v3/ws`.
- `ws.NewClient(appID, appSecret, ...)` compiles and initializes.
- `copyagentd feishu-probe --start` starts the long-connection client.
- Measured RSS after start: ~15.7 MB.

## Decision

Proceed with Go as production daemon path. The official Go Feishu SDK is acceptable for the next implementation phase.

## Next implementation phase

`007-go-feishu-text-reply`:

- Register `im.message.receive_v1` handler in Go.
- Normalize text messages.
- Copy text to clipboard.
- Reply with fixed template.
- Keep RSS target under 20 MB.
