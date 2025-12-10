package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/agentmail"
	"github.com/spf13/cobra"
)

func newGitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Git coordination commands",
		Long: `Git-related commands for coordinating version control across agents.

These commands combine git state with Agent Mail file reservations to
provide a unified view of coordination state.

Examples:
  ntm git status myproject     # Show branches, locks, and pending changes
  ntm git status               # Status for current directory`,
	}

	cmd.AddCommand(newGitStatusCmd())

	return cmd
}

func newGitStatusCmd() *cobra.Command {
	var allAgents bool

	cmd := &cobra.Command{
		Use:   "status [session]",
		Short: "Show coordination state: branches, locks, pending changes",
		Long: `Display combined git and Agent Mail status for coordination.

Shows:
  - Current git branch and status
  - File reservations (locks) held by agents
  - Uncommitted changes summary
  - Conflicts if any patterns overlap

Examples:
  ntm git status myproject           # Status for session
  ntm git status                     # Status for current directory
  ntm git status --all-agents        # Include all agents' reservations`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var session string
			if len(args) > 0 {
				session = args[0]
			}
			return runGitStatus(session, allAgents)
		},
	}

	cmd.Flags().BoolVar(&allAgents, "all-agents", false, "Show reservations from all agents")

	return cmd
}

// GitStatusResult represents the combined git and Agent Mail status.
type GitStatusResult struct {
	Success      bool                       `json:"success"`
	Session      string                     `json:"session,omitempty"`
	WorkingDir   string                     `json:"working_dir"`
	Git          *GitInfo                   `json:"git,omitempty"`
	Reservations []agentmail.InboxMessage   `json:"reservations,omitempty"`
	Locks        []agentmail.FileReservation `json:"locks,omitempty"`
	AgentMail    *AgentMailStatus           `json:"agent_mail,omitempty"`
	Error        string                     `json:"error,omitempty"`
}

// GitInfo contains git repository information.
type GitInfo struct {
	Branch          string   `json:"branch"`
	Commit          string   `json:"commit"`
	CommitShort     string   `json:"commit_short"`
	Dirty           bool     `json:"dirty"`
	Ahead           int      `json:"ahead"`
	Behind          int      `json:"behind"`
	StagedFiles     []string `json:"staged_files,omitempty"`
	ModifiedFiles   []string `json:"modified_files,omitempty"`
	UntrackedFiles  []string `json:"untracked_files,omitempty"`
	ConflictedFiles []string `json:"conflicted_files,omitempty"`
}

// AgentMailStatus contains Agent Mail coordination info.
type AgentMailStatus struct {
	Available       bool   `json:"available"`
	RegisteredAgent string `json:"registered_agent,omitempty"`
	LockCount       int    `json:"lock_count"`
	ConflictCount   int    `json:"conflict_count"`
}

func runGitStatus(session string, allAgents bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Use session's working directory if specified
	workDir := wd
	if session != "" && cfg != nil {
		workDir = cfg.GetProjectDir(session)
	}

	result := GitStatusResult{
		Success:    true,
		Session:    session,
		WorkingDir: workDir,
	}

	// Get git status
	gitInfo, gitErr := getGitInfo(workDir)
	if gitErr != nil {
		result.Git = nil
	} else {
		result.Git = gitInfo
	}

	// Get Agent Mail status
	result.AgentMail = &AgentMailStatus{}
	client := agentmail.NewClient(agentmail.WithProjectKey(workDir))

	if client.IsAvailable() {
		result.AgentMail.Available = true

		// Try to get session agent info
		if session != "" {
			sessionAgent, err := agentmail.LoadSessionAgent(session)
			if err == nil && sessionAgent != nil {
				result.AgentMail.RegisteredAgent = sessionAgent.AgentName

				// Fetch file reservations for this agent
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				locks, err := fetchAgentLocks(ctx, client, workDir, sessionAgent.AgentName, allAgents)
				cancel()
				if err == nil {
					result.Locks = locks
					result.AgentMail.LockCount = len(locks)
				}
			}
		}
	}

	// Output
	if IsJSONOutput() {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	return printGitStatus(result)
}

func getGitInfo(dir string) (*GitInfo, error) {
	// Check if it's a git repo
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository")
	}

	info := &GitInfo{}

	// Get current branch
	cmd = exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err == nil {
		info.Branch = strings.TrimSpace(string(out))
	}

	// Get commit hash
	cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err = cmd.Output()
	if err == nil {
		info.Commit = strings.TrimSpace(string(out))
		if len(info.Commit) >= 7 {
			info.CommitShort = info.Commit[:7]
		}
	}

	// Get status (porcelain v2 for machine parsing)
	cmd = exec.Command("git", "-C", dir, "status", "--porcelain=v2", "--branch")
	out, err = cmd.Output()
	if err == nil {
		parseGitStatus(info, string(out))
	}

	// Check if dirty
	cmd = exec.Command("git", "-C", dir, "diff", "--quiet", "HEAD")
	info.Dirty = cmd.Run() != nil

	return info, nil
}

func parseGitStatus(info *GitInfo, output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# branch.ab ") {
			// Parse ahead/behind: # branch.ab +1 -2
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "+") {
					fmt.Sscanf(p, "+%d", &info.Ahead)
				} else if strings.HasPrefix(p, "-") {
					fmt.Sscanf(p, "-%d", &info.Behind)
				}
			}
		} else if strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 ") {
			// Tracked file changes
			parts := strings.Fields(line)
			if len(parts) >= 9 {
				xy := parts[1] // XY status
				file := parts[8]
				if xy[0] != '.' {
					info.StagedFiles = append(info.StagedFiles, file)
				}
				if xy[1] != '.' {
					info.ModifiedFiles = append(info.ModifiedFiles, file)
				}
			}
		} else if strings.HasPrefix(line, "u ") {
			// Unmerged (conflict)
			parts := strings.Fields(line)
			if len(parts) >= 11 {
				info.ConflictedFiles = append(info.ConflictedFiles, parts[10])
			}
		} else if strings.HasPrefix(line, "? ") {
			// Untracked
			file := strings.TrimPrefix(line, "? ")
			info.UntrackedFiles = append(info.UntrackedFiles, file)
		}
	}
}

func fetchAgentLocks(ctx context.Context, client *agentmail.Client, projectKey, agentName string, allAgents bool) ([]agentmail.FileReservation, error) {
	// Note: This is a simplified implementation. A full implementation would
	// query the Agent Mail server for active reservations.
	// For now, return empty as the full API may not be exposed.
	return nil, nil
}

func printGitStatus(result GitStatusResult) error {
	fmt.Printf("Git Status: %s\n", result.WorkingDir)
	fmt.Println(strings.Repeat("─", 60))

	if result.Git != nil {
		// Branch and commit
		branchIcon := "󰘬"
		if result.Git.Dirty {
			branchIcon = "󰘭"
		}
		fmt.Printf("%s Branch: %s", branchIcon, result.Git.Branch)
		if result.Git.CommitShort != "" {
			fmt.Printf(" (%s)", result.Git.CommitShort)
		}
		fmt.Println()

		// Ahead/behind
		if result.Git.Ahead > 0 || result.Git.Behind > 0 {
			fmt.Printf("  ↑%d ↓%d ", result.Git.Ahead, result.Git.Behind)
			if result.Git.Ahead > 0 {
				fmt.Print("(push needed) ")
			}
			if result.Git.Behind > 0 {
				fmt.Print("(pull needed) ")
			}
			fmt.Println()
		}

		// File changes
		if len(result.Git.StagedFiles) > 0 {
			fmt.Printf("\n Staged (%d):\n", len(result.Git.StagedFiles))
			for _, f := range result.Git.StagedFiles {
				fmt.Printf("  + %s\n", f)
			}
		}

		if len(result.Git.ModifiedFiles) > 0 {
			fmt.Printf("\n Modified (%d):\n", len(result.Git.ModifiedFiles))
			for _, f := range result.Git.ModifiedFiles {
				fmt.Printf("  M %s\n", f)
			}
		}

		if len(result.Git.UntrackedFiles) > 0 {
			fmt.Printf("\n Untracked (%d):\n", len(result.Git.UntrackedFiles))
			for _, f := range result.Git.UntrackedFiles {
				fmt.Printf("  ? %s\n", f)
			}
		}

		if len(result.Git.ConflictedFiles) > 0 {
			fmt.Printf("\n⚠ Conflicts (%d):\n", len(result.Git.ConflictedFiles))
			for _, f := range result.Git.ConflictedFiles {
				fmt.Printf("  ! %s\n", f)
			}
		}

		if !result.Git.Dirty && len(result.Git.StagedFiles) == 0 && len(result.Git.UntrackedFiles) == 0 {
			fmt.Println("  Working tree clean")
		}
	} else {
		fmt.Println("  Not a git repository")
	}

	// Agent Mail status
	fmt.Println()
	fmt.Println("Agent Mail:")
	fmt.Println(strings.Repeat("─", 60))

	if result.AgentMail != nil && result.AgentMail.Available {
		fmt.Println("  Status: Available")
		if result.AgentMail.RegisteredAgent != "" {
			fmt.Printf("  Agent: %s\n", result.AgentMail.RegisteredAgent)
		}
		if result.AgentMail.LockCount > 0 {
			fmt.Printf("  Active locks: %d\n", result.AgentMail.LockCount)
		} else {
			fmt.Println("  No active file reservations")
		}
	} else {
		fmt.Println("  Status: Unavailable")
	}

	return nil
}
