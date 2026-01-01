package nzm

import (
	"context"
	"fmt"
	"strings"

	"github.com/theirongolddev/nzm/internal/zellij"
)

// PluginClient defines the interface for plugin operations
type PluginClient interface {
	ListPanes(ctx context.Context, session string) ([]zellij.PaneInfo, error)
	SendKeys(ctx context.Context, session string, paneID uint32, text string, enter bool) error
	SendInterrupt(ctx context.Context, session string, paneID uint32) error
}

// SendOptions configures the send operation
type SendOptions struct {
	Session   string // Session name
	Target    string // Target pane (name, type, or full pane name)
	Text      string // Text to send
	Enter     bool   // Press enter after text
	Interrupt bool   // Send Ctrl+C instead of text
}

// Validate checks if send options are valid
func (o SendOptions) Validate() error {
	if o.Session == "" {
		return fmt.Errorf("session name is required")
	}
	if o.Target == "" {
		return fmt.Errorf("target pane is required")
	}
	if !o.Interrupt && o.Text == "" {
		return fmt.Errorf("text is required (or use --interrupt)")
	}
	return nil
}

// Sender sends text to panes
type Sender struct {
	client PluginClient
}

// NewSender creates a new Sender
func NewSender(client PluginClient) *Sender {
	return &Sender{client: client}
}

// Send sends text or interrupt to a pane
func (s *Sender) Send(ctx context.Context, opts SendOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}

	// List panes to find target
	panes, err := s.client.ListPanes(ctx, opts.Session)
	if err != nil {
		return fmt.Errorf("failed to list panes: %w", err)
	}

	// Find matching pane
	pane, err := findPane(panes, opts.Session, opts.Target)
	if err != nil {
		return err
	}

	// Send interrupt or text
	if opts.Interrupt {
		return s.client.SendInterrupt(ctx, opts.Session, pane.ID)
	}
	return s.client.SendKeys(ctx, opts.Session, pane.ID, opts.Text, opts.Enter)
}

// findPane finds a pane by target string
// Target can be:
// - Full pane name: "proj__cc_1"
// - Short name: "cc_1"
// - Agent type: "cc" (matches first cc pane)
func findPane(panes []zellij.PaneInfo, session, target string) (*zellij.PaneInfo, error) {
	// Build expected full name for short targets
	fullTarget := target
	if !strings.Contains(target, "__") {
		// Could be "cc_1" or just "cc"
		if strings.Contains(target, "_") {
			// Looks like "cc_1", prefix with session
			fullTarget = fmt.Sprintf("%s__%s", session, target)
		}
	}

	// First pass: exact match on title or full target
	for i := range panes {
		if panes[i].Title == target || panes[i].Title == fullTarget {
			return &panes[i], nil
		}
	}

	// Second pass: match by agent type prefix
	// e.g., target "cc" should match first "proj__cc_N"
	agentPrefix := fmt.Sprintf("%s__%s_", session, target)
	for i := range panes {
		if strings.HasPrefix(panes[i].Title, agentPrefix) {
			return &panes[i], nil
		}
	}

	return nil, fmt.Errorf("pane %q not found in session %q", target, session)
}
