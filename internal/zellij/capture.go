package zellij

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ============== Optimized Capture Helpers ==============
//
// These helpers provide semantic capture operations with appropriate line budgets
// to reduce latency and CPU usage.

// Capture line budget constants for consistent usage across codebase.
const (
	// LinesStatusDetection is the default for quick status/state detection.
	LinesStatusDetection = 20

	// LinesHealthCheck is the default for health and error analysis.
	LinesHealthCheck = 50

	// LinesFullContext is the default for comprehensive context capture.
	LinesFullContext = 500

	// LinesCheckpoint is for session checkpoint/restore operations.
	LinesCheckpoint = 2000
)

// CapturePaneOutput captures the output from a pane.
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

// CaptureForStatusDetection captures minimal output for quick state detection.
func (c *Client) CaptureForStatusDetection(ctx context.Context, session string, paneID uint32) (string, error) {
	return c.CapturePaneOutput(ctx, session, paneID, LinesStatusDetection)
}

// CaptureForHealthCheck captures output for health analysis.
func (c *Client) CaptureForHealthCheck(ctx context.Context, session string, paneID uint32) (string, error) {
	return c.CapturePaneOutput(ctx, session, paneID, LinesHealthCheck)
}

// CaptureForFullContext captures comprehensive output for analysis.
func (c *Client) CaptureForFullContext(ctx context.Context, session string, paneID uint32) (string, error) {
	return c.CapturePaneOutput(ctx, session, paneID, LinesFullContext)
}

// CaptureForCheckpoint captures maximum output for checkpoint operations.
func (c *Client) CaptureForCheckpoint(ctx context.Context, session string, paneID uint32) (string, error) {
	return c.CapturePaneOutput(ctx, session, paneID, LinesCheckpoint)
}

// ============== Package-level capture functions ==============

// CaptureForStatusDetection captures for status detection (default client)
func CaptureForStatusDetection(target string) (string, error) {
	return CaptureForStatusDetectionContext(context.Background(), target)
}

// CaptureForStatusDetectionContext captures with context support
func CaptureForStatusDetectionContext(ctx context.Context, target string) (string, error) {
	session := DefaultClient.GetCurrentSession()
	id, err := parseTargetID(target)
	if err != nil {
		return "", err
	}
	return DefaultClient.CaptureForStatusDetection(ctx, session, id)
}

// CaptureForHealthCheck captures for health checks (default client)
func CaptureForHealthCheck(target string) (string, error) {
	return CaptureForHealthCheckContext(context.Background(), target)
}

// CaptureForHealthCheckContext captures with context support
func CaptureForHealthCheckContext(ctx context.Context, target string) (string, error) {
	session := DefaultClient.GetCurrentSession()
	id, err := parseTargetID(target)
	if err != nil {
		return "", err
	}
	return DefaultClient.CaptureForHealthCheck(ctx, session, id)
}

// CaptureForFullContext captures full context (default client)
func CaptureForFullContext(target string) (string, error) {
	return CaptureForFullContextContext(context.Background(), target)
}

// CaptureForFullContextContext captures with context support
func CaptureForFullContextContext(ctx context.Context, target string) (string, error) {
	session := DefaultClient.GetCurrentSession()
	id, err := parseTargetID(target)
	if err != nil {
		return "", err
	}
	return DefaultClient.CaptureForFullContext(ctx, session, id)
}

// CaptureForCheckpoint captures for checkpoints (default client)
func CaptureForCheckpoint(target string) (string, error) {
	return CaptureForCheckpointContext(context.Background(), target)
}

// CaptureForCheckpointContext captures with context support
func CaptureForCheckpointContext(ctx context.Context, target string) (string, error) {
	session := DefaultClient.GetCurrentSession()
	id, err := parseTargetID(target)
	if err != nil {
		return "", err
	}
	return DefaultClient.CaptureForCheckpoint(ctx, session, id)
}

// parseTargetID extracts pane ID from target string
func parseTargetID(target string) (uint32, error) {
	// Target can be just the pane ID or session:window.pane format
	// For simplicity, assume it's just the pane ID
	var id uint64
	_, err := fmt.Sscanf(target, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid target: %s", target)
	}
	return uint32(id), nil
}
