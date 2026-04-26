# Implementation Plan: Core/Adapter Refactor

## Strategy

Perform a behavior-preserving refactor. Add new core modules and executor first, cover with tests, then route Feishu bot through the new path. Keep legacy entry points as compatibility wrappers where useful.

## Target Layout

```text
src/core/
  events.js
  rules.js
  plans.js
  replies.js
  policy.js

src/actions/
  executor.js
  clipboard.js
  filesystem.js
  reply.js
  media.js

src/transports/
  feishu/
    normalize.js
    bot.js
    client.js
```

## Validation

- Unit tests for core action planning.
- Unit tests for executor with fake adapters.
- Existing Feishu/media/server tests remain green.
- Manual status check: `copyagent status`.
