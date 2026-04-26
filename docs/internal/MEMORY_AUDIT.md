# Internal Memory Audit

copyagent uses a lightweight daemon plus optional menu bar UI architecture.

## Current Guidance

- Keep `copyagentd` as the always-on core process.
- Keep Feishu/Lark long connections, downloads, and image clipboard writes in the daemon.
- Keep the Swift app as a companion UI, not a second bot service.
- Do not reintroduce a Node background process for the primary runtime.
- Measure before removing mature clipboard-history UI features.

## Suggested Commands

```bash
ps -axo pid,rss,command | grep -E 'copyagentd|copyagent.app|copyagent' | grep -v grep
/usr/bin/time -l ~/.local/bin/copyagentd doctor
```

For UI measurements, compare current RSS and macOS physical footprint after the app has been idle for at least one minute.
