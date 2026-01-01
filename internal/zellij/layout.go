package zellij

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// LayoutOptions configures session layout generation
type LayoutOptions struct {
	SessionName string // Session name (alias for Session)
	Session     string // Session name
	ProjectDir  string // Project directory (alias for WorkDir)
	WorkDir     string // Working directory for panes
	PluginPath  string // Path to nzm-agent.wasm
	CCCount     int    // Number of Claude panes
	CodCount    int    // Number of Codex panes
	GmiCount    int    // Number of Gemini panes
	IncludeUser bool   // Include a user pane
	ClaudeCmd   string // Command to run for Claude panes
	CodCmd      string // Command to run for Codex panes
	GmiCmd      string // Command to run for Gemini panes
}

// DefaultPluginPath is the default plugin location
const DefaultPluginPath = "nzm-agent"

// GenerateLayout creates a KDL layout string for a Zellij session
func GenerateLayout(opts LayoutOptions) (string, error) {
	var sb strings.Builder

	// Handle aliases
	session := opts.Session
	if session == "" {
		session = opts.SessionName
	}
	workDir := opts.WorkDir
	if workDir == "" {
		workDir = opts.ProjectDir
	}

	sb.WriteString("layout {\n")

	// Add default cwd if specified
	if workDir != "" {
		sb.WriteString(fmt.Sprintf("    cwd %q\n", workDir))
	}

	// Determine plugin path
	pluginPath := opts.PluginPath
	if pluginPath == "" {
		pluginPath = DefaultPluginPath
	}

	// Add plugin pane (minimal size, borderless)
	sb.WriteString("    pane size=1 borderless=true {\n")
	if strings.HasPrefix(pluginPath, "/") {
		sb.WriteString(fmt.Sprintf("        plugin location=\"file:%s\"\n", pluginPath))
	} else {
		sb.WriteString(fmt.Sprintf("        plugin location=\"%s\"\n", pluginPath))
	}
	sb.WriteString("    }\n")

	// Add main pane area
	sb.WriteString("    pane split_direction=\"vertical\" {\n")

	// Add Claude panes
	for i := 1; i <= opts.CCCount; i++ {
		name := GeneratePaneName(session, "cc", i)
		writePane(&sb, name, opts.ClaudeCmd)
	}

	// Add Codex panes
	for i := 1; i <= opts.CodCount; i++ {
		name := GeneratePaneName(session, "cod", i)
		writePane(&sb, name, opts.CodCmd)
	}

	// Add Gemini panes
	for i := 1; i <= opts.GmiCount; i++ {
		name := GeneratePaneName(session, "gmi", i)
		writePane(&sb, name, opts.GmiCmd)
	}

	// Add user pane if requested
	if opts.IncludeUser {
		name := GeneratePaneName(session, "user", 1)
		writePane(&sb, name, "")
	}

	// If no panes specified, add a default shell pane
	if opts.CCCount == 0 && opts.CodCount == 0 && opts.GmiCount == 0 && !opts.IncludeUser {
		sb.WriteString("        pane\n")
	}

	sb.WriteString("    }\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

// WriteLayoutFile generates a layout and writes it to a temporary file.
// Returns the path to the temporary file.
func WriteLayoutFile(opts LayoutOptions) (string, error) {
	content, err := GenerateLayout(opts)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "nzm-layout-*.kdl")
	if err != nil {
		return "", fmt.Errorf("creating temp layout file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("writing layout file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("closing layout file: %w", err)
	}

	return tmpFile.Name(), nil
}

// writePane writes a pane definition to the string builder
func writePane(sb *strings.Builder, name, command string) {
	sb.WriteString(fmt.Sprintf("        pane name=%q", name))

	if command != "" {
		// Parse command into parts
		parts := parseCommand(command)
		if len(parts) > 0 {
			sb.WriteString(fmt.Sprintf(" {\n            command %q\n", parts[0]))
			if len(parts) > 1 {
				sb.WriteString("            args")
				for _, arg := range parts[1:] {
					sb.WriteString(fmt.Sprintf(" %q", arg))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("        }\n")
		}
	} else {
		sb.WriteString("\n")
	}
}

// parseCommand splits a command string into command and arguments
func parseCommand(cmd string) []string {
	// Simple split by spaces (doesn't handle quotes)
	// For more complex cases, we'd need a proper shell parser
	return strings.Fields(cmd)
}

// GeneratePaneName creates a pane name following the convention:
// {session}__{type}_{index}
func GeneratePaneName(session, agentType string, index int) string {
	return fmt.Sprintf("%s__%s_%d", session, agentType, index)
}

// paneNameRegex matches pane names like "proj__cc_1"
var paneNameRegex = regexp.MustCompile(`^(.+)__([a-z]+)_(\d+)$`)

// ParsePaneName extracts session, agent type, and index from a pane name
func ParsePaneName(name string) (session, agentType string, index int, ok bool) {
	matches := paneNameRegex.FindStringSubmatch(name)
	if matches == nil {
		return "", "", 0, false
	}

	session = matches[1]
	agentType = matches[2]
	idx, err := strconv.Atoi(matches[3])
	if err != nil {
		return "", "", 0, false
	}
	index = idx

	return session, agentType, index, true
}

// AgentTypes supported by NZM
const (
	AgentTypeClaude = "cc"
	AgentTypeCodex  = "cod"
	AgentTypeGemini = "gmi"
	AgentTypeUser   = "user"
)

// GetAgentTypeDisplay returns a human-readable name for an agent type
func GetAgentTypeDisplay(agentType string) string {
	switch agentType {
	case AgentTypeClaude:
		return "Claude"
	case AgentTypeCodex:
		return "Codex"
	case AgentTypeGemini:
		return "Gemini"
	case AgentTypeUser:
		return "User"
	default:
		return agentType
	}
}
