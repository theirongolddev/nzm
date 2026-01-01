package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/Dicklesworthstone/ntm/internal/zellij"
	"github.com/Dicklesworthstone/ntm/internal/tui/theme"
)

func newViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "view [session-name]",
		Aliases: []string{"v", "tile"},
		Short:   "View all panes in a session (unzoom, tile, attach)",
		Long: `View all panes in a Zellij session by:
1. Unzooming any zoomed panes
2. Applying tiled layout to all windows
3. Attaching/switching to the session

If no session is specified:
- If inside Zellij, operates on the current session
- Otherwise, shows a session selector

Examples:
  ntm view myproject
  ntm view                 # Select session or use current
  ntm tile myproject       # Alias`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var session string
			if len(args) > 0 {
				session = args[0]
			}
			return runView(cmd.OutOrStdout(), session)
		},
	}
}

func runView(w io.Writer, session string) error {
	if err := zellij.EnsureInstalled(); err != nil {
		return err
	}

	t := theme.Current()

	res, err := ResolveSession(session, w)
	if err != nil {
		return err
	}
	if res.Session == "" {
		return nil
	}
	res.ExplainIfInferred(os.Stderr)
	session = res.Session

	if !zellij.SessionExists(session) {
		return fmt.Errorf("session '%s' not found", session)
	}

	// Apply tiled layout (includes unzoom)
	if err := zellij.ApplyTiledLayout(session); err != nil {
		return fmt.Errorf("failed to apply layout: %w", err)
	}

	fmt.Printf("%s✓%s Tiled layout applied to '%s'\n",
		colorize(t.Success), colorize(t.Text), session)

	// Attach or switch to session
	return zellij.AttachOrSwitch(session)
}
