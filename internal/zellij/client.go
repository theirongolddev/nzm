// Package zellij provides a client for interacting with Zellij terminal multiplexer.
package zellij

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Executor runs commands and returns output
type Executor interface {
	Run(ctx context.Context, args ...string) (string, error)
}

// realExecutor executes actual zellij commands
type realExecutor struct{}

func (e *realExecutor) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "zellij", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", fmt.Errorf("zellij %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Client handles Zellij operations
type Client struct {
	exec   Executor
	Remote string // For API compatibility - not actually used by Zellij
}

// ClientOption configures a Client
type ClientOption func(*Client)

// WithExecutor sets a custom executor (useful for testing)
func WithExecutor(exec Executor) ClientOption {
	return func(c *Client) {
		c.exec = exec
	}
}

// WithRemote sets the remote host for SSH-based sessions.
// Note: Zellij doesn't natively support remote like tmux -L,
// but this is kept for API compatibility.
func WithRemote(host string) ClientOption {
	return func(c *Client) {
		c.Remote = host
	}
}

// NewClient creates a new Zellij client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		exec: &realExecutor{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// DefaultClient is the default Zellij client
var DefaultClient = NewClient()

// Run executes a zellij command with context
func (c *Client) Run(ctx context.Context, args ...string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return c.exec.Run(ctx, args...)
}

// RunContext is an alias for Run (for tmux API compatibility)
func (c *Client) RunContext(ctx context.Context, args ...string) (string, error) {
	return c.Run(ctx, args...)
}

// RunSilent executes a zellij command ignoring output
func (c *Client) RunSilent(ctx context.Context, args ...string) error {
	_, err := c.Run(ctx, args...)
	return err
}

// RunSilentContext is an alias for RunSilent (for tmux API compatibility)
func (c *Client) RunSilentContext(ctx context.Context, args ...string) error {
	return c.RunSilent(ctx, args...)
}

// IsInstalled checks if zellij is available
func (c *Client) IsInstalled() bool {
	_, err := exec.LookPath("zellij")
	return err == nil
}

// GetRemote returns the remote host (always empty for Zellij - not supported)
func (c *Client) GetRemote() string {
	return c.Remote
}

// ============== Non-context convenience methods for tmux API compatibility ==============

// RunSimple executes a zellij command without context (uses Background)
func (c *Client) RunSimple(args ...string) (string, error) {
	return c.Run(context.Background(), args...)
}

// RunSilentSimple executes a zellij command without context, ignoring output
func (c *Client) RunSilentSimple(args ...string) error {
	return c.RunSilent(context.Background(), args...)
}
