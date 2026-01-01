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

var killCmd = &cobra.Command{
	Use:   "kill SESSION",
	Short: "Kill a session",
	Long: `Terminate an NZM session and all its panes.

Examples:
  # Kill a specific session
  nzm kill myproj

  # Force kill (don't check if session exists)
  nzm kill myproj --force

  # Kill all NZM sessions
  nzm kill --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: runKill,
}

var (
	killForce bool
	killAll   bool
)

func init() {
	rootCmd.AddCommand(killCmd)

	killCmd.Flags().BoolVarP(&killForce, "force", "f", false, "Force kill without checking session")
	killCmd.Flags().BoolVarP(&killAll, "all", "a", false, "Kill all sessions")
}

func runKill(cmd *cobra.Command, args []string) error {
	client := zellij.NewClient()
	killer := nzm.NewKiller(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	formatter := output.NZMDefaultFormatter(jsonFlag)

	if killAll {
		if err := killer.KillAll(ctx); err != nil {
			return err
		}
		if formatter.IsJSON() {
			return formatter.JSON(map[string]interface{}{
				"action":  "kill_all",
				"success": true,
			})
		}
		fmt.Println("All sessions killed.")
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("session name required (or use --all)")
	}

	session := args[0]
	opts := nzm.KillOptions{
		Session: session,
		Force:   killForce,
	}

	if err := killer.Kill(ctx, opts); err != nil {
		return err
	}

	if formatter.IsJSON() {
		return formatter.JSON(map[string]interface{}{
			"action":  "kill",
			"session": session,
			"success": true,
		})
	}

	fmt.Printf("Session %q killed.\n", session)
	return nil
}
