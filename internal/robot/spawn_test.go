package robot

import (
	"encoding/json"
	"testing"

	"github.com/Dicklesworthstone/ntm/internal/config"
	"github.com/Dicklesworthstone/ntm/internal/zellij"
)

func TestPrintSpawn(t *testing.T) {
	if !zellij.IsInstalled() {
		t.Skip("zellij not installed")
	}

	// Use mock options that don't actually spawn heavy processes if possible,
	// but PrintSpawn calls logic that calls zellij.

	// We can use a test session name
	opts := SpawnOptions{
		Session:    "test_spawn_robot",
		CCCount:    1,
		NoUserPane: true,
	}

	cfg := config.Default()
	// Override agent command to be fast
	cfg.Agents.Claude = "echo test"

	// Clean up potential session
	defer zellij.KillSession(opts.Session)

	output, err := captureStdout(t, func() error { return PrintSpawn(opts, cfg) })
	if err != nil {
		t.Fatalf("PrintSpawn failed: %v", err)
	}

	// Check JSON output
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if resp["session"] != opts.Session {
		t.Errorf("Expected session %q, got %v", opts.Session, resp["session"])
	}
	// SpawnOutput doesn't have Created bool, check Layout instead
	if resp["layout"] != "tiled" {
		t.Errorf("Expected layout 'tiled', got %v", resp["layout"])
	}
}

func TestAgentTypeShort(t *testing.T) {
	tests := []struct {
		input    zellij.AgentType
		expected string
	}{
		{zellij.AgentClaude, "cc"},
		{zellij.AgentCodex, "cod"},
		{zellij.AgentGemini, "gmi"},
		{zellij.AgentUser, "user"},
	}

	for _, tc := range tests {
		if got := agentTypeShort(string(tc.input)); got != tc.expected {
			t.Errorf("agentTypeShort(%v) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
