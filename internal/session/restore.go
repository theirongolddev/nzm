package session

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Dicklesworthstone/ntm/internal/zellij"
)

// Restore recreates a session from saved state.
func Restore(state *SessionState, opts RestoreOptions) error {
	name := opts.Name
	if name == "" {
		name = state.Name
	}

	// Check if session already exists
	if zellij.SessionExists(name) {
		if !opts.Force {
			return fmt.Errorf("session '%s' already exists (use --force to overwrite)", name)
		}
		if err := zellij.KillSession(name); err != nil {
			return fmt.Errorf("killing existing session: %w", err)
		}
	}

	// Validate and prepare working directory
	workDir := state.WorkDir
	if workDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}
		workDir = home
	}

	// Check if directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		// Try to create it if it looks like a project path
		if shouldCreateDir(workDir) {
			if err := os.MkdirAll(workDir, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", workDir, err)
			}
		} else {
			// Fall back to home directory
			home, _ := os.UserHomeDir()
			workDir = home
		}
	}

	// Create the session
	if err := zellij.CreateSession(name, workDir); err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	// Create additional panes
	totalPanes := len(state.Panes)
	if totalPanes > 1 {
		for i := 1; i < totalPanes; i++ {
			if _, err := zellij.SplitWindow(name, workDir); err != nil {
				return fmt.Errorf("creating pane %d: %w", i+1, err)
			}
		}
	}

	// Get pane list
	panes, err := zellij.GetPanes(name)
	if err != nil {
		return fmt.Errorf("getting panes: %w", err)
	}

	// Set pane titles
	for i, paneState := range state.Panes {
		if i >= len(panes) {
			break
		}
		if paneState.Title != "" {
			if err := zellij.SetPaneTitle(panes[i].ID, paneState.Title); err != nil {
				// Non-fatal - continue with other panes
				continue
			}
		}
	}

	// Apply layout
	if err := applyLayout(name, state.Layout); err != nil {
		// Non-fatal - tiled layout will be used
	}

	// Check git branch if requested
	if !opts.SkipGitCheck && state.GitBranch != "" {
		currentBranch := getCurrentGitBranch(workDir)
		if currentBranch != "" && currentBranch != state.GitBranch {
			// Just warn, don't fail
			log.Printf("restore: current branch '%s' differs from saved branch '%s'", currentBranch, state.GitBranch)
		}
	}

	return nil
}

// RestoreAgents launches the agents in the restored session.
// This is separated from Restore to allow for customization.
func RestoreAgents(sessionName string, state *SessionState, cmds AgentCommands) error {
	panes, err := zellij.GetPanes(sessionName)
	if err != nil {
		return fmt.Errorf("getting panes: %w", err)
	}

	for i, paneState := range state.Panes {
		if i >= len(panes) {
			break
		}

		// Skip user panes
		if paneState.AgentType == string(zellij.AgentUser) || paneState.AgentType == "user" {
			continue
		}

		// Get agent command based on type
		agentCmd := getAgentCommand(paneState.AgentType, cmds)
		if agentCmd == "" {
			continue
		}

		// Launch agent
		safeAgentCmd, err := zellij.SanitizePaneCommand(agentCmd)
		if err != nil {
			continue
		}

		cmd, err := zellij.BuildPaneCommand(state.WorkDir, safeAgentCmd)
		if err != nil {
			continue
		}

		if err := zellij.SendKeys(panes[i].ID, cmd, true); err != nil {
			// Non-fatal - continue with other agents
			continue
		}
	}

	return nil
}

// getAgentCommand returns the command for an agent type.
func getAgentCommand(agentType string, cmds AgentCommands) string {
	switch agentType {
	case "cc", "claude":
		return cmds.Claude
	case "cod", "codex":
		return cmds.Codex
	case "gmi", "gemini":
		return cmds.Gemini
	default:
		return ""
	}
}

// applyLayout applies a layout to the session.
// Zellij layouts are applied at session creation time via KDL files,
// not dynamically like tmux. This is a no-op for Zellij.
func applyLayout(session, layout string) error {
	// Zellij doesn't support dynamic layout changes like tmux.
	// Layouts are defined in KDL files at session creation.
	// This function is kept for API compatibility but does nothing.
	return nil
}

// getCurrentGitBranch returns the current git branch for a directory.
func getCurrentGitBranch(dir string) string {
	output, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// shouldCreateDir determines if a path should be auto-created.
func shouldCreateDir(path string) bool {
	// Don't create root or home-level directories
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Must be under home directory
	if !strings.HasPrefix(path, home) {
		return false
	}

	// Should be at least 2 levels deep from home
	// e.g., ~/Developer/project is ok, ~/project is not
	rel, err := filepath.Rel(home, path)
	if err != nil {
		return false
	}

	parts := strings.Split(rel, string(filepath.Separator))
	return len(parts) >= 2
}
