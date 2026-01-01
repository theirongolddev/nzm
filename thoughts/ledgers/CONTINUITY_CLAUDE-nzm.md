# Session: nzm
Updated: 2026-01-01

## Goal
**NZM (Named Zellij Manager)** - A standalone CLI tool for orchestrating AI coding agents in Zellij sessions. This is a new tool (not extending NTM) purpose-built for Zellij with a custom Rust WASM plugin.

## Constraints
- Go CLI (go 1.22+) + Rust plugin (WASM)
- TDD workflow required per CLAUDE.md
- Must run tests before completing tasks
- Target Zellij 0.40+ for stable plugin API

## Architecture
```
┌─────────────────────┐     zellij pipe      ┌─────────────────────┐
│   NZM CLI (Go)      │ ◄──────────────────► │ nzm-agent (Rust)    │
│   cmd/nzm/          │     JSON IPC         │ plugin/nzm-agent/   │
└─────────────────────┘                      └─────────────────────┘
         │                                            │
         └────────────────────┬───────────────────────┘
                              ▼
                    ┌─────────────────────┐
                    │   Zellij Session    │
                    │   [cc_1] [cc_2]     │
                    │   [cod_1] [gmi_1]   │
                    └─────────────────────┘
```

## Key Decisions
1. **Standalone tool** - Not extending NTM, clean Zellij-native implementation
2. **Rust WASM plugin** - For proper pane targeting via `write_chars_to_pane_id()`
3. **Pipe IPC** - CLI communicates with plugin via `zellij pipe` + JSON
4. **Plugin tracks panes** - Subscribes to `PaneUpdate` events for pane state
5. **KDL layouts** - Generate layouts dynamically at spawn time

## State
- Done:
  - Explored NTM codebase (~87 files use tmux)
  - Researched Zellij plugin API extensively
  - Created detailed architecture plan (`ZELLIJ_MIGRATION_PLAN.md`)
  - Identified key Zellij APIs: `write_chars_to_pane_id()`, `PaneUpdate`, pipes
  - **Phase 1 complete**: Rust plugin with 30 tests, WASM builds (856 KB)
    - IPC protocol (Request/Response parsing)
    - Pane state management (PaneUpdate event handling)
    - Command handlers (list_panes, send_keys, send_interrupt, get_pane_info)
    - Plugin main loop (pipe message handling)
- Now: Phase 2 - Go CLI foundation
- Next: spawn, send, status, attach, kill commands

## Implementation Phases
1. **Phase 1**: Rust plugin (list_panes, send_keys, send_interrupt) ✅
2. **Phase 2**: Go CLI (spawn, send, status, attach, kill)
3. **Phase 3**: Session management (save/restore, config)
4. **Phase 4**: Advanced features (capture, broadcast, tagging)
5. **Phase 5**: Polish & distribution

## Open Questions
- Output capture strategy (dump-screen vs plugin API if added)
- Remote SSH support (defer to later phase?)

## Working Set
- Branch: `main`
- Key files:
  - Plan: `ZELLIJ_MIGRATION_PLAN.md`, `NZM_TDD_DEVELOPMENT_PLAN.md`
  - Plugin: `plugin/nzm-agent/` (implemented)
  - CLI: `cmd/nzm/`, `internal/` (to be created)
- Plugin build: `cargo build --release --target wasm32-wasip1`
- CLI build: `go build -o nzm ./cmd/nzm`
- Test: `cargo test` (plugin), `go test ./...` (CLI)

## References
- [Zellij Plugin Tutorial](https://zellij.dev/tutorials/developing-a-rust-plugin/)
- [Plugin API Commands](https://zellij.dev/documentation/plugin-api-commands.html)
- [Plugin API Events](https://zellij.dev/documentation/plugin-api-events.html)
- [zellij-tile crate](https://docs.rs/zellij-tile/latest/zellij_tile/)
