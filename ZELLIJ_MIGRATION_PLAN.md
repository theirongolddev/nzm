# NZM: Named Zellij Manager

> **A standalone CLI tool for orchestrating AI coding agents in Zellij sessions.**
> 
> This is a **new tool**, not an extension of NTM. While inspired by NTM's design, NZM is purpose-built for Zellij with a custom Rust WASM plugin for proper pane targeting.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         NZM CLI (Go)                            │
│  cmd/nzm/main.go                                                │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Commands: spawn, send, status, attach, kill, save, restore ││
│  └─────────────────────────────────────────────────────────────┘│
│                              │                                   │
│                              ▼                                   │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  internal/zellij/client.go - Zellij CLI wrapper             ││
│  │  - Session management (zellij attach, kill-session, etc.)  ││
│  │  - Pipe communication with plugin                           ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                               │
                               │ zellij pipe --plugin nzm-agent
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    NZM Agent Plugin (Rust/WASM)                 │
│  plugin/nzm-agent/                                              │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Capabilities:                                              ││
│  │  • Subscribes to PaneUpdate events → maintains pane state  ││
│  │  • write_to_pane_id() → send keys to specific panes        ││
│  │  • focus_pane_with_id() → focus specific panes             ││
│  │  • Pipe IPC → receives commands, returns results           ││
│  │  • dump-screen equivalent → capture pane output            ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Zellij Session                            │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ cc_1     │ │ cc_2     │ │ cod_1    │ │ gmi_1    │           │
│  │ (Claude) │ │ (Claude) │ │ (Codex)  │ │ (Gemini) │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Details

### 1. NZM CLI (Go)

**Purpose**: User-facing CLI tool for managing Zellij sessions with AI agents.

**Structure**:
```
nzm/
├── cmd/nzm/
│   └── main.go              # Entry point
├── internal/
│   ├── cli/                 # Cobra commands
│   │   ├── root.go
│   │   ├── spawn.go         # Create session with agents
│   │   ├── send.go          # Send prompts to agents
│   │   ├── status.go        # Show session status
│   │   ├── attach.go        # Attach to session
│   │   ├── kill.go          # Kill session
│   │   ├── save.go          # Save session state
│   │   └── restore.go       # Restore session
│   ├── zellij/
│   │   ├── client.go        # Zellij CLI wrapper
│   │   ├── session.go       # Session management
│   │   ├── plugin.go        # Plugin communication via pipes
│   │   └── types.go         # Pane, Session structs
│   ├── config/
│   │   └── config.go        # TOML configuration
│   ├── session/
│   │   ├── state.go         # Session state types
│   │   ├── capture.go       # Capture session state
│   │   └── storage.go       # Save/load from disk
│   └── output/
│       └── format.go        # JSON/text output formatting
├── go.mod
├── go.sum
└── Makefile
```

### 2. NZM Agent Plugin (Rust/WASM)

**Purpose**: Background plugin that enables pane-specific operations via CLI.

**Structure**:
```
plugin/nzm-agent/
├── Cargo.toml
├── src/
│   ├── main.rs              # Plugin entry point
│   ├── state.rs             # Pane state management
│   ├── commands.rs          # Command handlers
│   ├── ipc.rs               # Pipe message protocol
│   └── capture.rs           # Output capture logic
└── zellij.kdl               # Development layout
```

**Plugin Capabilities**:

| Capability | Zellij API | Description |
|------------|------------|-------------|
| List panes | `PaneUpdate` event | Subscribe to get `PaneManifest` with all pane info |
| Write to pane | `write_chars_to_pane_id()` | Send text to specific pane's STDIN |
| Focus pane | `focus_pane_with_id()` | Focus a specific pane |
| Get pane info | `get_focused_pane()` | Get `PaneInfo` for focused pane |
| Clear pane | `clear_screen_for_pane_id()` | Clear scrollback |
| Rename pane | `rename_terminal_pane()` | Set pane title |
| Close pane | `close_terminal_pane()` | Kill a pane |
| IPC | Pipe messages | Receive commands from CLI, send responses |

**Required Permissions**:
- `ReadApplicationState` - For `PaneUpdate` events
- `WriteToStdin` - For `write_to_pane_id()`
- `ChangeApplicationState` - For focus, rename, close operations

---

## Plugin IPC Protocol

Communication between NZM CLI and the plugin uses Zellij pipes with JSON payloads.

### Request Format (CLI → Plugin)

```json
{
  "id": "uuid-request-id",
  "action": "send_keys",
  "params": {
    "pane_id": 3,
    "text": "fix the bug in auth.go",
    "enter": true
  }
}
```

### Response Format (Plugin → CLI)

```json
{
  "id": "uuid-request-id",
  "success": true,
  "data": {
    "pane_id": 3,
    "chars_sent": 25
  }
}
```

### Supported Actions

| Action | Params | Description |
|--------|--------|-------------|
| `list_panes` | `{}` | Get all panes with their info |
| `send_keys` | `{pane_id, text, enter}` | Send text to pane |
| `send_interrupt` | `{pane_id}` | Send Ctrl+C to pane |
| `focus_pane` | `{pane_id}` | Focus a specific pane |
| `rename_pane` | `{pane_id, title}` | Set pane title |
| `close_pane` | `{pane_id}` | Close a pane |
| `capture_output` | `{pane_id, lines}` | Capture pane output (via dump) |
| `get_pane_info` | `{pane_id}` | Get detailed pane info |

### CLI Usage Example

```bash
# Send command to plugin
echo '{"id":"1","action":"list_panes","params":{}}' | \
  zellij pipe --plugin file:~/.local/share/nzm/nzm-agent.wasm --name nzm
```

The Go client wraps this:

```go
func (c *Client) ListPanes(session string) ([]Pane, error) {
    req := Request{
        ID:     uuid.New().String(),
        Action: "list_panes",
        Params: map[string]any{},
    }
    resp, err := c.sendPluginCommand(session, req)
    if err != nil {
        return nil, err
    }
    return parsePanes(resp.Data)
}
```

---

## Pane Naming Convention

Same as NTM for consistency:

```
{session}__{type}_{index}[_{variant}][tags]
```

**Examples**:
- `myproject__cc_1` - Claude agent 1
- `myproject__cc_2_opus` - Claude agent 2 with Opus variant
- `myproject__cod_1[backend,api]` - Codex agent 1 with tags
- `myproject__gmi_1` - Gemini agent 1
- `myproject__user_1` - User pane

**Agent Types**:
- `cc` = Claude Code
- `cod` = Codex
- `gmi` = Gemini
- `user` = User/shell pane

---

## Session Lifecycle

### 1. Spawn Session

```bash
nzm spawn myproject --cc=2 --cod=1 --gmi=1
```

**Flow**:
1. Generate KDL layout file with pane definitions
2. Start Zellij: `zellij --session myproject --layout /tmp/nzm-layout.kdl`
3. Ensure nzm-agent plugin is loaded
4. Wait for agents to initialize (check for ready indicators)
5. Report success with pane summary

**Generated Layout** (`/tmp/nzm-myproject.kdl`):
```kdl
layout {
    cwd "/path/to/project"
    
    // NZM agent plugin (hidden, runs in background)
    pane plugin="file:~/.local/share/nzm/nzm-agent.wasm" {
        borderless true
        size "1%"
    }
    
    // Agent panes
    pane name="myproject__cc_1" command="claude" {
        focus true
    }
    pane name="myproject__cc_2" command="claude"
    pane name="myproject__cod_1" command="codex"
    pane name="myproject__gmi_1" command="gemini"
}
```

### 2. Send Command

```bash
nzm send myproject --all "review the PR"
```

**Flow**:
1. Connect to session via pipe
2. Request pane list from plugin
3. For each target pane, send `send_keys` command
4. Wait for acknowledgment
5. Report results

### 3. Status

```bash
nzm status myproject
```

**Flow**:
1. Query plugin for pane list with details
2. Format and display pane status

### 4. Attach

```bash
nzm attach myproject
```

**Flow**:
1. Execute `zellij attach myproject`

### 5. Kill

```bash
nzm kill myproject
```

**Flow**:
1. Execute `zellij kill-session myproject`

---

## Output Capture Strategy

Zellij doesn't have tmux's `capture-pane -p` equivalent that outputs to stdout. Options:

### Option A: dump-screen + Read (Recommended)

```go
func (c *Client) CaptureOutput(session string, paneID int, lines int) (string, error) {
    // 1. Focus the pane (required for dump-screen)
    c.FocusPane(session, paneID)
    
    // 2. Dump to temp file
    tmpFile := fmt.Sprintf("/tmp/nzm-capture-%d.txt", paneID)
    c.runAction(session, "dump-screen", tmpFile)
    
    // 3. Read and return last N lines
    content, _ := os.ReadFile(tmpFile)
    return lastNLines(string(content), lines), nil
}
```

**Pros**: Works with current Zellij
**Cons**: Requires focus switching, file I/O

### Option B: Plugin-Based Capture (Future)

If Zellij adds scrollback access API, the plugin could provide direct capture.

### Option C: Scrollback Serialization

Configure Zellij to serialize scrollback on exit:
```kdl
pane_viewport_serialization true
scrollback_lines_to_serialize 10000
```

Then read from session resurrection data.

---

## Key Differences from NTM

| Aspect | NTM (tmux) | NZM (Zellij) |
|--------|------------|--------------|
| Pane targeting | Direct: `send-keys -t %5` | Via plugin: `write_to_pane_id()` |
| Pane listing | `list-panes -F format` | Plugin state from `PaneUpdate` |
| Output capture | `capture-pane -p` | `dump-screen` to file |
| Remote SSH | Native support | Deferred (command forwarding possible) |
| Activity tracking | `#{pane_last_activity}` | Plugin-based or heartbeat |
| Layout | Dynamic: `select-layout` | Static: KDL layout files |
| Plugin required | No | Yes (nzm-agent) |

---

## Implementation Phases

### Phase 1: Plugin Foundation

**Goal**: Working nzm-agent plugin with core capabilities.

**Tasks**:
1. Scaffold Rust plugin with `zellij-tile`
2. Implement `PaneUpdate` subscription and state tracking
3. Implement pipe message handling (IPC)
4. Implement `list_panes` command
5. Implement `send_keys` command using `write_chars_to_pane_id()`
6. Implement `send_interrupt` command
7. Add tests using Zellij plugin test utilities

**Deliverable**: `nzm-agent.wasm` that can list panes and send keys via pipe.

### Phase 2: CLI Foundation

**Goal**: Basic NZM CLI that communicates with plugin.

**Tasks**:
1. Initialize Go project structure
2. Implement Zellij client wrapper (`internal/zellij/`)
3. Implement plugin communication via pipes
4. Implement `nzm spawn` command with layout generation
5. Implement `nzm send` command
6. Implement `nzm status` command
7. Implement `nzm attach` and `nzm kill` commands
8. Add unit and integration tests

**Deliverable**: Working `nzm` CLI with spawn, send, status, attach, kill.

### Phase 3: Session Management

**Goal**: Save/restore sessions, configuration.

**Tasks**:
1. Implement session state capture
2. Implement session save/restore
3. Add TOML configuration support
4. Add agent command customization
5. Add preset support (like NTM's presets)

**Deliverable**: Full session lifecycle management.

### Phase 4: Advanced Features

**Goal**: Feature parity with key NTM capabilities.

**Tasks**:
1. Output capture implementation
2. Broadcast mode (send to all/filtered panes)
3. Pane tagging system
4. Activity detection (plugin-based)
5. TUI dashboard (optional, using BubbleTea)
6. Shell completions

**Deliverable**: Production-ready NZM.

### Phase 5: Polish & Distribution

**Goal**: Release-ready tool.

**Tasks**:
1. Comprehensive documentation
2. Installation scripts
3. GitHub Actions for releases
4. Plugin distribution (GitHub releases)
5. Homebrew formula / AUR package

---

## Testing Strategy

### Plugin Tests (Rust)

```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_parse_pipe_message() {
        let msg = r#"{"id":"1","action":"list_panes","params":{}}"#;
        let req: Request = serde_json::from_str(msg).unwrap();
        assert_eq!(req.action, "list_panes");
    }
    
    #[test]
    fn test_pane_state_update() {
        let mut state = State::default();
        // Simulate PaneUpdate event
        let manifest = create_test_manifest();
        state.update_panes(manifest);
        assert_eq!(state.panes.len(), 4);
    }
}
```

### CLI Tests (Go)

```go
func TestSpawnGeneratesLayout(t *testing.T) {
    opts := SpawnOptions{
        Session:  "test",
        CCCount:  2,
        CodCount: 1,
    }
    layout, err := generateLayout(opts)
    require.NoError(t, err)
    assert.Contains(t, layout, "test__cc_1")
    assert.Contains(t, layout, "test__cc_2")
    assert.Contains(t, layout, "test__cod_1")
}

func TestSendKeysParsesResponse(t *testing.T) {
    // Mock plugin response
    resp := `{"id":"1","success":true,"data":{"chars_sent":25}}`
    result, err := parseResponse(resp)
    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

### Integration Tests

```bash
#!/bin/bash
# test/integration/spawn_test.sh

# Test: spawn creates session with correct panes
nzm spawn test-session --cc=2 --cod=1

# Verify session exists
zellij list-sessions | grep -q "test-session"

# Verify panes via plugin
panes=$(nzm status test-session --json | jq '.panes | length')
[[ "$panes" -eq 3 ]] || exit 1

# Cleanup
nzm kill test-session
```

---

## Dependencies

### Go CLI

```go
// go.mod
module github.com/youruser/nzm

go 1.22

require (
    github.com/spf13/cobra v1.8.0
    github.com/google/uuid v1.6.0
    github.com/pelletier/go-toml/v2 v2.2.0
    // Optional for TUI:
    github.com/charmbracelet/bubbletea v0.25.0
)
```

### Rust Plugin

```toml
# Cargo.toml
[package]
name = "nzm-agent"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
zellij-tile = "0.40"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

[profile.release]
opt-level = "s"
lto = true
```

---

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/youruser/nzm
cd nzm

# Build plugin
cd plugin/nzm-agent
cargo build --release --target wasm32-wasi
mkdir -p ~/.local/share/nzm
cp target/wasm32-wasi/release/nzm_agent.wasm ~/.local/share/nzm/nzm-agent.wasm

# Build CLI
cd ../..
go build -o nzm ./cmd/nzm
sudo mv nzm /usr/local/bin/
```

### One-liner (Future)

```bash
curl -sSL https://raw.githubusercontent.com/youruser/nzm/main/install.sh | bash
```

---

## Configuration

```toml
# ~/.config/nzm/config.toml

[agents]
claude_cmd = "claude"
codex_cmd = "codex"
gemini_cmd = "gemini"

[defaults]
working_directory = "."
layout = "tiled"

[presets.fullstack]
cc = 2
cod = 1
gmi = 1

[presets.solo]
cc = 1
```

---

## Sources & References

### Zellij Documentation
- [Zellij Plugin Development Tutorial](https://zellij.dev/tutorials/developing-a-rust-plugin/)
- [Plugin API Commands](https://zellij.dev/documentation/plugin-api-commands.html)
- [Plugin API Events](https://zellij.dev/documentation/plugin-api-events.html)
- [Plugin Pipes](https://zellij.dev/documentation/plugin-pipes)
- [Creating Layouts](https://zellij.dev/documentation/creating-a-layout.html)
- [CLI Actions](https://zellij.dev/documentation/cli-actions)

### Rust API
- [zellij-tile crate](https://docs.rs/zellij-tile/latest/zellij_tile/)
- [zellij-tile shim module](https://docs.rs/zellij-tile/latest/zellij_tile/shim/index.html)
- [PaneInfo struct](https://docs.rs/zellij-utils/latest/zellij_utils/data/struct.PaneInfo.html)

### Examples
- [Rust Plugin Example](https://github.com/zellij-org/rust-plugin-example)
- [Create Rust Plugin Scaffolding](https://github.com/zellij-org/create-rust-plugin)

### Community
- [Zellij GitHub Discussions](https://github.com/zellij-org/zellij/discussions)
- [awesome-zellij](https://github.com/zellij-org/awesome-zellij)
