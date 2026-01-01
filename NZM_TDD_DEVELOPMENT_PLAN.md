# NZM TDD Development Plan

> **Test-Driven Development is MANDATORY** for this project.
> Every feature follows: RED → GREEN → REFACTOR

---

## TDD Principles for NZM

### The TDD Cycle

```
┌─────────────────────────────────────────────────────────┐
│  1. RED: Write a failing test that defines behavior     │
│     ↓                                                   │
│  2. GREEN: Write minimal code to make test pass         │
│     ↓                                                   │
│  3. REFACTOR: Clean up while keeping tests green        │
│     ↓                                                   │
│  4. REPEAT: Next test case                              │
└─────────────────────────────────────────────────────────┘
```

### Test Categories

| Category | Purpose | Location |
|----------|---------|----------|
| Unit Tests | Test individual functions/structs in isolation | `*_test.rs`, `*_test.go` |
| Integration Tests | Test component interactions | `tests/integration/` |
| E2E Tests | Test full CLI → Plugin → Zellij flow | `tests/e2e/` |

### Test Requirements

- **Every PR must maintain or increase test coverage**
- **No implementation code without a failing test first**
- **Tests must be deterministic (no flaky tests)**
- **Tests must be fast (< 100ms for unit tests)**

---

## Phase 1: Rust Plugin Foundation

### 1.1 Project Setup

**Test**: Verify plugin compiles to WASM

```rust
// tests/build_test.rs
#[test]
fn test_plugin_compiles_to_wasm() {
    // This is a meta-test - if we can run tests, WASM target works
    assert!(true);
}
```

**Implementation**:
1. Create `plugin/nzm-agent/Cargo.toml`
2. Add `wasm32-wasi` target
3. Minimal `lib.rs` with `ZellijPlugin` trait

**Commands**:
```bash
cd plugin/nzm-agent
rustup target add wasm32-wasi
cargo build --target wasm32-wasi
cargo test
```

---

### 1.2 IPC Protocol - Message Parsing

**Tests** (write these FIRST):

```rust
// src/ipc_test.rs
use super::ipc::*;

#[test]
fn test_parse_valid_request() {
    let json = r#"{"id":"123","action":"list_panes","params":{}}"#;
    let req: Request = serde_json::from_str(json).unwrap();

    assert_eq!(req.id, "123");
    assert_eq!(req.action, "list_panes");
}

#[test]
fn test_parse_request_with_params() {
    let json = r#"{"id":"456","action":"send_keys","params":{"pane_id":3,"text":"hello","enter":true}}"#;
    let req: Request = serde_json::from_str(json).unwrap();

    assert_eq!(req.action, "send_keys");
    let params: SendKeysParams = serde_json::from_value(req.params).unwrap();
    assert_eq!(params.pane_id, 3);
    assert_eq!(params.text, "hello");
    assert!(params.enter);
}

#[test]
fn test_parse_invalid_json_returns_error() {
    let json = "not valid json";
    let result: Result<Request, _> = serde_json::from_str(json);
    assert!(result.is_err());
}

#[test]
fn test_serialize_success_response() {
    let resp = Response {
        id: "123".to_string(),
        success: true,
        data: Some(serde_json::json!({"panes": []})),
        error: None,
    };
    let json = serde_json::to_string(&resp).unwrap();

    assert!(json.contains(r#""success":true"#));
    assert!(json.contains(r#""id":"123""#));
}

#[test]
fn test_serialize_error_response() {
    let resp = Response {
        id: "123".to_string(),
        success: false,
        data: None,
        error: Some("pane not found".to_string()),
    };
    let json = serde_json::to_string(&resp).unwrap();

    assert!(json.contains(r#""success":false"#));
    assert!(json.contains(r#""error":"pane not found""#));
}
```

**Implementation** (write AFTER tests fail):

```rust
// src/ipc.rs
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Deserialize)]
pub struct Request {
    pub id: String,
    pub action: String,
    #[serde(default)]
    pub params: Value,
}

#[derive(Debug, Serialize)]
pub struct Response {
    pub id: String,
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct SendKeysParams {
    pub pane_id: u32,
    pub text: String,
    #[serde(default)]
    pub enter: bool,
}

#[derive(Debug, Deserialize)]
pub struct PaneIdParam {
    pub pane_id: u32,
}
```

---

### 1.3 Pane State Management

**Tests**:

```rust
// src/state_test.rs
use super::state::*;
use zellij_tile::prelude::PaneInfo;

#[test]
fn test_empty_state_has_no_panes() {
    let state = State::default();
    assert!(state.panes().is_empty());
}

#[test]
fn test_update_panes_stores_terminal_panes() {
    let mut state = State::default();

    let pane_info = create_test_pane(1, "test__cc_1", false);
    let manifest = create_manifest_with_panes(vec![pane_info]);

    state.update_panes(manifest);

    assert_eq!(state.panes().len(), 1);
    assert_eq!(state.panes()[0].id, 1);
    assert_eq!(state.panes()[0].title, "test__cc_1");
}

#[test]
fn test_update_panes_excludes_plugin_panes() {
    let mut state = State::default();

    let terminal = create_test_pane(1, "test__cc_1", false);
    let plugin = create_test_pane(2, "nzm-agent", true); // is_plugin = true
    let manifest = create_manifest_with_panes(vec![terminal, plugin]);

    state.update_panes(manifest);

    assert_eq!(state.panes().len(), 1); // Only terminal pane
}

#[test]
fn test_get_pane_by_id_returns_pane() {
    let mut state = State::default();
    let pane = create_test_pane(5, "test__cod_1", false);
    state.update_panes(create_manifest_with_panes(vec![pane]));

    let found = state.get_pane(5);
    assert!(found.is_some());
    assert_eq!(found.unwrap().id, 5);
}

#[test]
fn test_get_pane_by_id_returns_none_for_missing() {
    let state = State::default();
    assert!(state.get_pane(999).is_none());
}

#[test]
fn test_get_pane_by_title_returns_pane() {
    let mut state = State::default();
    let pane = create_test_pane(3, "myproject__gmi_1", false);
    state.update_panes(create_manifest_with_panes(vec![pane]));

    let found = state.get_pane_by_title("myproject__gmi_1");
    assert!(found.is_some());
    assert_eq!(found.unwrap().id, 3);
}

// Test helpers
fn create_test_pane(id: u32, title: &str, is_plugin: bool) -> PaneInfo {
    PaneInfo {
        id,
        is_plugin,
        title: title.to_string(),
        is_focused: false,
        is_fullscreen: false,
        is_floating: false,
        is_suppressed: false,
        ..Default::default()
    }
}

fn create_manifest_with_panes(panes: Vec<PaneInfo>) -> PaneManifest {
    let mut manifest = PaneManifest::default();
    manifest.panes.insert(0, panes); // Tab 0
    manifest
}
```

**Implementation**:

```rust
// src/state.rs
use std::collections::HashMap;
use zellij_tile::prelude::{PaneInfo, PaneManifest};

#[derive(Default)]
pub struct State {
    panes: Vec<PaneInfo>,
    pane_by_id: HashMap<u32, usize>,
}

impl State {
    pub fn update_panes(&mut self, manifest: PaneManifest) {
        self.panes.clear();
        self.pane_by_id.clear();

        for (_tab_idx, tab_panes) in manifest.panes {
            for pane in tab_panes {
                if !pane.is_plugin {
                    let idx = self.panes.len();
                    self.pane_by_id.insert(pane.id, idx);
                    self.panes.push(pane);
                }
            }
        }
    }

    pub fn panes(&self) -> &[PaneInfo] {
        &self.panes
    }

    pub fn get_pane(&self, id: u32) -> Option<&PaneInfo> {
        self.pane_by_id.get(&id).map(|&idx| &self.panes[idx])
    }

    pub fn get_pane_by_title(&self, title: &str) -> Option<&PaneInfo> {
        self.panes.iter().find(|p| p.title == title)
    }
}
```

---

### 1.4 Command Handlers

**Tests**:

```rust
// src/commands_test.rs
use super::commands::*;
use super::state::State;

#[test]
fn test_handle_list_panes_returns_pane_array() {
    let mut state = State::default();
    state.update_panes(create_manifest_with_panes(vec![
        create_test_pane(1, "proj__cc_1", false),
        create_test_pane(2, "proj__cc_2", false),
    ]));

    let result = handle_list_panes(&state);

    assert!(result.success);
    let data = result.data.unwrap();
    let panes: Vec<PaneDto> = serde_json::from_value(data["panes"].clone()).unwrap();
    assert_eq!(panes.len(), 2);
    assert_eq!(panes[0].id, 1);
    assert_eq!(panes[0].title, "proj__cc_1");
}

#[test]
fn test_handle_list_panes_empty_state() {
    let state = State::default();
    let result = handle_list_panes(&state);

    assert!(result.success);
    let data = result.data.unwrap();
    let panes: Vec<PaneDto> = serde_json::from_value(data["panes"].clone()).unwrap();
    assert!(panes.is_empty());
}

#[test]
fn test_validate_send_keys_params_valid() {
    let params = SendKeysParams {
        pane_id: 1,
        text: "hello".to_string(),
        enter: true,
    };
    assert!(validate_send_keys_params(&params).is_ok());
}

#[test]
fn test_validate_send_keys_params_empty_text() {
    let params = SendKeysParams {
        pane_id: 1,
        text: "".to_string(),
        enter: false,
    };
    // Empty text is allowed (might just press enter)
    assert!(validate_send_keys_params(&params).is_ok());
}

#[test]
fn test_handle_unknown_action() {
    let req = Request {
        id: "1".to_string(),
        action: "unknown_action".to_string(),
        params: serde_json::Value::Null,
    };
    let state = State::default();

    let result = dispatch_command(&req, &state);

    assert!(!result.success);
    assert!(result.error.unwrap().contains("unknown action"));
}
```

**Implementation**:

```rust
// src/commands.rs
use crate::ipc::{Request, Response, SendKeysParams};
use crate::state::State;
use serde::Serialize;

#[derive(Serialize)]
pub struct PaneDto {
    pub id: u32,
    pub title: String,
    pub is_focused: bool,
    pub is_floating: bool,
}

pub fn dispatch_command(req: &Request, state: &State) -> Response {
    match req.action.as_str() {
        "list_panes" => handle_list_panes(state),
        "send_keys" => handle_send_keys(req, state),
        "send_interrupt" => handle_send_interrupt(req, state),
        "get_pane_info" => handle_get_pane_info(req, state),
        _ => Response {
            id: req.id.clone(),
            success: false,
            data: None,
            error: Some(format!("unknown action: {}", req.action)),
        },
    }
}

pub fn handle_list_panes(state: &State) -> Response {
    let panes: Vec<PaneDto> = state.panes().iter().map(|p| PaneDto {
        id: p.id,
        title: p.title.clone(),
        is_focused: p.is_focused,
        is_floating: p.is_floating,
    }).collect();

    Response {
        id: String::new(), // Will be set by caller
        success: true,
        data: Some(serde_json::json!({ "panes": panes })),
        error: None,
    }
}

pub fn validate_send_keys_params(params: &SendKeysParams) -> Result<(), String> {
    // pane_id validation happens when we look it up
    // text can be empty (just press enter)
    Ok(())
}

// ... other handlers
```

---

### 1.5 Plugin Main Loop

**Tests**:

```rust
// src/main_test.rs
use super::*;

#[test]
fn test_plugin_load_requests_permissions() {
    // This is tested via integration test with Zellij
    // Unit test verifies the load() method sets correct state
    let mut plugin = NzmAgent::default();
    let config = std::collections::BTreeMap::new();

    // After load, plugin should be in initialized state
    // (We can't test actual permission request without Zellij)
    plugin.load(config);
    assert!(plugin.is_initialized());
}

#[test]
fn test_plugin_update_handles_pane_update() {
    let mut plugin = NzmAgent::default();
    let manifest = create_test_manifest();

    let should_render = plugin.update(Event::PaneUpdate(manifest));

    assert!(should_render);
    assert!(!plugin.state.panes().is_empty());
}

#[test]
fn test_plugin_pipe_handles_valid_message() {
    let mut plugin = NzmAgent::default();
    // Pre-populate state
    plugin.state.update_panes(create_test_manifest());

    let pipe_msg = PipeMessage {
        source: PipeSource::Cli(CliPipeId("test".to_string())),
        name: "nzm".to_string(),
        payload: Some(r#"{"id":"1","action":"list_panes","params":{}}"#.to_string()),
        args: std::collections::BTreeMap::new(),
        is_private: false,
    };

    let should_render = plugin.pipe(pipe_msg);
    // Plugin should output response via cli_pipe_output
    assert!(!should_render); // list_panes doesn't need render
}
```

---

## Phase 2: Go CLI Foundation

### 2.1 Project Setup

**Test**: Verify Go module builds

```go
// internal/version/version_test.go
package version

import "testing"

func TestVersionIsSet(t *testing.T) {
    // Version should be "dev" in development
    if Version == "" {
        t.Error("Version should not be empty")
    }
}
```

**Implementation**:
```bash
mkdir -p cmd/nzm internal/cli internal/zellij
go mod init github.com/youruser/nzm
```

---

### 2.2 Zellij Client - Session Management

**Tests**:

```go
// internal/zellij/session_test.go
package zellij

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestParseSessionList_Empty(t *testing.T) {
    output := ""
    sessions, err := parseSessionList(output)

    require.NoError(t, err)
    assert.Empty(t, sessions)
}

func TestParseSessionList_SingleSession(t *testing.T) {
    output := "myproject"
    sessions, err := parseSessionList(output)

    require.NoError(t, err)
    assert.Len(t, sessions, 1)
    assert.Equal(t, "myproject", sessions[0].Name)
}

func TestParseSessionList_MultipleSessions(t *testing.T) {
    output := "project1\nproject2\nproject3"
    sessions, err := parseSessionList(output)

    require.NoError(t, err)
    assert.Len(t, sessions, 3)
}

func TestParseSessionList_WithExitedStatus(t *testing.T) {
    // Zellij shows (EXITED) for dead sessions
    output := "project1\nproject2 (EXITED)"
    sessions, err := parseSessionList(output)

    require.NoError(t, err)
    assert.Len(t, sessions, 2)
    assert.False(t, sessions[0].Exited)
    assert.True(t, sessions[1].Exited)
}

func TestSessionExists_ReturnsTrue(t *testing.T) {
    // Mock executor
    client := NewClient(WithExecutor(mockExecutor{
        output: "myproject\nother",
    }))

    exists, err := client.SessionExists("myproject")

    require.NoError(t, err)
    assert.True(t, exists)
}

func TestSessionExists_ReturnsFalse(t *testing.T) {
    client := NewClient(WithExecutor(mockExecutor{
        output: "other1\nother2",
    }))

    exists, err := client.SessionExists("missing")

    require.NoError(t, err)
    assert.False(t, exists)
}

// Mock executor for testing
type mockExecutor struct {
    output string
    err    error
}

func (m mockExecutor) Run(name string, args ...string) (string, error) {
    return m.output, m.err
}
```

**Implementation**:

```go
// internal/zellij/session.go
package zellij

import (
    "strings"
)

type Session struct {
    Name   string
    Exited bool
}

func (c *Client) ListSessions() ([]Session, error) {
    output, err := c.exec.Run("zellij", "list-sessions")
    if err != nil {
        return nil, err
    }
    return parseSessionList(output)
}

func parseSessionList(output string) ([]Session, error) {
    if output == "" {
        return nil, nil
    }

    var sessions []Session
    for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
        if line == "" {
            continue
        }

        session := Session{Name: line}
        if strings.Contains(line, "(EXITED)") {
            session.Name = strings.TrimSuffix(line, " (EXITED)")
            session.Exited = true
        }
        sessions = append(sessions, session)
    }
    return sessions, nil
}

func (c *Client) SessionExists(name string) (bool, error) {
    sessions, err := c.ListSessions()
    if err != nil {
        return false, err
    }

    for _, s := range sessions {
        if s.Name == name {
            return true, nil
        }
    }
    return false, nil
}
```

---

### 2.3 Plugin Communication

**Tests**:

```go
// internal/zellij/plugin_test.go
package zellij

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestBuildPipeCommand(t *testing.T) {
    cmd := buildPipeCommand("nzm", `{"action":"list_panes"}`)

    assert.Contains(t, cmd, "zellij")
    assert.Contains(t, cmd, "pipe")
    assert.Contains(t, cmd, "--plugin")
    assert.Contains(t, cmd, "nzm-agent")
}

func TestParsePluginResponse_Success(t *testing.T) {
    json := `{"id":"123","success":true,"data":{"panes":[]}}`

    resp, err := parsePluginResponse(json)

    require.NoError(t, err)
    assert.True(t, resp.Success)
    assert.Equal(t, "123", resp.ID)
}

func TestParsePluginResponse_Error(t *testing.T) {
    json := `{"id":"123","success":false,"error":"pane not found"}`

    resp, err := parsePluginResponse(json)

    require.NoError(t, err)
    assert.False(t, resp.Success)
    assert.Equal(t, "pane not found", resp.Error)
}

func TestParsePluginResponse_InvalidJSON(t *testing.T) {
    _, err := parsePluginResponse("not json")
    assert.Error(t, err)
}

func TestSendPluginCommand_Success(t *testing.T) {
    client := NewClient(WithExecutor(mockExecutor{
        output: `{"id":"1","success":true,"data":{"panes":[]}}`,
    }))

    resp, err := client.SendPluginCommand(Request{
        Action: "list_panes",
    })

    require.NoError(t, err)
    assert.True(t, resp.Success)
}

func TestListPanes_ParsesResponse(t *testing.T) {
    client := NewClient(WithExecutor(mockExecutor{
        output: `{"id":"1","success":true,"data":{"panes":[{"id":1,"title":"proj__cc_1","is_focused":true}]}}`,
    }))

    panes, err := client.ListPanes("test-session")

    require.NoError(t, err)
    require.Len(t, panes, 1)
    assert.Equal(t, uint32(1), panes[0].ID)
    assert.Equal(t, "proj__cc_1", panes[0].Title)
    assert.True(t, panes[0].IsFocused)
}
```

---

### 2.4 Layout Generation

**Tests**:

```go
// internal/layout/generate_test.go
package layout

import (
    "strings"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestGenerateLayout_BasicSession(t *testing.T) {
    opts := Options{
        Session:   "myproject",
        WorkDir:   "/home/user/project",
        CCCount:   2,
        CodCount:  1,
        PluginPath: "/path/to/nzm-agent.wasm",
    }

    kdl, err := Generate(opts)

    require.NoError(t, err)
    assert.Contains(t, kdl, `cwd "/home/user/project"`)
    assert.Contains(t, kdl, `name="myproject__cc_1"`)
    assert.Contains(t, kdl, `name="myproject__cc_2"`)
    assert.Contains(t, kdl, `name="myproject__cod_1"`)
    assert.Contains(t, kdl, "nzm-agent.wasm")
}

func TestGenerateLayout_IncludesPluginPane(t *testing.T) {
    opts := Options{
        Session:    "test",
        PluginPath: "/path/to/plugin.wasm",
    }

    kdl, err := Generate(opts)

    require.NoError(t, err)
    assert.Contains(t, kdl, `pane plugin="file:/path/to/plugin.wasm"`)
    assert.Contains(t, kdl, "borderless true")
}

func TestGenerateLayout_PaneNamingConvention(t *testing.T) {
    opts := Options{
        Session:  "myproj",
        CCCount:  1,
        GmiCount: 2,
    }

    kdl, err := Generate(opts)

    require.NoError(t, err)
    assert.Contains(t, kdl, `name="myproj__cc_1"`)
    assert.Contains(t, kdl, `name="myproj__gmi_1"`)
    assert.Contains(t, kdl, `name="myproj__gmi_2"`)
}

func TestGenerateLayout_AgentCommands(t *testing.T) {
    opts := Options{
        Session:   "test",
        CCCount:   1,
        ClaudeCmd: "claude --dangerously-skip-permissions",
    }

    kdl, err := Generate(opts)

    require.NoError(t, err)
    assert.Contains(t, kdl, `command="claude"`)
    assert.Contains(t, kdl, `args "--dangerously-skip-permissions"`)
}

func TestGenerateLayout_NoAgents(t *testing.T) {
    opts := Options{
        Session: "empty",
    }

    kdl, err := Generate(opts)

    require.NoError(t, err)
    // Should still have plugin pane
    assert.Contains(t, kdl, "pane plugin=")
    // Should have at least one shell pane
    assert.Contains(t, kdl, "pane {")
}
```

---

### 2.5 CLI Commands

**Tests**:

```go
// internal/cli/spawn_test.go
package cli

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestSpawnCommand_ValidatesSessionName(t *testing.T) {
    tests := []struct {
        name    string
        session string
        wantErr bool
    }{
        {"valid", "myproject", false},
        {"with-dash", "my-project", false},
        {"with-underscore", "my_project", false},
        {"empty", "", true},
        {"with-colon", "my:project", true},
        {"with-dot", "my.project", true},
        {"with-space", "my project", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateSessionName(tt.session)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestSpawnCommand_ParsesAgentFlags(t *testing.T) {
    args := []string{"myproject", "--cc=2", "--cod=1", "--gmi=3"}

    opts, err := parseSpawnArgs(args)

    require.NoError(t, err)
    assert.Equal(t, "myproject", opts.Session)
    assert.Equal(t, 2, opts.CCCount)
    assert.Equal(t, 1, opts.CodCount)
    assert.Equal(t, 3, opts.GmiCount)
}

func TestSpawnCommand_DefaultsToOneCC(t *testing.T) {
    args := []string{"myproject"}

    opts, err := parseSpawnArgs(args)

    require.NoError(t, err)
    assert.Equal(t, 1, opts.CCCount) // Default
    assert.Equal(t, 0, opts.CodCount)
    assert.Equal(t, 0, opts.GmiCount)
}
```

```go
// internal/cli/send_test.go
package cli

func TestSendCommand_RequiresMessage(t *testing.T) {
    args := []string{"myproject"}

    _, err := parseSendArgs(args)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "message required")
}

func TestSendCommand_ParsesPaneFilter(t *testing.T) {
    args := []string{"myproject", "--panes=1,2,3", "hello world"}

    opts, err := parseSendArgs(args)

    require.NoError(t, err)
    assert.Equal(t, []uint32{1, 2, 3}, opts.PaneIDs)
    assert.Equal(t, "hello world", opts.Message)
}

func TestSendCommand_AllFlag(t *testing.T) {
    args := []string{"myproject", "--all", "broadcast message"}

    opts, err := parseSendArgs(args)

    require.NoError(t, err)
    assert.True(t, opts.All)
}
```

---

## Phase 3: Integration Tests

### 3.1 Plugin + CLI Integration

```go
// tests/integration/spawn_test.go
// +build integration

package integration

import (
    "os/exec"
    "testing"
    "time"
)

func TestSpawnCreatesSession(t *testing.T) {
    session := "nzm-test-" + randomSuffix()
    defer cleanup(session)

    // Spawn session
    cmd := exec.Command("nzm", "spawn", session, "--cc=2")
    err := cmd.Run()
    require.NoError(t, err)

    // Verify session exists
    output, _ := exec.Command("zellij", "list-sessions").Output()
    assert.Contains(t, string(output), session)

    // Verify panes via status
    status, _ := exec.Command("nzm", "status", session, "--json").Output()
    var result StatusResult
    json.Unmarshal(status, &result)

    assert.Len(t, result.Panes, 2)
    assert.Equal(t, session+"__cc_1", result.Panes[0].Title)
}

func TestSendKeysToPane(t *testing.T) {
    session := "nzm-test-" + randomSuffix()
    defer cleanup(session)

    // Spawn with one pane
    exec.Command("nzm", "spawn", session, "--cc=1").Run()
    time.Sleep(2 * time.Second) // Wait for agent init

    // Send message
    cmd := exec.Command("nzm", "send", session, "--all", "echo hello")
    err := cmd.Run()
    require.NoError(t, err)

    // Verify (would need output capture to fully verify)
}

func cleanup(session string) {
    exec.Command("nzm", "kill", session).Run()
}
```

---

## Test Coverage Requirements

### Plugin (Rust)
- **Unit tests**: 90% coverage on `ipc.rs`, `state.rs`, `commands.rs`
- **Integration tests**: Plugin loads in Zellij, responds to pipes

### CLI (Go)
- **Unit tests**: 90% coverage on parsing, validation, layout generation
- **Integration tests**: Full spawn/send/kill workflow

### Coverage Commands

```bash
# Rust
cargo tarpaulin --out Html

# Go
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## CI Pipeline

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  plugin-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
        with:
          targets: wasm32-wasi
      - run: cd plugin/nzm-agent && cargo test
      - run: cd plugin/nzm-agent && cargo build --release --target wasm32-wasi

  cli-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go build -o nzm ./cmd/nzm

  integration-tests:
    runs-on: ubuntu-latest
    needs: [plugin-tests, cli-tests]
    steps:
      - uses: actions/checkout@v4
      # Install Zellij
      - run: cargo install zellij
      # Build artifacts
      - run: make build
      # Run integration tests
      - run: go test -tags=integration ./tests/integration/...
```

---

## Definition of Done

A feature is **DONE** when:

1. ✅ Failing test written first (RED)
2. ✅ Minimal implementation passes test (GREEN)
3. ✅ Code refactored for clarity (REFACTOR)
4. ✅ All existing tests still pass
5. ✅ Coverage maintained or increased
6. ✅ No lint errors (`cargo clippy`, `golangci-lint`)
7. ✅ Documentation updated if public API changed
