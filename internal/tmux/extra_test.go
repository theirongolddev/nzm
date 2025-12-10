package tmux

import (
	"testing"
)

func TestBuildPaneCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		projectDir string
		cmd        string
		want       string
		wantErr    bool
	}{
		{
			name:       "simple command",
			projectDir: "/projects/foo",
			cmd:        "ls -la",
			want:       "cd \"/projects/foo\" && ls -la",
			wantErr:    false,
		},
		{
			name:       "command with spaces",
			projectDir: "/projects/foo bar",
			cmd:        "echo hello",
			want:       "cd \"/projects/foo bar\" && echo hello",
			wantErr:    false,
		},
		{
			name:       "unsafe command",
			projectDir: "/projects/foo",
			cmd:        "echo hello\nrm -rf /",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildPaneCommand(tt.projectDir, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPaneCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("BuildPaneCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPanes_Error(t *testing.T) {
	t.Parallel()
	skipIfNoTmux(t)
	_, err := GetPanes("nonexistent_session_12345")
	if err == nil {
		t.Error("GetPanes should fail for non-existent session")
	}
}

func TestGetFirstWindow_Error(t *testing.T) {
	t.Parallel()
	skipIfNoTmux(t)
	_, err := GetFirstWindow("nonexistent_session_12345")
	if err == nil {
		t.Error("GetFirstWindow should fail for non-existent session")
	}
}

func TestGetDefaultPaneIndex_Error(t *testing.T) {
	t.Parallel()
	skipIfNoTmux(t)
	_, err := GetDefaultPaneIndex("nonexistent_session_12345")
	if err == nil {
		t.Error("GetDefaultPaneIndex should fail for non-existent session")
	}
}

func TestZoomPane_Error(t *testing.T) {
	t.Parallel()
	skipIfNoTmux(t)
	err := ZoomPane("nonexistent_session_12345", 0)
	if err == nil {
		t.Error("ZoomPane should fail for non-existent session")
	}
}

func TestGetCurrentSession_Simulated(t *testing.T) {
	// cannot run in parallel due to t.Setenv
	skipIfNoTmux(t)
	// Simulate being in tmux but command failing (since we aren't actually in a client)
	t.Setenv("TMUX", "/tmp/tmux-1000/default,123,0")

	// usage of run() will likely fail or return empty because we aren't attached
	// but we want to ensure it doesn't panic
	session := GetCurrentSession()
	if session != "" {
		// If it actually works (e.g. nested tmux), that's fine too, but unlikely in this env
		t.Logf("GetCurrentSession returned: %q", session)
	}
}
func TestParseAgentFromTitle_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		title       string
		wantType    AgentType
		wantVariant string
	}{
		{"invalid_format", AgentUser, ""},
		{"session__invalid_1", AgentUser, ""}, // valid regex but invalid type
		{"session__cc_1", AgentClaude, ""},
		{"session__cc_1_variant", AgentClaude, "variant"},
		{"session__cod_2_gpt4", AgentCodex, "gpt4"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gotType, gotVariant := parseAgentFromTitle(tt.title)
			if gotType != tt.wantType {
				t.Errorf("parseAgentFromTitle() type = %v, want %v", gotType, tt.wantType)
			}
			if gotVariant != tt.wantVariant {
				t.Errorf("parseAgentFromTitle() variant = %q, want %q", gotVariant, tt.wantVariant)
			}
		})
	}
}
