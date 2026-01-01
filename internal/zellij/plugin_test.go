package zellij

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestRequest_Serialize(t *testing.T) {
	req := Request{
		ID:     "123",
		Action: "list_panes",
		Params: nil,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(data)
	if jsonStr != `{"id":"123","action":"list_panes"}` {
		t.Errorf("unexpected JSON: %s", jsonStr)
	}
}

func TestRequest_SerializeWithParams(t *testing.T) {
	req := Request{
		ID:     "456",
		Action: "send_keys",
		Params: map[string]any{
			"pane_id": 1,
			"text":    "hello",
			"enter":   true,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Verify it can be round-tripped
	var parsed Request
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed.ID != "456" {
		t.Errorf("expected ID '456', got %q", parsed.ID)
	}
	if parsed.Action != "send_keys" {
		t.Errorf("expected action 'send_keys', got %q", parsed.Action)
	}
}

func TestParseResponse_Success(t *testing.T) {
	jsonStr := `{"id":"123","success":true,"data":{"panes":[]}}`

	resp, err := ParseResponse(jsonStr)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.ID != "123" {
		t.Errorf("expected ID '123', got %q", resp.ID)
	}
	if resp.Error != "" {
		t.Errorf("expected no error, got %q", resp.Error)
	}
}

func TestParseResponse_Error(t *testing.T) {
	jsonStr := `{"id":"123","success":false,"error":"pane not found"}`

	resp, err := ParseResponse(jsonStr)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "pane not found" {
		t.Errorf("expected error 'pane not found', got %q", resp.Error)
	}
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	_, err := ParseResponse("not json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPaneInfo_Parse(t *testing.T) {
	jsonStr := `{"id":"1","success":true,"data":{"panes":[{"id":1,"title":"proj__cc_1","is_focused":true,"is_floating":false}]}}`

	resp, err := ParseResponse(jsonStr)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	panes, err := resp.GetPanes()
	if err != nil {
		t.Fatalf("failed to get panes: %v", err)
	}

	if len(panes) != 1 {
		t.Fatalf("expected 1 pane, got %d", len(panes))
	}
	if panes[0].ID != 1 {
		t.Errorf("expected ID 1, got %d", panes[0].ID)
	}
	if panes[0].Title != "proj__cc_1" {
		t.Errorf("expected title 'proj__cc_1', got %q", panes[0].Title)
	}
	if !panes[0].IsFocused {
		t.Error("expected is_focused=true")
	}
}

func TestClient_SendPluginCommand(t *testing.T) {
	mock := &mockExecutor{output: `{"id":"1","success":true,"data":{"panes":[]}}`}
	client := NewClient(WithExecutor(mock))

	req := Request{
		ID:     "1",
		Action: "list_panes",
	}

	resp, err := client.SendPluginCommand(context.Background(), "nzm-test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}

	// Verify the command was called correctly
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}

	// Should call: zellij --session nzm-test pipe --plugin nzm-agent -- <json>
	args := mock.calls[0]
	if args[0] != "--session" {
		t.Errorf("expected first arg '--session', got %q", args[0])
	}
}

func TestClient_SendPluginCommand_PluginError(t *testing.T) {
	mock := &mockExecutor{output: `{"id":"1","success":false,"error":"pane not found"}`}
	client := NewClient(WithExecutor(mock))

	req := Request{
		ID:     "1",
		Action: "get_pane_info",
		Params: map[string]any{"pane_id": 999},
	}

	resp, err := client.SendPluginCommand(context.Background(), "nzm-test", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Success {
		t.Error("expected success=false")
	}
	if resp.Error != "pane not found" {
		t.Errorf("expected error 'pane not found', got %q", resp.Error)
	}
}

func TestClient_SendPluginCommand_ExecutionError(t *testing.T) {
	mock := &mockExecutor{err: errors.New("zellij not running")}
	client := NewClient(WithExecutor(mock))

	req := Request{
		ID:     "1",
		Action: "list_panes",
	}

	_, err := client.SendPluginCommand(context.Background(), "nzm-test", req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_ListPanes(t *testing.T) {
	mock := &mockExecutor{output: `{"id":"1","success":true,"data":{"panes":[{"id":1,"title":"proj__cc_1","is_focused":true,"is_floating":false},{"id":2,"title":"proj__cc_2","is_focused":false,"is_floating":false}]}}`}
	client := NewClient(WithExecutor(mock))

	panes, err := client.ListPanes(context.Background(), "test-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(panes) != 2 {
		t.Fatalf("expected 2 panes, got %d", len(panes))
	}
	if panes[0].Title != "proj__cc_1" {
		t.Errorf("expected title 'proj__cc_1', got %q", panes[0].Title)
	}
}

func TestClient_SendKeys(t *testing.T) {
	mock := &mockExecutor{output: `{"id":"1","success":true,"data":{"action":"send_keys","pane_id":1,"text":"hello","enter":true}}`}
	client := NewClient(WithExecutor(mock))

	err := client.SendKeys(context.Background(), "test-session", 1, "hello", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SendKeys_PaneNotFound(t *testing.T) {
	mock := &mockExecutor{output: `{"id":"1","success":false,"error":"pane not found: 999"}`}
	client := NewClient(WithExecutor(mock))

	err := client.SendKeys(context.Background(), "test-session", 999, "hello", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "pane not found: 999" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_SendInterrupt(t *testing.T) {
	mock := &mockExecutor{output: `{"id":"1","success":true,"data":{"action":"send_interrupt","pane_id":1}}`}
	client := NewClient(WithExecutor(mock))

	err := client.SendInterrupt(context.Background(), "test-session", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := GenerateRequestID()
	id2 := GenerateRequestID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}
