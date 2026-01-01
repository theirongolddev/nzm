package nzm

import (
	"context"
	"errors"
	"testing"

	"github.com/theirongolddev/nzm/internal/zellij"
)

// killMockClient for kill tests
type killMockClient struct {
	sessions      []zellij.Session
	listErr       error
	killErr       error
	killedSession string
}

func (m *killMockClient) ListSessions(ctx context.Context) ([]zellij.Session, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.sessions, nil
}

func (m *killMockClient) SessionExists(ctx context.Context, name string) (bool, error) {
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

func (m *killMockClient) KillSession(ctx context.Context, name string) error {
	m.killedSession = name
	return m.killErr
}

func TestKillOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    KillOptions
		wantErr bool
	}{
		{
			name: "valid session name",
			opts: KillOptions{
				Session: "myproj",
			},
			wantErr: false,
		},
		{
			name: "empty session name",
			opts: KillOptions{
				Session: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKill_Session(t *testing.T) {
	mock := &killMockClient{
		sessions: []zellij.Session{
			{Name: "proj"},
		},
	}
	killer := NewKiller(mock)

	err := killer.Kill(context.Background(), KillOptions{
		Session: "proj",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.killedSession != "proj" {
		t.Errorf("expected killed session 'proj', got %q", mock.killedSession)
	}
}

func TestKill_SessionNotFound(t *testing.T) {
	mock := &killMockClient{
		sessions: []zellij.Session{
			{Name: "other"},
		},
	}
	killer := NewKiller(mock)

	err := killer.Kill(context.Background(), KillOptions{
		Session: "nonexistent",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestKill_SessionNotFoundWithForce(t *testing.T) {
	mock := &killMockClient{
		sessions: []zellij.Session{},
	}
	killer := NewKiller(mock)

	// With force, should still try to kill
	err := killer.Kill(context.Background(), KillOptions{
		Session: "maybe-exists",
		Force:   true,
	})
	// Should succeed (or at least attempt kill) even if session not found
	if err != nil && mock.killedSession != "maybe-exists" {
		t.Errorf("force kill should attempt kill anyway: %v", err)
	}
}

func TestKill_ZellijError(t *testing.T) {
	mock := &killMockClient{
		sessions: []zellij.Session{
			{Name: "proj"},
		},
		killErr: errors.New("zellij kill failed"),
	}
	killer := NewKiller(mock)

	err := killer.Kill(context.Background(), KillOptions{
		Session: "proj",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestKill_ListError(t *testing.T) {
	mock := &killMockClient{
		listErr: errors.New("zellij not running"),
	}
	killer := NewKiller(mock)

	err := killer.Kill(context.Background(), KillOptions{
		Session: "proj",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestKill_All(t *testing.T) {
	killedSessions := make([]string, 0)
	mock := &killMockClient{
		sessions: []zellij.Session{
			{Name: "proj1"},
			{Name: "proj2"},
			{Name: "proj3"},
		},
	}
	// Override kill to track all killed sessions
	killer := NewKiller(&multiKillMock{
		sessions: mock.sessions,
		killed:   &killedSessions,
	})

	err := killer.KillAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(killedSessions) != 3 {
		t.Errorf("expected 3 killed sessions, got %d", len(killedSessions))
	}
}

type multiKillMock struct {
	sessions []zellij.Session
	killed   *[]string
	killErr  error
}

func (m *multiKillMock) ListSessions(ctx context.Context) ([]zellij.Session, error) {
	return m.sessions, nil
}

func (m *multiKillMock) SessionExists(ctx context.Context, name string) (bool, error) {
	for _, s := range m.sessions {
		if s.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *multiKillMock) KillSession(ctx context.Context, name string) error {
	*m.killed = append(*m.killed, name)
	return m.killErr
}
