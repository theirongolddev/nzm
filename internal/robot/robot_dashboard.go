package robot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/alerts"
	"github.com/Dicklesworthstone/ntm/internal/bv"
	"github.com/Dicklesworthstone/ntm/internal/tmux"
	"github.com/Dicklesworthstone/ntm/internal/tracker"
)

// DashboardOutput provides a concise dashboard view for AI orchestrators.
type DashboardOutput struct {
	GeneratedAt  time.Time           `json:"generated_at"`
	Fleet        string              `json:"fleet"`
	Agents       []SnapshotSession   `json:"agents"`
	Metrics      map[string]any      `json:"metrics,omitempty"`
	System       SystemInfo          `json:"system"`
	Summary      StatusSummary       `json:"summary"`
	Beads        *bv.BeadsSummary    `json:"beads,omitempty"`
	Alerts       []AlertInfo         `json:"alerts,omitempty"`
	AlertSummary *AlertSummaryInfo   `json:"alert_summary,omitempty"`
	Conflicts    []tracker.Conflict  `json:"conflicts,omitempty"`
	FileChanges  []FileChangeInfo    `json:"file_changes,omitempty"`
	AgentMail    *SnapshotAgentMail  `json:"agent_mail,omitempty"`
}

// PrintDashboard outputs a dashboard-oriented view for AI orchestrators.
func PrintDashboard(jsonMode bool) error {
	wd, _ := os.Getwd()
	fleet := "ntm"
	if wd != "" {
		fleet = filepath.Base(wd)
	}

	output := DashboardOutput{
		GeneratedAt: time.Now().UTC(),
		Fleet:       fleet,
		Agents:      []SnapshotSession{},
		Metrics:     map[string]any{},
		System: SystemInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			GoVersion: runtime.Version(),
			Version:   Version,
			Commit:    Commit,
			BuildDate: Date,
			TmuxOK:    tmux.IsInstalled(),
		},
		Summary: StatusSummary{},
	}

	// Sessions and agents (best-effort)
	if tmux.IsInstalled() {
		sessions, err := tmux.ListSessions()
		if err == nil {
			for _, sess := range sessions {
				snapSession := SnapshotSession{
					Name:     sess.Name,
					Attached: sess.Attached,
					Agents:   []SnapshotAgent{},
				}
			panes, err := tmux.GetPanes(sess.Name)
			if err == nil {
				for _, pane := range panes {
					agentType := agentTypeString(pane.Type)
					snapSession.Agents = append(snapSession.Agents, SnapshotAgent{
						Pane:           fmt.Sprintf("%d.%d", 0, pane.Index),
						Type:           agentType,
						Variant:        pane.Variant,
						TypeConfidence: 0.5,
						TypeMethod:     "tmux-pane",
						State:          "unknown",
					})

					switch agentType {
					case "claude":
						output.Summary.ClaudeCount++
					case "codex":
						output.Summary.CodexCount++
					case "gemini":
						output.Summary.GeminiCount++
					case "cursor":
						output.Summary.CursorCount++
					case "windsurf":
						output.Summary.WindsurfCount++
					case "aider":
						output.Summary.AiderCount++
					}
					output.Summary.TotalAgents++
				}
			}
			output.Agents = append(output.Agents, snapSession)
			output.Summary.TotalSessions++
			if sess.Attached {
				output.Summary.AttachedCount++
			}
			}
		}
	}

	// Beads summary (best-effort)
	if bv.IsInstalled() {
		output.Beads = bv.GetBeadsSummary(BeadLimit)
	}

	// Alerts (best-effort)
	alertCfg := alerts.DefaultConfig()
	activeAlerts := alerts.GetActiveAlerts(alertCfg)
	for _, a := range activeAlerts {
		output.Alerts = append(output.Alerts, AlertInfo{
			ID:         a.ID,
			Type:       string(a.Type),
			Severity:   string(a.Severity),
			Message:    a.Message,
			Session:    a.Session,
			Pane:       a.Pane,
			BeadID:     a.BeadID,
			Context:    a.Context,
			CreatedAt:  a.CreatedAt.Format(time.RFC3339),
			DurationMs: a.Duration().Milliseconds(),
			Count:      a.Count,
		})
	}
	output.AlertSummary = &AlertSummaryInfo{
		TotalActive: len(activeAlerts),
	}

	// Conflicts and file changes (best-effort)
	statusOutput := StatusOutput{}
	appendConflicts(&statusOutput)
	appendFileChanges(&statusOutput)
	output.Conflicts = statusOutput.Conflicts
	output.FileChanges = statusOutput.FileChanges

	// Agent Mail summary (best-effort)
	output.AgentMail = buildSnapshotAgentMail()

	if jsonMode {
		return encodeJSON(output)
	}

	// Markdown output
	fmt.Printf("# NTM Fleet Dashboard: %s\n\n", output.Fleet)
	fmt.Printf("**Generated:** %s\n", output.GeneratedAt.Format(time.RFC3339))
	fmt.Printf("**System:** %s/%s %s\n\n", output.System.OS, output.System.Arch, output.System.Version)

	fmt.Println("## Summary")
	fmt.Printf("- Sessions: %d (%d attached)\n", output.Summary.TotalSessions, output.Summary.AttachedCount)
	fmt.Printf("- Agents: %d (CC:%d COD:%d GMI:%d)\n\n", output.Summary.TotalAgents, output.Summary.ClaudeCount, output.Summary.CodexCount, output.Summary.GeminiCount)

	if len(output.Alerts) > 0 {
		fmt.Println("## Critical Alerts")
		for _, a := range output.Alerts {
			icon := "ℹ️"
			if a.Severity == "critical" {
				icon = "🚨"
			} else if a.Severity == "warning" {
				icon = "⚠️"
			}
			fmt.Printf("- %s [%s] %s\n", icon, a.ID, a.Message)
		}
		fmt.Println()
	}

	if output.Beads != nil && len(output.Beads.ReadyPreview) > 0 {
		fmt.Println("## Ready Beads")
		for _, b := range output.Beads.ReadyPreview {
			fmt.Printf("- %s %s (%s)\n", b.ID, b.Title, b.Priority)
		}
		fmt.Println()
	}

	return nil
}