package zellij

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// paneNameWithMetaRegex matches the NZM pane naming convention with variants and tags:
// session__type_index or session__type_index_variant, optionally with tags [tag1,tag2]
// Examples:
//
//	session__cc_1
//	session__cc_1[frontend]
//	session__cc_1_opus[backend,api]
var paneNameWithMetaRegex = regexp.MustCompile(`^.+__(\w+)_\d+(?:_([A-Za-z0-9._/@:+-]+))?(?:\[([^\]]*)\])?$`)

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentClaude AgentType = "cc"
	AgentCodex  AgentType = "cod"
	AgentGemini AgentType = "gmi"
	AgentUser   AgentType = "user"
)

// Pane represents a Zellij pane with NZM metadata
type Pane struct {
	ID         string    // String ID for API compatibility (internally uint32)
	Index      int
	Title      string
	Type       AgentType
	Variant    string   // Model alias or persona name
	Tags       []string // User-defined tags
	Command    string   // Not available in Zellij - kept for compatibility
	Width      int      // Not available in Zellij - kept for compatibility
	Height     int      // Not available in Zellij - kept for compatibility
	Active     bool     // Mapped from IsFocused
	IsPlugin   bool     // Whether this is a plugin pane
	IsFloating bool
}

// parseAgentFromTitle extracts agent type, variant, and tags from a pane title.
// Title format: {session}__{type}_{index}[tags] or {session}__{type}_{index}_{variant}[tags]
// Returns AgentUser, empty variant, and nil tags if title doesn't match NZM format.
func parseAgentFromTitle(title string) (AgentType, string, []string) {
	matches := paneNameWithMetaRegex.FindStringSubmatch(title)
	if matches == nil {
		return AgentUser, "", nil
	}

	agentType := AgentType(matches[1])
	variant := matches[2]
	var tags []string
	if len(matches) >= 4 {
		tags = parseTags(matches[3])
	}

	switch agentType {
	case AgentClaude, AgentCodex, AgentGemini:
		return agentType, variant, tags
	default:
		return AgentUser, "", nil
	}
}

// parseTags parses a comma-separated tag string into a slice.
func parseTags(tagStr string) []string {
	if tagStr == "" {
		return nil
	}
	parts := strings.Split(tagStr, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// FormatTags formats tags as a bracket-enclosed string for pane titles.
func FormatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "[" + strings.Join(tags, ",") + "]"
}

// stripTags removes the [tags] suffix from a pane title.
func stripTags(title string) string {
	idx := strings.LastIndex(title, "[")
	if idx == -1 {
		return title
	}
	if strings.HasSuffix(title, "]") && idx < len(title)-1 {
		return title[:idx]
	}
	return title
}

// ConvertPaneInfo converts a PaneInfo from the plugin to a full Pane struct
func ConvertPaneInfo(info PaneInfo) Pane {
	agentType, variant, tags := parseAgentFromTitle(info.Title)
	return Pane{
		ID:         strconv.FormatUint(uint64(info.ID), 10),
		Index:      int(info.ID), // Use ID as index for Zellij
		Title:      info.Title,
		Type:       agentType,
		Variant:    variant,
		Tags:       tags,
		Active:     info.IsFocused,
		IsFloating: info.IsFloating,
	}
}

// PaneActivity contains pane info with activity tracking
type PaneActivity struct {
	Pane         Pane
	LastActivity time.Time
	IsActive     bool // Whether pane is currently focused/active
}
