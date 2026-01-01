package nzm

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/theirongolddev/nzm/internal/zellij"
)

// mockZellijClient implements the ZellijClient interface for testing
type mockZellijClient struct {
	sessions       []zellij.Session
	sessionExists  bool
	createErr      error
	listErr        error
	killErr        error
	createdSession string
	createdLayout  string
}

func (m *mockZellijClient) ListSessions(ctx context.Context) ([]zellij.Session, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.sessions, nil
}

func (m *mockZellijClient) SessionExists(ctx context.Context, name string) (bool, error) {
	if m.listErr != nil {
		return false, m.listErr
	}
	return m.sessionExists, nil
}

func (m *mockZellijClient) CreateSession(ctx context.Context, name, layoutPath string) error {
	m.createdSession = name
	m.createdLayout = layoutPath
	return m.createErr
}

func (m *mockZellijClient) CreateSessionDetached(ctx context.Context, name, layoutPath string) error {
	m.createdSession = name
	m.createdLayout = layoutPath
	return m.createErr
}

func (m *mockZellijClient) KillSession(ctx context.Context, name string) error {
	return m.killErr
}

func (m *mockZellijClient) AttachSession(ctx context.Context, name string) error {
	return nil
}

func TestSpawnOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    SpawnOptions
		wantErr bool
	}{
		{
			name: "valid with session and agents",
			opts: SpawnOptions{
				Session: "myproj",
				CCCount: 2,
			},
			wantErr: false,
		},
		{
			name: "empty session name",
			opts: SpawnOptions{
				Session: "",
				CCCount: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid session name",
			opts: SpawnOptions{
				Session: "invalid name!",
				CCCount: 1,
			},
			wantErr: true,
		},
		{
			name: "no agents specified",
			opts: SpawnOptions{
				Session: "test",
			},
			wantErr: true,
		},
		{
			name: "valid with user pane only",
			opts: SpawnOptions{
				Session:     "test",
				IncludeUser: true,
			},
			wantErr: false,
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

func TestSpawn_CreatesSession(t *testing.T) {
	mock := &mockZellijClient{}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session:    "test-proj",
		WorkDir:    "/tmp/test",
		CCCount:    2,
		PluginPath: "/path/to/plugin.wasm",
		Detached:   true,
	}

	result, err := spawner.Spawn(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Session != "test-proj" {
		t.Errorf("expected session 'test-proj', got %q", result.Session)
	}
	if result.PaneCount != 2 {
		t.Errorf("expected 2 panes, got %d", result.PaneCount)
	}
	if mock.createdSession != "test-proj" {
		t.Errorf("expected created session 'test-proj', got %q", mock.createdSession)
	}
}

func TestSpawn_WritesLayout(t *testing.T) {
	mock := &mockZellijClient{}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session: "layout-test",
		CCCount: 1,
		GmiCount: 1,
		Detached: true,
	}

	_, err := spawner.Spawn(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have written a layout file
	if mock.createdLayout == "" {
		t.Fatal("expected layout path to be set")
	}

	// Check layout file exists
	if _, err := os.Stat(mock.createdLayout); os.IsNotExist(err) {
		t.Errorf("layout file should exist at %s", mock.createdLayout)
	}

	// Read layout and verify content
	content, err := os.ReadFile(mock.createdLayout)
	if err != nil {
		t.Fatalf("failed to read layout: %v", err)
	}

	layoutStr := string(content)
	if !strings.Contains(layoutStr, `name="layout-test__cc_1"`) {
		t.Error("layout should contain cc pane")
	}
	if !strings.Contains(layoutStr, `name="layout-test__gmi_1"`) {
		t.Error("layout should contain gmi pane")
	}
}

func TestSpawn_SessionAlreadyExists(t *testing.T) {
	mock := &mockZellijClient{
		sessionExists: true,
	}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session: "existing",
		CCCount: 1,
	}

	_, err := spawner.Spawn(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for existing session")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestSpawn_CreateSessionError(t *testing.T) {
	mock := &mockZellijClient{
		createErr: errors.New("zellij failed"),
	}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session: "fail-test",
		CCCount: 1,
		Detached: true,
	}

	_, err := spawner.Spawn(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSpawn_DefaultWorkDir(t *testing.T) {
	mock := &mockZellijClient{}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session:  "workdir-test",
		CCCount:  1,
		Detached: true,
	}

	result, err := spawner.Spawn(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use current directory
	cwd, _ := os.Getwd()
	if result.WorkDir != cwd {
		t.Errorf("expected workdir %q, got %q", cwd, result.WorkDir)
	}
}

func TestSpawn_CleansUpLayoutOnError(t *testing.T) {
	mock := &mockZellijClient{
		createErr: errors.New("session create failed"),
	}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session:  "cleanup-test",
		CCCount:  1,
		Detached: true,
	}

	_, _ = spawner.Spawn(context.Background(), opts)

	// Layout file should be cleaned up on error
	if mock.createdLayout != "" {
		if _, err := os.Stat(mock.createdLayout); err == nil {
			t.Error("layout file should be cleaned up on error")
		}
	}
}

func TestSpawnResult_LayoutPath(t *testing.T) {
	mock := &mockZellijClient{}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session:  "path-test",
		CCCount:  1,
		Detached: true,
	}

	result, err := spawner.Spawn(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Layout should be in temp directory
	if !strings.HasPrefix(result.LayoutPath, os.TempDir()) {
		// Also check if path contains "nzm" somewhere
		if !strings.Contains(filepath.Dir(result.LayoutPath), "nzm") {
			t.Errorf("layout path should be in temp or nzm directory: %s", result.LayoutPath)
		}
	}
}

func TestSpawn_WithAgentCommands(t *testing.T) {
	mock := &mockZellijClient{}
	spawner := NewSpawner(mock)

	opts := SpawnOptions{
		Session:   "cmd-test",
		CCCount:   1,
		ClaudeCmd: "claude --dangerously-skip-permissions",
		Detached:  true,
	}

	_, err := spawner.Spawn(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify layout contains command
	content, err := os.ReadFile(mock.createdLayout)
	if err != nil {
		t.Fatalf("failed to read layout: %v", err)
	}

	if !strings.Contains(string(content), `command "claude"`) {
		t.Error("layout should contain command clause")
	}
}
