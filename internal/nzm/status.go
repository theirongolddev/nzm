package nzm

import (
	"context"
	"fmt"

	"github.com/theirongolddev/nzm/internal/zellij"
)

// StatusClient defines the interface for status operations
type StatusClient interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	SessionExists(ctx context.Context, name string) (bool, error)
	ListPanes(ctx context.Context, session string) ([]zellij.PaneInfo, error)
}

// StatusOptions configures the status query
type StatusOptions struct {
	Session string // Optional: specific session to query
}

// SessionStatus contains status information for a session
type SessionStatus struct {
	Name        string
	Attached    bool
	Exited      bool
	Panes       []zellij.PaneInfo
	AgentCounts map[string]int
}

// StatusResult contains the status query results
type StatusResult struct {
	Sessions []SessionStatus
}

// Status queries session and pane status
type Status struct {
	client StatusClient
}

// NewStatus creates a new Status querier
func NewStatus(client StatusClient) *Status {
	return &Status{client: client}
}

// GetStatus retrieves status information
func (s *Status) GetStatus(ctx context.Context, opts StatusOptions) (*StatusResult, error) {
	sessions, err := s.client.ListSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	result := &StatusResult{
		Sessions: make([]SessionStatus, 0),
	}

	// If specific session requested, filter
	if opts.Session != "" {
		found := false
		for _, sess := range sessions {
			if sess.Name == opts.Session {
				found = true
				sessions = []zellij.Session{sess}
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("session %q not found", opts.Session)
		}
	}

	// Build status for each session
	for _, sess := range sessions {
		status := SessionStatus{
			Name:        sess.Name,
			Attached:    sess.Attached,
			Exited:      sess.Exited,
			Panes:       make([]zellij.PaneInfo, 0),
			AgentCounts: make(map[string]int),
		}

		// Get panes for this session (ignore errors - session might not have plugin)
		panes, err := s.client.ListPanes(ctx, sess.Name)
		if err == nil {
			status.Panes = panes
			// Count agents by type
			for _, pane := range panes {
				_, agentType, _, ok := zellij.ParsePaneName(pane.Title)
				if ok {
					status.AgentCounts[agentType]++
				}
			}
		}

		result.Sessions = append(result.Sessions, status)
	}

	return result, nil
}
