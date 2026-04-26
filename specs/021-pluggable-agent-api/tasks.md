# Tasks: cc-connect-Style Pluggable Clipboard Agent

## Phase 1: cc-connect-style transport skeleton

- [x] Add `internal/agent/message.go` with normalized message and attachment types.
- [x] Add `internal/agent/registry.go` modeled after cc-connect `core/registry.go`.
- [x] Add `Transport`, `MessageHandler`, `ReplyCapable`, and `ResourceCapable` interfaces.
- [x] Add registry tests for create/list/unknown transport errors.
- [x] Register future chat transport placeholders as disabled no-op plugins.

## Phase 2: Product mode config

- [x] Add `agent.enabled` config parsing.
- [x] Add `agent.type`, `agent.command`, `agent.sessionMode`, and `agent.idleTimeoutMins` config parsing.
- [x] Preserve existing flat Feishu JSON config compatibility.
- [ ] Document Direct Mode and Agent Mode in README and daemon docs.

## Phase 3: Direct Mode engine path

- [x] Add `internal/agent/engine.go` message handling path for Direct Mode.
- [x] Move actor allowlist, dedupe, and max text checks into `policy.go`.
- [x] Port deterministic Node planner behavior into Go `planner.go`.
- [x] Add executor for `copy_text`, `save_resource`, `copy_image`, and `reply`.
- [x] Use mock adapters in unit tests.

## Phase 4: Feishu as transport plugin

- [x] Add `cmd/copyagentd/plugin_transport_feishu.go` blank import.
- [x] Add `init()` registration inside `internal/transport/feishu`.
- [x] Refactor Feishu SDK handler to normalize into `agent.Message`.
- [x] Keep Feishu reply/resource download as optional capabilities.
- [x] Preserve current reply strings and image/file behavior.

## Phase 5: Coding agent adapters

- [x] Add `CodingAgent`, `AgentSession`, `AgentEvent`, and `AgentFactory` interfaces.
- [x] Add registry support for agent factories.
- [x] Add session store keyed by normalized `Message.SessionKey`.
- [x] Add per-session busy lock and queue; never send a second message mid-turn.
- [x] Add session resume fallback: if resume fails, clear stored ID and start fresh.
- [x] Add `cmd/copyagentd/plugin_agent_codex.go` blank import.
- [x] Add `cmd/copyagentd/plugin_agent_claude.go` blank import.
- [x] Implement minimal Codex CLI `exec --json` adapter with thread resume.
- [x] Implement minimal Claude Code persistent stdin/stdout JSON adapter.
- [ ] Add idle timeout/session lifecycle tests.

## Phase 6: Agent Mode routing

- [x] Fast-path explicit copy/file/image commands before LLM routing.
- [x] Forward non-command messages to configured CLI agent when `agent.enabled=true`.
- [x] Save inbound images/files and pass local paths to the agent prompt.
- [x] Inject copyagent action instructions into the agent prompt.
- [x] Relay agent output back through the originating transport.
- [x] Preserve Feishu reply context as opaque `ReplyCtx` across the engine boundary.

## Phase 6.5: Runtime mode controls

- [x] Add `/agent` command to enable Agent Mode at runtime.
- [x] Add `/copy` command to return to no-AI Direct Mode at runtime.
- [x] Persist runtime mode changes to `~/.copyagent/config.json`.
- [x] Add Feishu typing reaction during Agent Mode turns.
- [x] Ignore Feishu reaction created/deleted events triggered by typing reactions.
- [x] Live-test `/agent`, Agent reply, Direct fast-path, `/copy`, and typing reaction.

## Phase 7: Copyagent action API/CLI

- [ ] Add narrow `copyagent action copy-text` command.
- [ ] Add `copyagent action copy-image --from <path>` command.
- [ ] Add `copyagent action save-file --from <path>` command.
- [ ] Optionally back action CLI with `~/.copyagent/run/api.sock` later.

## Phase 8: Validation and docs

- [x] Update `config/config.example.json`.
- [ ] Update README, daemon README, UI README, and development docs.
- [ ] Run `go test ./...`, `npm test`, and Swift build/unit validation.
- [x] Update `results.md`.
