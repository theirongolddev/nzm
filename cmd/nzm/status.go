package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/nzm"
	"github.com/Dicklesworthstone/ntm/internal/output"
	"github.com/Dicklesworthstone/ntm/internal/zellij"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [SESSION]",
	Short: "Show session and pane status",
	Long: `Show status of NZM sessions and their panes.

Without arguments, lists all sessions.
With a session name, shows detailed pane information.

Examples:
  # List all sessions
  nzm status

  # Show details for specific session
  nzm status myproj
  
  # Output as JSON
  nzm status --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	client := zellij.NewClient()
	status := nzm.NewStatus(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := nzm.StatusOptions{}
	if len(args) > 0 {
		opts.Session = args[0]
	}

	result, err := status.GetStatus(ctx, opts)
	if err != nil {
		return err
	}

	// Output formatting
	formatter := output.NZMDefaultFormatter(jsonFlag)

	if formatter.IsJSON() {
		return formatter.JSON(result)
	}

	// Text output
	if len(result.Sessions) == 0 {
		fmt.Println("No NZM sessions found.")
		return nil
	}

	for _, sess := range result.Sessions {
		// Session header
		statusStr := ""
		if sess.Attached {
			statusStr = " (attached)"
		}
		if sess.Exited {
			statusStr = " (exited)"
		}
		fmt.Printf("Session: %s%s\n", sess.Name, statusStr)

		// Agent counts
		if len(sess.AgentCounts) > 0 {
			fmt.Print("  Agents:")
			for agentType, count := range sess.AgentCounts {
				display := zellij.GetAgentTypeDisplay(agentType)
				fmt.Printf(" %s:%d", display, count)
			}
			fmt.Println()
		}

		// Pane details (if querying specific session)
		if opts.Session != "" && len(sess.Panes) > 0 {
			fmt.Println("  Panes:")
			for _, pane := range sess.Panes {
				focusStr := ""
				if pane.IsFocused {
					focusStr = " *"
				}
				fmt.Printf("    [%d] %s%s\n", pane.ID, pane.Title, focusStr)
			}
		}
		fmt.Println()
	}

	return nil
}
