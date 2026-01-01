package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nzm",
	Short: "Named Zellij Manager - orchestrate AI coding agents",
	Long: `NZM (Named Zellij Manager) orchestrates multiple AI coding agents
in a Zellij terminal multiplexer session.

It allows you to:
  - Spawn sessions with multiple Claude, Codex, or Gemini panes
  - Send commands and text to specific panes
  - View status of active sessions and agents
  - Kill sessions when done`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
