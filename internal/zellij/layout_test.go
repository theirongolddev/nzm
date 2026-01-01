package zellij

import (
	"strings"
	"testing"
)

func TestLayoutOptions_Defaults(t *testing.T) {
	opts := LayoutOptions{
		Session: "myproject",
	}

	if opts.CCCount != 0 {
		t.Errorf("expected CCCount=0, got %d", opts.CCCount)
	}
}

func TestGenerateLayout_BasicSession(t *testing.T) {
	opts := LayoutOptions{
		Session:    "myproject",
		WorkDir:    "/home/user/project",
		CCCount:    2,
		PluginPath: "/path/to/nzm-agent.wasm",
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check for required elements
	checks := []string{
		`cwd "/home/user/project"`,
		`name="myproject__cc_1"`,
		`name="myproject__cc_2"`,
		"nzm-agent.wasm",
	}

	for _, check := range checks {
		if !strings.Contains(kdl, check) {
			t.Errorf("expected layout to contain %q", check)
		}
	}
}

func TestGenerateLayout_IncludesPluginPane(t *testing.T) {
	opts := LayoutOptions{
		Session:    "test",
		PluginPath: "/path/to/plugin.wasm",
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(kdl, `plugin location="file:/path/to/plugin.wasm"`) {
		t.Error("expected plugin pane in layout")
	}
	if !strings.Contains(kdl, "borderless=true") {
		t.Error("expected borderless=true for plugin pane")
	}
}

func TestGenerateLayout_PaneNamingConvention(t *testing.T) {
	opts := LayoutOptions{
		Session:  "myproj",
		CCCount:  1,
		GmiCount: 2,
		CodCount: 1,
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		`name="myproj__cc_1"`,
		`name="myproj__gmi_1"`,
		`name="myproj__gmi_2"`,
		`name="myproj__cod_1"`,
	}

	for _, check := range checks {
		if !strings.Contains(kdl, check) {
			t.Errorf("expected layout to contain %q", check)
		}
	}
}

func TestGenerateLayout_AgentCommands(t *testing.T) {
	opts := LayoutOptions{
		Session:   "test",
		CCCount:   1,
		ClaudeCmd: "claude --dangerously-skip-permissions",
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have command in the pane
	if !strings.Contains(kdl, `command "claude"`) {
		t.Error("expected command clause in layout")
	}
}

func TestGenerateLayout_NoAgents(t *testing.T) {
	opts := LayoutOptions{
		Session:    "empty",
		PluginPath: "/path/to/plugin.wasm",
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have plugin pane
	if !strings.Contains(kdl, "plugin location=") {
		t.Error("expected plugin pane even with no agents")
	}
}

func TestGenerateLayout_WithUserPane(t *testing.T) {
	opts := LayoutOptions{
		Session:      "myproj",
		IncludeUser:  true,
		CCCount:      1,
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(kdl, `name="myproj__user_1"`) {
		t.Error("expected user pane in layout")
	}
}

func TestGenerateLayout_DefaultPluginPath(t *testing.T) {
	opts := LayoutOptions{
		Session: "test",
		CCCount: 1,
	}

	kdl, err := GenerateLayout(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use default plugin path
	if !strings.Contains(kdl, "nzm-agent") {
		t.Error("expected default plugin reference in layout")
	}
}

func TestGeneratePaneName(t *testing.T) {
	tests := []struct {
		session  string
		agentType string
		index    int
		expected string
	}{
		{"proj", "cc", 1, "proj__cc_1"},
		{"proj", "cod", 2, "proj__cod_2"},
		{"my-proj", "gmi", 1, "my-proj__gmi_1"},
		{"test", "user", 1, "test__user_1"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := GeneratePaneName(tt.session, tt.agentType, tt.index)
			if name != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, name)
			}
		})
	}
}

func TestParsePaneName(t *testing.T) {
	tests := []struct {
		name        string
		session     string
		agentType   string
		index       int
		shouldParse bool
	}{
		{"proj__cc_1", "proj", "cc", 1, true},
		{"my-proj__gmi_2", "my-proj", "gmi", 2, true},
		{"test__user_1", "test", "user", 1, true},
		{"invalid", "", "", 0, false},
		{"no__type", "", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, agentType, index, ok := ParsePaneName(tt.name)
			if ok != tt.shouldParse {
				t.Errorf("expected parse=%v, got %v", tt.shouldParse, ok)
				return
			}
			if !ok {
				return
			}
			if session != tt.session {
				t.Errorf("expected session=%q, got %q", tt.session, session)
			}
			if agentType != tt.agentType {
				t.Errorf("expected agentType=%q, got %q", tt.agentType, agentType)
			}
			if index != tt.index {
				t.Errorf("expected index=%d, got %d", tt.index, index)
			}
		})
	}
}
