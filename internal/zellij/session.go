package zellij

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Session represents a Zellij session
type Session struct {
	Name     string
	Exited   bool
	Attached bool
}

// ListSessions returns all Zellij sessions
func (c *Client) ListSessions(ctx context.Context) ([]Session, error) {
	output, err := c.Run(ctx, "list-sessions")
	if err != nil {
		return nil, err
	}
	return parseSessionList(output)
}

// parseSessionList parses the output of `zellij list-sessions`
func parseSessionList(output string) ([]Session, error) {
	if output == "" {
		return nil, nil
	}

	var sessions []Session
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		session := Session{Name: line}

		// Check for (EXITED) suffix
		if strings.HasSuffix(line, " (EXITED)") {
			session.Name = strings.TrimSuffix(line, " (EXITED)")
			session.Exited = true
		}

		// Check for (current) suffix - indicates attached session
		if strings.HasSuffix(line, " (current)") {
			session.Name = strings.TrimSuffix(line, " (current)")
			session.Attached = true
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}

// SessionExists checks if a session with the given name exists
func (c *Client) SessionExists(ctx context.Context, name string) (bool, error) {
	sessions, err := c.ListSessions(ctx)
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

// KillSession terminates a session
func (c *Client) KillSession(ctx context.Context, name string) error {
	return c.RunSilent(ctx, "kill-session", name)
}

// AttachSession attaches to an existing session
func (c *Client) AttachSession(ctx context.Context, name string) error {
	return c.RunSilent(ctx, "attach", name)
}

// CreateSession creates a new session with a layout
func (c *Client) CreateSession(ctx context.Context, name, layoutPath string) error {
	return c.RunSilent(ctx, "--session", name, "--layout", layoutPath)
}

// CreateSessionDetached creates a new session in the background
func (c *Client) CreateSessionDetached(ctx context.Context, name, layoutPath string) error {
	// Use zellij action new-session for detached creation
	return c.RunSilent(ctx, "--session", name, "--layout", layoutPath, "--detached")
}

// validSessionNameRegex matches valid session names
// Must start with letter/number, can contain letters, numbers, dashes, underscores
var validSessionNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateSessionName checks if a session name is valid
func ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	if !validSessionNameRegex.MatchString(name) {
		return fmt.Errorf("invalid session name %q: must start with letter/number and contain only letters, numbers, dashes, underscores", name)
	}

	return nil
}
