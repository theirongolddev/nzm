package zellij

import (
	"context"
	"errors"
	"testing"
)

// mockExecutor implements Executor for testing
type mockExecutor struct {
	output string
	err    error
	calls  [][]string // record all calls
}

func (m *mockExecutor) Run(_ context.Context, args ...string) (string, error) {
	m.calls = append(m.calls, args)
	return m.output, m.err
}

func TestParseSessionList_Empty(t *testing.T) {
	sessions, err := parseSessionList("")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestParseSessionList_SingleSession(t *testing.T) {
	output := "myproject"
	sessions, err := parseSessionList(output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Name != "myproject" {
		t.Errorf("expected name 'myproject', got %q", sessions[0].Name)
	}
	if sessions[0].Exited {
		t.Error("expected Exited=false")
	}
}

func TestParseSessionList_MultipleSessions(t *testing.T) {
	output := "project1\nproject2\nproject3"
	sessions, err := parseSessionList(output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}

func TestParseSessionList_WithExitedStatus(t *testing.T) {
	// Zellij shows (EXITED) for dead sessions
	output := "project1\nproject2 (EXITED)"
	sessions, err := parseSessionList(output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if sessions[0].Exited {
		t.Error("expected first session Exited=false")
	}
	if !sessions[1].Exited {
		t.Error("expected second session Exited=true")
	}
	if sessions[1].Name != "project2" {
		t.Errorf("expected name 'project2', got %q", sessions[1].Name)
	}
}

func TestParseSessionList_WithAttachedStatus(t *testing.T) {
	// Zellij shows (current) for attached sessions
	output := "project1 (current)\nproject2"
	sessions, err := parseSessionList(output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	if !sessions[0].Attached {
		t.Error("expected first session Attached=true")
	}
	if sessions[0].Name != "project1" {
		t.Errorf("expected name 'project1', got %q", sessions[0].Name)
	}
}

func TestParseSessionList_SkipsEmptyLines(t *testing.T) {
	output := "project1\n\nproject2\n"
	sessions, err := parseSessionList(output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestClient_ListSessions(t *testing.T) {
	mock := &mockExecutor{output: "session1\nsession2"}
	client := NewClient(WithExecutor(mock))

	sessions, err := client.ListSessions(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	// Should call: zellij list-sessions
	if mock.calls[0][0] != "list-sessions" {
		t.Errorf("expected 'list-sessions', got %q", mock.calls[0][0])
	}
}

func TestClient_ListSessions_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("zellij not running")}
	client := NewClient(WithExecutor(mock))

	_, err := client.ListSessions(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClient_SessionExists_True(t *testing.T) {
	mock := &mockExecutor{output: "myproject\nother"}
	client := NewClient(WithExecutor(mock))

	exists, err := client.SessionExists(context.Background(), "myproject")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected session to exist")
	}
}

func TestClient_SessionExists_False(t *testing.T) {
	mock := &mockExecutor{output: "other1\nother2"}
	client := NewClient(WithExecutor(mock))

	exists, err := client.SessionExists(context.Background(), "missing")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected session to not exist")
	}
}

func TestClient_KillSession(t *testing.T) {
	mock := &mockExecutor{output: ""}
	client := NewClient(WithExecutor(mock))

	err := client.KillSession(context.Background(), "myproject")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	// Should call: zellij kill-session myproject
	expected := []string{"kill-session", "myproject"}
	for i, arg := range expected {
		if mock.calls[0][i] != arg {
			t.Errorf("expected arg %d to be %q, got %q", i, arg, mock.calls[0][i])
		}
	}
}

func TestClient_AttachSession(t *testing.T) {
	mock := &mockExecutor{output: ""}
	client := NewClient(WithExecutor(mock))

	err := client.AttachSession(context.Background(), "myproject")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	// Should call: zellij attach myproject
	expected := []string{"attach", "myproject"}
	for i, arg := range expected {
		if mock.calls[0][i] != arg {
			t.Errorf("expected arg %d to be %q, got %q", i, arg, mock.calls[0][i])
		}
	}
}

func TestClient_CreateSession(t *testing.T) {
	mock := &mockExecutor{output: ""}
	client := NewClient(WithExecutor(mock))

	err := client.CreateSession(context.Background(), "myproject", "/tmp/layout.kdl")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.calls))
	}
	// Should call: zellij --session myproject --layout /tmp/layout.kdl
	args := mock.calls[0]
	if args[0] != "--session" || args[1] != "myproject" {
		t.Errorf("expected '--session myproject', got %v", args[:2])
	}
	if args[2] != "--layout" || args[3] != "/tmp/layout.kdl" {
		t.Errorf("expected '--layout /tmp/layout.kdl', got %v", args[2:4])
	}
}

func TestValidateSessionName(t *testing.T) {
	tests := []struct {
		name    string
		session string
		wantErr bool
	}{
		{"valid", "myproject", false},
		{"with-dash", "my-project", false},
		{"with-underscore", "my_project", false},
		{"with-numbers", "project123", false},
		{"empty", "", true},
		{"with-colon", "my:project", true},
		{"with-dot", "my.project", true},
		{"with-space", "my project", true},
		{"with-slash", "my/project", true},
		{"starts-with-dash", "-project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionName(tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSessionName(%q) error = %v, wantErr %v", tt.session, err, tt.wantErr)
			}
		})
	}
}
