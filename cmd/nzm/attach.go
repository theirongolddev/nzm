package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/zellij"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach SESSION",
	Short: "Attach to an existing session",
	Long: `Attach to an existing NZM session.

Examples:
  # Attach to a session
  nzm attach myproj`,
	Args: cobra.ExactArgs(1),
	RunE: runAttach,
}

func init() {
	rootCmd.AddCommand(attachCmd)
}

func runAttach(cmd *cobra.Command, args []string) error {
	session := args[0]

	client := zellij.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if session exists
	exists, err := client.SessionExists(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}
	if !exists {
		return fmt.Errorf("session %q not found", session)
	}

	// Attach (this replaces the current process)
	return client.AttachSession(context.Background(), session)
}
