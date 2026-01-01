package nzm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/theirongolddev/nzm/internal/zellij"
)

// ZellijClient defines the interface for Zellij session operations
type ZellijClient interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	SessionExists(ctx context.Context, name string) (bool, error)
	CreateSession(ctx context.Context, name, layoutPath string) error
	CreateSessionDetached(ctx context.Context, name, layoutPath string) error
	KillSession(ctx context.Context, name string) error
	AttachSession(ctx context.Context, name string) error
}

// SpawnOptions configures session creation
type SpawnOptions struct {
	Session     string // Session name
	WorkDir     string // Working directory
	PluginPath  string // Path to nzm-agent plugin
	CCCount     int    // Number of Claude panes
	CodCount    int    // Number of Codex panes
	GmiCount    int    // Number of Gemini panes
	IncludeUser bool   // Include a user pane
	ClaudeCmd   string // Command for Claude panes
	CodCmd      string // Command for Codex panes
	GmiCmd      string // Command for Gemini panes
	Detached    bool   // Create session in background
}

// Validate checks if spawn options are valid
func (o SpawnOptions) Validate() error {
	if o.Session == "" {
		return fmt.Errorf("session name is required")
	}
	if err := zellij.ValidateSessionName(o.Session); err != nil {
		return err
	}
	if o.CCCount == 0 && o.CodCount == 0 && o.GmiCount == 0 && !o.IncludeUser {
		return fmt.Errorf("at least one agent or user pane is required")
	}
	return nil
}

// SpawnResult contains information about the created session
type SpawnResult struct {
	Session    string
	WorkDir    string
	PaneCount  int
	LayoutPath string
}

// Spawner creates new NZM sessions
type Spawner struct {
	client ZellijClient
}

// NewSpawner creates a new Spawner
func NewSpawner(client ZellijClient) *Spawner {
	return &Spawner{client: client}
}

// Spawn creates a new session with the specified configuration
func (s *Spawner) Spawn(ctx context.Context, opts SpawnOptions) (*SpawnResult, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// Check if session already exists
	exists, err := s.client.SessionExists(ctx, opts.Session)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("session %q already exists", opts.Session)
	}

	// Default working directory to current directory
	workDir := opts.WorkDir
	if workDir == "" {
		workDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Generate layout
	layoutOpts := zellij.LayoutOptions{
		Session:     opts.Session,
		WorkDir:     workDir,
		PluginPath:  opts.PluginPath,
		CCCount:     opts.CCCount,
		CodCount:    opts.CodCount,
		GmiCount:    opts.GmiCount,
		IncludeUser: opts.IncludeUser,
		ClaudeCmd:   opts.ClaudeCmd,
		CodCmd:      opts.CodCmd,
		GmiCmd:      opts.GmiCmd,
	}

	layoutKDL, err := zellij.GenerateLayout(layoutOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate layout: %w", err)
	}

	// Write layout to temp file
	layoutPath := filepath.Join(os.TempDir(), fmt.Sprintf("nzm-%s.kdl", opts.Session))
	if err := os.WriteFile(layoutPath, []byte(layoutKDL), 0644); err != nil {
		return nil, fmt.Errorf("failed to write layout file: %w", err)
	}

	// Create the session
	var createErr error
	if opts.Detached {
		createErr = s.client.CreateSessionDetached(ctx, opts.Session, layoutPath)
	} else {
		createErr = s.client.CreateSession(ctx, opts.Session, layoutPath)
	}

	if createErr != nil {
		// Clean up layout file on error
		os.Remove(layoutPath)
		return nil, fmt.Errorf("failed to create session: %w", createErr)
	}

	// Calculate pane count
	paneCount := opts.CCCount + opts.CodCount + opts.GmiCount
	if opts.IncludeUser {
		paneCount++
	}

	return &SpawnResult{
		Session:    opts.Session,
		WorkDir:    workDir,
		PaneCount:  paneCount,
		LayoutPath: layoutPath,
	}, nil
}
