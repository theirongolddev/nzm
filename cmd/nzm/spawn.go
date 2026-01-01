package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/nzm"
	"github.com/Dicklesworthstone/ntm/internal/output"
	"github.com/Dicklesworthstone/ntm/internal/zellij"
	"github.com/spf13/cobra"
)

var spawnCmd = &cobra.Command{
	Use:   "spawn SESSION",
	Short: "Create a new session with AI agent panes",
	Long: `Create a new Zellij session with the specified AI agent panes.

Examples:
  # Create session with 2 Claude panes
  nzm spawn myproject --cc 2

  # Create session with Claude and Gemini panes
  nzm spawn proj --cc 2 --gmi 1

  # Create session with all agent types and a user pane
  nzm spawn proj --cc 2 --cod 1 --gmi 1 --user

  # Create session in background (detached)
  nzm spawn proj --cc 2 --detached`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing required argument: SESSION_NAME\n\nUsage: nzm spawn <session-name> [flags]\n\nExample: nzm spawn myproject --cc=2")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments: expected 1 session name, got %d", len(args))
		}
		return nil
	},
	RunE: runSpawn,
}

var (
	spawnCCCount     int
	spawnCodCount    int
	spawnGmiCount    int
	spawnIncludeUser bool
	spawnWorkDir     string
	spawnPluginPath  string
	spawnClaudeCmd   string
	spawnCodCmd      string
	spawnGmiCmd      string
	spawnDetached    bool
)

func init() {
	rootCmd.AddCommand(spawnCmd)

	spawnCmd.Flags().IntVar(&spawnCCCount, "cc", 0, "Number of Claude panes")
	spawnCmd.Flags().IntVar(&spawnCodCount, "cod", 0, "Number of Codex panes")
	spawnCmd.Flags().IntVar(&spawnGmiCount, "gmi", 0, "Number of Gemini panes")
	spawnCmd.Flags().BoolVar(&spawnIncludeUser, "user", false, "Include a user shell pane")
	spawnCmd.Flags().StringVar(&spawnWorkDir, "workdir", "", "Working directory for panes (default: current directory)")
	spawnCmd.Flags().StringVar(&spawnPluginPath, "plugin", "", "Path to nzm-agent plugin")
	spawnCmd.Flags().StringVar(&spawnClaudeCmd, "claude-cmd", "", "Command to run in Claude panes")
	spawnCmd.Flags().StringVar(&spawnCodCmd, "cod-cmd", "", "Command to run in Codex panes")
	spawnCmd.Flags().StringVar(&spawnGmiCmd, "gmi-cmd", "", "Command to run in Gemini panes")
	spawnCmd.Flags().BoolVarP(&spawnDetached, "detached", "d", false, "Create session in background")
}

func runSpawn(cmd *cobra.Command, args []string) error {
	session := args[0]

	client := zellij.NewClient()
	spawner := nzm.NewSpawner(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use config defaults if flags not set
	workDir := spawnWorkDir
	if workDir == "" && cfg != nil {
		workDir = cfg.GetProjectDir(session)
	}

	pluginPath := spawnPluginPath
	if pluginPath == "" && cfg != nil {
		pluginPath = cfg.Zellij.PluginPath
	}

	claudeCmd := spawnClaudeCmd
	if claudeCmd == "" && cfg != nil {
		claudeCmd = cfg.Agents.Claude
	}

	codCmd := spawnCodCmd
	if codCmd == "" && cfg != nil {
		codCmd = cfg.Agents.Codex
	}

	gmiCmd := spawnGmiCmd
	if gmiCmd == "" && cfg != nil {
		gmiCmd = cfg.Agents.Gemini
	}

	opts := nzm.SpawnOptions{
		Session:     session,
		WorkDir:     workDir,
		PluginPath:  pluginPath,
		CCCount:     spawnCCCount,
		CodCount:    spawnCodCount,
		GmiCount:    spawnGmiCount,
		IncludeUser: spawnIncludeUser,
		ClaudeCmd:   claudeCmd,
		CodCmd:      codCmd,
		GmiCmd:      gmiCmd,
		Detached:    spawnDetached,
	}

	result, err := spawner.Spawn(ctx, opts)
	if err != nil {
		return err
	}

	// Output formatting
	formatter := output.NZMDefaultFormatter(jsonFlag)

	if formatter.IsJSON() {
		return formatter.JSON(map[string]interface{}{
			"session":    result.Session,
			"pane_count": result.PaneCount,
			"layout":     result.LayoutPath,
			"workdir":    result.WorkDir,
			"detached":   spawnDetached,
		})
	}

	if spawnDetached {
		fmt.Printf("Created session %q with %d panes (detached)\n", result.Session, result.PaneCount)
		fmt.Printf("  Layout: %s\n", result.LayoutPath)
		fmt.Printf("  WorkDir: %s\n", result.WorkDir)
		fmt.Printf("\nAttach with: nzm attach %s\n", result.Session)
	} else {
		// When attached, we don't print anything as we're inside Zellij
		os.Exit(0)
	}

	return nil
}
