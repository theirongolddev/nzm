package zellij

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

// Request is a command sent to the nzm-agent plugin
type Request struct {
	ID     string         `json:"id"`
	Action string         `json:"action"`
	Params map[string]any `json:"params,omitempty"`
}

// Response is the reply from the nzm-agent plugin
type Response struct {
	ID      string         `json:"id"`
	Success bool           `json:"success"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// PaneInfo represents a terminal pane
type PaneInfo struct {
	ID         uint32 `json:"id"`
	Title      string `json:"title"`
	IsFocused  bool   `json:"is_focused"`
	IsFloating bool   `json:"is_floating"`
}

// ParseResponse parses a JSON response from the plugin
func ParseResponse(jsonStr string) (*Response, error) {
	var resp Response
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse plugin response: %w", err)
	}
	return &resp, nil
}

// GetPanes extracts pane information from the response data
func (r *Response) GetPanes() ([]PaneInfo, error) {
	if r.Data == nil {
		return nil, nil
	}

	panesRaw, ok := r.Data["panes"]
	if !ok {
		return nil, nil
	}

	// Re-marshal and unmarshal to convert to []PaneInfo
	panesJSON, err := json.Marshal(panesRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal panes: %w", err)
	}

	var panes []PaneInfo
	if err := json.Unmarshal(panesJSON, &panes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal panes: %w", err)
	}

	return panes, nil
}

// PluginPath is the default path to the nzm-agent plugin
const PluginPath = "nzm-agent"

// requestCounter for generating unique IDs
var requestCounter uint64

// GenerateRequestID creates a unique request ID
func GenerateRequestID() string {
	counter := atomic.AddUint64(&requestCounter, 1)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), counter)
}

// SendPluginCommand sends a command to the nzm-agent plugin and waits for response
func (c *Client) SendPluginCommand(ctx context.Context, session string, req Request) (*Response, error) {
	// Assign ID if not set
	if req.ID == "" {
		req.ID = GenerateRequestID()
	}

	// Serialize request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call: zellij --session <session> pipe --plugin nzm-agent -- <json>
	output, err := c.Run(ctx,
		"--session", session,
		"pipe",
		"--plugin", PluginPath,
		"--",
		string(reqJSON),
	)
	if err != nil {
		return nil, err
	}

	// Parse response
	return ParseResponse(output)
}

// ListPanes returns all terminal panes in a session
func (c *Client) ListPanes(ctx context.Context, session string) ([]PaneInfo, error) {
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "list_panes",
	})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	return resp.GetPanes()
}

// SendKeys sends text to a specific pane
func (c *Client) SendKeys(ctx context.Context, session string, paneID uint32, text string, enter bool) error {
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "send_keys",
		Params: map[string]any{
			"pane_id": paneID,
			"text":    text,
			"enter":   enter,
		},
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

// SendInterrupt sends Ctrl+C to a specific pane
func (c *Client) SendInterrupt(ctx context.Context, session string, paneID uint32) error {
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "send_interrupt",
		Params: map[string]any{
			"pane_id": paneID,
		},
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.Error)
	}

	return nil
}

// GetPaneInfo gets information about a specific pane
func (c *Client) GetPaneInfo(ctx context.Context, session string, paneID uint32) (*PaneInfo, error) {
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "get_pane_info",
		Params: map[string]any{
			"pane_id": paneID,
		},
	})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	// Extract pane from response
	if resp.Data == nil {
		return nil, fmt.Errorf("no pane data in response")
	}

	paneRaw, ok := resp.Data["pane"]
	if !ok {
		return nil, fmt.Errorf("no pane in response data")
	}

	paneJSON, err := json.Marshal(paneRaw)
	if err != nil {
		return nil, err
	}

	var pane PaneInfo
	if err := json.Unmarshal(paneJSON, &pane); err != nil {
		return nil, err
	}

	return &pane, nil
}
