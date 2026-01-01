# NZM Feature Parity Plan

## Current Status

### Already Implemented in NZM (Zellij-native)
- [x] `internal/zellij/` - Zellij client (replaces `internal/tmux/`)
  - Session management: ListSessions, SessionExists, CreateSession, KillSession, AttachSession
  - Pane management: ListPanes, GetPanesEnriched, GetPaneInfo, SetPaneTitle
  - Pane I/O: SendKeys, SendInterrupt, CapturePaneOutput, GetPaneActivity
  - Agent types: AgentType, Pane struct with metadata parsing
  - Utilities: SanitizePaneCommand, BuildPaneCommand, FormatPaneName
- [x] `internal/nzm/` - Core commands (spawn, send, status, kill)
- [x] `plugin/nzm-agent/` - Rust WASM plugin for pane communication
- [x] `cmd/nzm/` - CLI entry point with --config and --json flags
- [x] `internal/config/zellij.go` - NZM config system (ZellijConfig, NZMConfig, NZMLoad)
- [x] `internal/output/nzm.go` - NZM output formatting (NZMDetectFormat, NZMDefaultFormatter)

### NTM Features to Port

#### Priority 1: Core Features (Required for basic use)
| Feature | NTM Package | NZM Status | Notes |
|---------|-------------|------------|-------|
| Session spawn | `cli/spawn.go` | ✅ Done | Basic spawn implemented |
| Send to panes | `cli/send.go` | ✅ Done | Basic send implemented |
| Status display | `cli/status.go` | ✅ Done | Basic status implemented |
| Kill session | `cli/kill.go` | ✅ Done | Basic kill implemented |
| Attach session | `cli/attach.go` | ✅ Done | Basic attach implemented |
| Config system | `config/` | ✅ Done | NZMConfig, ZellijConfig, NZMLoad |
| Output formatting | `output/` | ✅ Done | NZMDetectFormat, --json flag |

#### Priority 2: Robot Mode (Core differentiator)
| Feature | NTM Package | NZM Status | Notes |
|---------|-------------|------------|-------|
| Robot orchestrator | `robot/robot.go` | ❌ TODO | Main robot loop |
| Task routing | `robot/routing.go` | ❌ TODO | Route tasks to agents |
| Activity detection | `robot/activity.go` | ❌ TODO | Detect agent state |
| Pattern matching | `robot/patterns.go` | ❌ TODO | Detect prompts/errors |
| Agent acknowledgment | `robot/ack.go` | ❌ TODO | Track task completion |

#### Priority 3: Dashboard TUI
| Feature | NTM Package | NZM Status | Notes |
|---------|-------------|------------|-------|
| Dashboard main | `tui/dashboard/` | ❌ TODO | Main TUI |
| Components | `tui/components/` | ❌ TODO | Reusable widgets |
| Styles/themes | `tui/styles/`, `tui/theme/` | ❌ TODO | Visual styling |
| Layout system | `tui/layout/` | ❌ TODO | Panel layout |

#### Priority 4: Agent Communication
| Feature | NTM Package | NZM Status | Notes |
|---------|-------------|------------|-------|
| Agent mail | `agentmail/` | ❌ TODO | Inter-agent messaging |
| Status detection | `status/` | ❌ TODO | Detect agent states |
| Alerts | `alerts/` | ❌ TODO | Error/completion alerts |

#### Priority 5: Advanced Features
| Feature | NTM Package | NZM Status | Notes |
|---------|-------------|------------|-------|
| Pipelines | `pipeline/` | ❌ TODO | Task automation |
| Templates | `templates/` | ❌ TODO | Prompt templates |
| Recipes | `recipe/` | ❌ TODO | Saved workflows |
| Hooks | `hooks/` | ❌ TODO | Pre/post hooks |
| Personas | `persona/` | ❌ TODO | Agent personas |
| History | `history/` | ❌ TODO | Command history |
| Quota tracking | `quota/` | ❌ TODO | API usage |
| File watcher | `watcher/` | ❌ TODO | Watch for changes |
| Scanner | `scanner/` | ❌ TODO | Code scanning |
| Profiler | `profiler/` | ❌ TODO | Performance |

#### Lower Priority / Optional
| Feature | NTM Package | NZM Status | Notes |
|---------|-------------|------------|-------|
| Tutorial | `tutorial/` | ⏸️ Later | Interactive tutorial |
| Updater | `updater/` | ⏸️ Later | Auto-update |
| Checkpoints | `checkpoint/` | ⏸️ Later | Session save |
| Session persist | `session/` | ⏸️ Later | Session restore |
| Clipboard | `clipboard/` | ⏸️ Later | Copy output |
| Command palette | `palette/` | ⏸️ Later | Quick commands |
| Plugins | `plugins/` | ⏸️ Later | Plugin system |
| CASS integration | `cass/` | ⏸️ Later | Memory search |
| Context rotation | `context/` | ⏸️ Later | Context management |

## Porting Strategy

### Phase 1: Terminal-Agnostic Packages (No tmux dependency)
These packages don't depend on tmux and can be copied with minimal changes:
- `config/` - Just update module path
- `output/` - No terminal deps
- `events/` - No terminal deps
- `tokens/` - No terminal deps
- `util/` - No terminal deps
- `status/patterns.go`, `status/errors.go` - Pattern matching only

### Phase 2: Packages Needing Zellij Adaptation
These use tmux but can be adapted to use `internal/zellij`:
- `robot/` - Replace tmux client with zellij client
- `status/` - Use zellij plugin for pane content
- `alerts/` - Use zellij for pane detection

### Phase 3: TUI (Mostly terminal-agnostic)
The bubbletea TUI is terminal-agnostic, just needs zellij integration:
- `tui/` - Replace tmux calls with zellij

### Phase 4: Complex Features
- `pipeline/` - Needs robot mode first
- `agentmail/` - May need redesign for zellij

## Module Path

When porting, update imports from:
```go
github.com/Dicklesworthstone/ntm/internal/...
```
to:
```go
github.com/theirongolddev/nzm/internal/...
```

## Testing Strategy

1. Port package with tests
2. Update imports
3. Run tests to verify
4. Integration test with zellij
