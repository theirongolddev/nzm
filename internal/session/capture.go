package session

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/zellij"
)

// Capture captures the current state of a tmux session.
func Capture(sessionName string) (*SessionState, error) {
	session, err := zellij.GetSession(sessionName)
	if err != nil {
		return nil, err
	}

	panes, err := zellij.GetPanes(sessionName)
	if err != nil {
		return nil, err
	}

	// Count agents by type
	agents := countAgents(panes)

	// Map pane states
	paneStates := mapPaneStates(panes)

	// Detect working directory from first pane or session
	cwd := detectWorkDir(sessionName, panes)

	// Get git info if in a repo
	gitBranch, gitRemote, gitCommit := getGitInfo(cwd)

	// Get layout
	layout := getLayout(sessionName)

	// Parse session creation time (tmux format varies, try common formats)
	var createdAt time.Time
	if session.Created != "" {
		// Try parsing various tmux date formats
		formats := []string{
			"Mon Jan 2 15:04:05 2006",
			"Mon Jan _2 15:04:05 2006",
			time.UnixDate,
			time.ANSIC,
		}
		for _, format := range formats {
			if t, err := time.Parse(format, session.Created); err == nil {
				createdAt = t.UTC()
				break
			}
		}
	}

	state := &SessionState{
		Name:      sessionName,
		SavedAt:   time.Now().UTC(),
		WorkDir:   cwd,
		GitBranch: gitBranch,
		GitRemote: gitRemote,
		GitCommit: gitCommit,
		Agents:    agents,
		Panes:     paneStates,
		Layout:    layout,
		CreatedAt: createdAt,
		Version:   StateVersion,
	}

	return state, nil
}

// countAgents counts agents by type from pane list.
func countAgents(panes []zellij.Pane) AgentConfig {
	config := AgentConfig{}
	for _, p := range panes {
		switch p.Type {
		case zellij.AgentClaude:
			config.Claude++
		case zellij.AgentCodex:
			config.Codex++
		case zellij.AgentGemini:
			config.Gemini++
		case zellij.AgentUser:
			config.User++
		}
	}
	return config
}

// mapPaneStates converts tmux panes to PaneState.
func mapPaneStates(panes []zellij.Pane) []PaneState {
	states := make([]PaneState, len(panes))
	for i, p := range panes {
		states[i] = PaneState{
			Title:     p.Title,
			Index:     p.Index,
			AgentType: string(p.Type),
			Model:     p.Variant,
			Active:    p.Active,
			Width:     p.Width,
			Height:    p.Height,
			PaneID:    p.ID,
		}
	}
	return states
}

// detectWorkDir attempts to detect the working directory for the session.
func detectWorkDir(sessionName string, panes []zellij.Pane) string {
	// Zellij doesn't expose pane CWD directly like tmux
	// Use current process working directory as best guess
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	// Final fallback: user home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		return homeDir
	}

	return ""
}

// getGitInfo extracts git branch, remote, and commit from a directory.
func getGitInfo(dir string) (branch, remote, commit string) {
	if dir == "" {
		return "", "", ""
	}

	// Get current branch
	branchOutput, err := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOutput))
	}

	// Get remote URL
	remoteOutput, err := exec.Command("git", "-C", dir, "remote", "get-url", "origin").Output()
	if err == nil {
		remote = strings.TrimSpace(string(remoteOutput))
	}

	// Get current commit
	commitOutput, err := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commit = strings.TrimSpace(string(commitOutput))
	}

	return branch, remote, commit
}

// getLayout gets the current layout for the session.
func getLayout(sessionName string) string {
	// Zellij layouts work differently - return default
	return "tiled"
}
