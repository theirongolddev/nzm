package nzm

import (
	"context"
	"fmt"

	"github.com/Dicklesworthstone/ntm/internal/zellij"
)

// KillClient defines the interface for kill operations
type KillClient interface {
	ListSessions(ctx context.Context) ([]zellij.Session, error)
	SessionExists(ctx context.Context, name string) (bool, error)
	KillSession(ctx context.Context, name string) error
}

// KillOptions configures the kill operation
type KillOptions struct {
	Session string // Session name to kill
	Force   bool   // Force kill even if session not found
}

// Validate checks if kill options are valid
func (o KillOptions) Validate() error {
	if o.Session == "" {
		return fmt.Errorf("session name is required")
	}
	return nil
}

// Killer kills NZM sessions
type Killer struct {
	client KillClient
}

// NewKiller creates a new Killer
func NewKiller(client KillClient) *Killer {
	return &Killer{client: client}
}

// Kill terminates a session
func (k *Killer) Kill(ctx context.Context, opts KillOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// Check if session exists (unless force)
	if !opts.Force {
		exists, err := k.client.SessionExists(ctx, opts.Session)
		if err != nil {
			return fmt.Errorf("failed to check session: %w", err)
		}
		if !exists {
			return fmt.Errorf("session %q not found", opts.Session)
		}
	}

	return k.client.KillSession(ctx, opts.Session)
}

// KillAll terminates all sessions
func (k *Killer) KillAll(ctx context.Context) error {
	sessions, err := k.client.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var lastErr error
	for _, sess := range sessions {
		if err := k.client.KillSession(ctx, sess.Name); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
