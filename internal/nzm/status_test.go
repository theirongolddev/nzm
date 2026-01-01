package nzm

import (
	"context"
	"errors"
	"testing"

	"github.com/Dicklesworthstone/ntm/internal/zellij"
)

// combinedMockClient implements both session and pane operations for status tests
type combinedMockClient struct {
	sessions  []zellij.Session
	panes     []zellij.PaneInfo
	listErr   error
	panesErr  error
}

func (m *combinedMockClient) ListSessions(ctx context.Context) ([]zellij.Session, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.sessions, nil
}

func (m *combinedMockClient) SessionExists(ctx context.Context, name string) (bool, error) {
	if m.listErr != nil {
		return false, m.listErr
	}
	for _, s := range m.sessions {
		if s.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *combinedMockClient) ListPanes(ctx context.Context, session string) ([]zellij.PaneInfo, error) {
	if m.panesErr != nil {
		return nil, m.panesErr
	}
	return m.panes, nil
}

func TestStatus_ListAllSessions(t *testing.T) {
	mock := &combinedMockClient{
		sessions: []zellij.Session{
			{Name: "proj1"},
			{Name: "proj2"},
		},
	}
	status := NewStatus(mock)

	result, err := status.GetStatus(context.Background(), StatusOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(result.Sessions))
	}
}

func TestStatus_SessionDetails(t *testing.T) {
	mock := &combinedMockClient{
		sessions: []zellij.Session{
			{Name: "proj", Attached: true},
		},
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1", IsFocused: true},
			{ID: 2, Title: "proj__cc_2"},
			{ID: 3, Title: "proj__gmi_1"},
		},
	}
	status := NewStatus(mock)

	result, err := status.GetStatus(context.Background(), StatusOptions{
		Session: "proj",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result.Sessions))
	}

	sess := result.Sessions[0]
	if sess.Name != "proj" {
		t.Errorf("expected name 'proj', got %q", sess.Name)
	}
	if !sess.Attached {
		t.Error("expected attached=true")
	}
	if len(sess.Panes) != 3 {
		t.Errorf("expected 3 panes, got %d", len(sess.Panes))
	}
}

func TestStatus_SessionNotFound(t *testing.T) {
	mock := &combinedMockClient{
		sessions: []zellij.Session{
			{Name: "other"},
		},
	}
	status := NewStatus(mock)

	_, err := status.GetStatus(context.Background(), StatusOptions{
		Session: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestStatus_ListSessionsError(t *testing.T) {
	mock := &combinedMockClient{
		listErr: errors.New("zellij not running"),
	}
	status := NewStatus(mock)

	_, err := status.GetStatus(context.Background(), StatusOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStatus_ListPanesError(t *testing.T) {
	mock := &combinedMockClient{
		sessions: []zellij.Session{
			{Name: "proj"},
		},
		panesErr: errors.New("plugin not responding"),
	}
	status := NewStatus(mock)

	// Should still return session info, just no panes
	result, err := status.GetStatus(context.Background(), StatusOptions{
		Session: "proj",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Sessions) != 1 {
		t.Fatalf("expected 1 session")
	}
	// Panes should be empty due to error
	if len(result.Sessions[0].Panes) != 0 {
		t.Errorf("expected 0 panes due to error, got %d", len(result.Sessions[0].Panes))
	}
}

func TestStatus_CountsByAgentType(t *testing.T) {
	mock := &combinedMockClient{
		sessions: []zellij.Session{
			{Name: "proj"},
		},
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
			{ID: 2, Title: "proj__cc_2"},
			{ID: 3, Title: "proj__gmi_1"},
			{ID: 4, Title: "proj__cod_1"},
		},
	}
	status := NewStatus(mock)

	result, err := status.GetStatus(context.Background(), StatusOptions{
		Session: "proj",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sess := result.Sessions[0]
	if sess.AgentCounts["cc"] != 2 {
		t.Errorf("expected 2 cc agents, got %d", sess.AgentCounts["cc"])
	}
	if sess.AgentCounts["gmi"] != 1 {
		t.Errorf("expected 1 gmi agent, got %d", sess.AgentCounts["gmi"])
	}
	if sess.AgentCounts["cod"] != 1 {
		t.Errorf("expected 1 cod agent, got %d", sess.AgentCounts["cod"])
	}
}

func TestStatus_SessionExitedFlag(t *testing.T) {
	mock := &combinedMockClient{
		sessions: []zellij.Session{
			{Name: "dead", Exited: true},
		},
	}
	status := NewStatus(mock)

	result, err := status.GetStatus(context.Background(), StatusOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Sessions[0].Exited {
		t.Error("expected exited=true")
	}
}
