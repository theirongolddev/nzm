package cli

import (
	"fmt"
	"strings"

	"github.com/Dicklesworthstone/ntm/internal/codeblock"
	"github.com/Dicklesworthstone/ntm/internal/output"
	"github.com/Dicklesworthstone/ntm/internal/tmux"
	"github.com/spf13/cobra"
)

// ExtractResponse is the JSON output format for extract command.
type ExtractResponse struct {
	output.TimestampedResponse
	Session    string               `json:"session"`
	Source     string               `json:"source"`     // Session or session:pane
	Blocks     []codeblock.CodeBlock `json:"blocks"`
	TotalFound int                  `json:"total_found"`
}

func newExtractCmd() *cobra.Command {
	var (
		language  string
		lastPane  bool
		lines     int
		paneIndex string
	)

	cmd := &cobra.Command{
		Use:   "extract <session> [pane]",
		Short: "Extract code blocks from agent output",
		Long: `Extract markdown code blocks from agent pane output.

Examples:
  ntm extract myproject              # Extract from all panes
  ntm extract myproject cc_1         # Extract from specific pane
  ntm extract myproject --last       # From most recent active pane
  ntm extract myproject --lang=python  # Only Python blocks
  ntm extract myproject --json       # Output as JSON`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionName := args[0]

			// Get optional pane argument
			if len(args) > 1 {
				paneIndex = args[1]
			}

			return runExtract(sessionName, paneIndex, language, lastPane, lines)
		},
	}

	cmd.Flags().StringVarP(&language, "lang", "l", "", "Filter by language (e.g., python, go, bash)")
	cmd.Flags().BoolVar(&lastPane, "last", false, "Extract from most recent active pane")
	cmd.Flags().IntVarP(&lines, "lines", "n", 500, "Number of lines to capture from pane")

	return cmd
}

func runExtract(sessionName, paneIndex, language string, lastPane bool, lines int) error {
	// Check session exists
	if !tmux.SessionExists(sessionName) {
		if IsJSONOutput() {
			return output.PrintJSON(output.NewErrorWithCode("SESSION_NOT_FOUND",
				fmt.Sprintf("session '%s' does not exist", sessionName)))
		}
		return fmt.Errorf("session '%s' does not exist", sessionName)
	}

	// Get panes
	panes, err := tmux.GetPanes(sessionName)
	if err != nil {
		if IsJSONOutput() {
			return output.PrintJSON(output.NewErrorWithDetails("failed to get panes", err.Error()))
		}
		return fmt.Errorf("failed to get panes: %w", err)
	}

	// Filter panes
	var targetPanes []tmux.Pane
	if lastPane {
		// Find most recently active pane
		for _, p := range panes {
			if p.Active {
				targetPanes = []tmux.Pane{p}
				break
			}
		}
		if len(targetPanes) == 0 && len(panes) > 0 {
			targetPanes = []tmux.Pane{panes[len(panes)-1]}
		}
	} else if paneIndex != "" {
		// Find specific pane by index or title
		for _, p := range panes {
			paneIdxStr := fmt.Sprintf("%d", p.Index)
			if paneIdxStr == paneIndex || p.Title == paneIndex || strings.Contains(p.Title, paneIndex) {
				targetPanes = []tmux.Pane{p}
				break
			}
		}
		if len(targetPanes) == 0 {
			if IsJSONOutput() {
				return output.PrintJSON(output.NewErrorWithCode("PANE_NOT_FOUND",
					fmt.Sprintf("pane '%s' not found", paneIndex)))
			}
			return fmt.Errorf("pane '%s' not found in session '%s'", paneIndex, sessionName)
		}
	} else {
		// All panes
		targetPanes = panes
	}

	// Create parser with optional language filter
	parser := codeblock.NewParser()
	if language != "" {
		parser.WithLanguageFilter([]string{language})
	}

	// Extract from all target panes
	var allBlocks []codeblock.CodeBlock
	source := sessionName
	if paneIndex != "" {
		source = fmt.Sprintf("%s:%s", sessionName, paneIndex)
	}

	for _, pane := range targetPanes {
		// Capture pane output
		captured, err := tmux.CapturePaneOutput(pane.ID, lines)
		if err != nil {
			continue // Skip panes that fail
		}

		// Parse code blocks
		blocks := parser.Parse(captured)

		// Add source pane info
		for i := range blocks {
			blocks[i].SourcePane = fmt.Sprintf("%s:%d", sessionName, pane.Index)
		}

		allBlocks = append(allBlocks, blocks...)
	}

	// Output
	if IsJSONOutput() {
		response := ExtractResponse{
			TimestampedResponse: output.NewTimestamped(),
			Session:             sessionName,
			Source:              source,
			Blocks:              allBlocks,
			TotalFound:          len(allBlocks),
		}
		return output.PrintJSON(response)
	}

	// Text output
	if len(allBlocks) == 0 {
		fmt.Printf("No code blocks found in %s\n", source)
		return nil
	}

	fmt.Printf("Found %d code block(s) in %s:\n\n", len(allBlocks), source)

	for i, block := range allBlocks {
		// Header
		langDisplay := block.Language
		if langDisplay == "" {
			langDisplay = "text"
		}

		fileInfo := ""
		if block.FilePath != "" {
			if block.IsNew {
				fileInfo = fmt.Sprintf(" (new file: %s)", block.FilePath)
			} else {
				fileInfo = fmt.Sprintf(" (%s)", block.FilePath)
			}
		}

		fmt.Printf("[%d] %s%s (from %s)\n", i+1, langDisplay, fileInfo, block.SourcePane)

		// Box around content
		fmt.Println("    ┌" + strings.Repeat("─", 60))
		for _, line := range strings.Split(block.Content, "\n") {
			if len(line) > 58 {
				line = line[:55] + "..."
			}
			fmt.Printf("    │ %s\n", line)
		}
		fmt.Println("    └" + strings.Repeat("─", 60))
		fmt.Println()
	}

	return nil
}

func init() {
	// Note: This will be added to rootCmd in root.go's init()
}
