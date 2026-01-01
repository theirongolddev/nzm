package zellij

import (
	"testing"
)

func TestParseAgentFromTitle_Claude(t *testing.T) {
	agentType, variant, tags := parseAgentFromTitle("project__cc_1")
	if agentType != AgentClaude {
		t.Errorf("expected AgentClaude, got %v", agentType)
	}
	if variant != "" {
		t.Errorf("expected empty variant, got %q", variant)
	}
	if len(tags) != 0 {
		t.Errorf("expected no tags, got %v", tags)
	}
}

func TestParseAgentFromTitle_ClaudeWithVariant(t *testing.T) {
	agentType, variant, tags := parseAgentFromTitle("project__cc_1_opus")
	if agentType != AgentClaude {
		t.Errorf("expected AgentClaude, got %v", agentType)
	}
	if variant != "opus" {
		t.Errorf("expected variant 'opus', got %q", variant)
	}
	if len(tags) != 0 {
		t.Errorf("expected no tags, got %v", tags)
	}
}

func TestParseAgentFromTitle_WithTags(t *testing.T) {
	agentType, variant, tags := parseAgentFromTitle("project__cc_1[frontend,api]")
	if agentType != AgentClaude {
		t.Errorf("expected AgentClaude, got %v", agentType)
	}
	if variant != "" {
		t.Errorf("expected empty variant, got %q", variant)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", tags)
	}
	if tags[0] != "frontend" || tags[1] != "api" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestParseAgentFromTitle_VariantAndTags(t *testing.T) {
	agentType, variant, tags := parseAgentFromTitle("project__cc_1_opus[backend]")
	if agentType != AgentClaude {
		t.Errorf("expected AgentClaude, got %v", agentType)
	}
	if variant != "opus" {
		t.Errorf("expected variant 'opus', got %q", variant)
	}
	if len(tags) != 1 || tags[0] != "backend" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestParseAgentFromTitle_Codex(t *testing.T) {
	agentType, _, _ := parseAgentFromTitle("project__cod_2")
	if agentType != AgentCodex {
		t.Errorf("expected AgentCodex, got %v", agentType)
	}
}

func TestParseAgentFromTitle_Gemini(t *testing.T) {
	agentType, _, _ := parseAgentFromTitle("project__gmi_3")
	if agentType != AgentGemini {
		t.Errorf("expected AgentGemini, got %v", agentType)
	}
}

func TestParseAgentFromTitle_UnknownType(t *testing.T) {
	agentType, _, _ := parseAgentFromTitle("project__xyz_1")
	if agentType != AgentUser {
		t.Errorf("expected AgentUser for unknown type, got %v", agentType)
	}
}

func TestParseAgentFromTitle_InvalidFormat(t *testing.T) {
	agentType, _, _ := parseAgentFromTitle("just a regular title")
	if agentType != AgentUser {
		t.Errorf("expected AgentUser for invalid format, got %v", agentType)
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		tags []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"frontend"}, "[frontend]"},
		{[]string{"frontend", "api"}, "[frontend,api]"},
	}

	for _, tt := range tests {
		got := FormatTags(tt.tags)
		if got != tt.want {
			t.Errorf("FormatTags(%v) = %q, want %q", tt.tags, got, tt.want)
		}
	}
}

func TestConvertPaneInfo(t *testing.T) {
	info := PaneInfo{
		ID:         1,
		Title:      "project__cc_1_opus[frontend]",
		IsFocused:  true,
		IsFloating: false,
	}

	pane := ConvertPaneInfo(info)

	if pane.ID != "1" {
		t.Errorf("expected ID \"1\", got %q", pane.ID)
	}
	if pane.Type != AgentClaude {
		t.Errorf("expected AgentClaude, got %v", pane.Type)
	}
	if pane.Variant != "opus" {
		t.Errorf("expected variant 'opus', got %q", pane.Variant)
	}
	if len(pane.Tags) != 1 || pane.Tags[0] != "frontend" {
		t.Errorf("unexpected tags: %v", pane.Tags)
	}
	if !pane.Active {
		t.Error("expected Active=true")
	}
}
