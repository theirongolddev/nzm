package main

import (
	"fmt"
	"os"

	"github.com/Dicklesworthstone/ntm/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	cfgFile  string
	jsonFlag bool

	// Global config (loaded once at startup)
	cfg *config.NZMConfig
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.NZMLoad(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.config/nzm/config.toml)")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in JSON format")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
