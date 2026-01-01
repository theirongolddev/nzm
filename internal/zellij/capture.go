package zellij

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CapturePaneOutput captures the output from a pane.
// Uses zellij's dump-screen action to capture pane contents.
// Note: This requires the session to be attached or uses pipe for detached sessions.
func (c *Client) CapturePaneOutput(ctx context.Context, session string, paneID uint32, lines int) (string, error) {
	// Create temp file for capture
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("nzm-capture-%d.txt", paneID))
	defer os.Remove(tmpFile)

	// First, try to use the plugin to get pane content if available
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "get_pane_content",
		Params: map[string]any{
			"pane_id": paneID,
			"lines":   lines,
		},
	})
	if err == nil && resp.Success {
		if content, ok := resp.Data["content"].(string); ok {
			return content, nil
		}
	}

	// Fallback: use dump-screen (requires attached session)
	// This is a limitation - dump-screen only works when attached
	_, err = c.Run(ctx, "action", "dump-screen", tmpFile, "--session", session)
	if err != nil {
		return "", fmt.Errorf("dump-screen failed: %w", err)
	}

	// Read captured content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("reading capture file: %w", err)
	}

	// Limit to requested lines
	output := string(content)
	if lines > 0 {
		outputLines := strings.Split(output, "\n")
		if len(outputLines) > lines {
			outputLines = outputLines[len(outputLines)-lines:]
		}
		output = strings.Join(outputLines, "\n")
	}

	return output, nil
}

// GetPaneActivity checks if a pane is currently active (focused).
// This is approximated by checking the pane info from the plugin.
func (c *Client) GetPaneActivity(ctx context.Context, session string, paneID uint32) (bool, error) {
	// Get pane info from plugin
	info, err := c.GetPaneInfo(ctx, session, paneID)
	if err != nil {
		return false, err
	}

	// A pane is considered active if it's focused
	// This is a heuristic - true activity detection would require
	// tracking output changes over time
	return info.IsFocused, nil
}
