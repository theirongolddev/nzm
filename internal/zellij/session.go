package zellij

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Session represents a Zellij session
type Session struct {
	Name      string
	Directory string
	Windows   int // Mapped from tabs
	Panes     []Pane
	Attached  bool
	Exited    bool
	Created   string // Not available in Zellij - kept for compatibility
}

// ListSessions returns all Zellij sessions
func (c *Client) ListSessions(ctx context.Context) ([]Session, error) {
	output, err := c.Run(ctx, "list-sessions")
	if err != nil {
		// No sessions is not an error
		errMsg := err.Error()
		if strings.Contains(errMsg, "no sessions") ||
			strings.Contains(errMsg, "No active") {
			return nil, nil
		}
		return nil, err
	}
	return parseSessionList(output)
}

// parseSessionList parses the output of `zellij list-sessions`
func parseSessionList(output string) ([]Session, error) {
	if output == "" {
		return nil, nil
	}

	var sessions []Session
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		session := Session{Name: line}

		// Check for (EXITED) suffix
		if strings.HasSuffix(line, " (EXITED)") {
			session.Name = strings.TrimSuffix(line, " (EXITED)")
			session.Exited = true
		}

		// Check for (current) suffix - indicates attached session
		if strings.HasSuffix(line, " (current)") {
			session.Name = strings.TrimSuffix(line, " (current)")
			session.Attached = true
		}

		sessions = append(sessions, session)
	}
	return sessions, nil
}

// SessionExists checks if a session with the given name exists
func (c *Client) SessionExists(ctx context.Context, name string) (bool, error) {
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return false, err
	}

	for _, s := range sessions {
		if s.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// GetSession returns detailed info about a session
func (c *Client) GetSession(ctx context.Context, name string) (*Session, error) {
	exists, err := c.SessionExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("session '%s' not found", name)
	}

	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return nil, err
	}

	var session *Session
	for _, s := range sessions {
		if s.Name == name {
			session = &s
			break
		}
	}
	if session == nil {
		return nil, fmt.Errorf("session '%s' not found", name)
	}

	// Get panes
	panes, err := c.GetPanes(ctx, name)
	if err != nil {
		return nil, err
	}
	session.Panes = panes

	return session, nil
}

// KillSession terminates a session
func (c *Client) KillSession(ctx context.Context, name string) error {
	return c.RunSilent(ctx, "kill-session", name)
}

// AttachSession attaches to an existing session
func (c *Client) AttachSession(ctx context.Context, name string) error {
	return c.RunSilent(ctx, "attach", name)
}

// CreateSession creates a new session with a layout
func (c *Client) CreateSession(ctx context.Context, name, layoutPath string) error {
	return c.RunSilent(ctx, "--session", name, "--layout", layoutPath)
}

// CreateSessionDetached creates a new session in the background.
// Zellij doesn't have a --detached flag, so we start the process in background.
func (c *Client) CreateSessionDetached(ctx context.Context, name, layoutPath string) error {
	cmd := exec.CommandContext(ctx, "zellij", "--session", name, "--layout", layoutPath)
	// Detach from terminal and process group
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting zellij session: %w", err)
	}
	
	// Don't wait for the process - let it run in background
	// Release the process so it's not killed when parent exits
	go func() {
		_ = cmd.Wait()
	}()
	
	// Wait briefly for session to be created
	time.Sleep(500 * time.Millisecond)
	return nil
}

// CreateSessionSimple creates a new session with just a directory (no layout)
func (c *Client) CreateSessionSimple(ctx context.Context, name, directory string) error {
	// Zellij doesn't support creating sessions without a layout like tmux does
	// We create a minimal layout on the fly
	opts := LayoutOptions{
		SessionName: name,
		ProjectDir:  directory,
	}
	layoutPath, err := WriteLayoutFile(opts)
	if err != nil {
		return err
	}
	defer os.Remove(layoutPath)
	return c.CreateSessionDetached(ctx, name, layoutPath)
}

// EnsureInstalled returns an error if zellij is not installed
func (c *Client) EnsureInstalled() error {
	if !c.IsInstalled() {
		return fmt.Errorf("zellij is not installed. Install it with: cargo install zellij")
	}
	return nil
}

// GetPanes returns all panes in a session
func (c *Client) GetPanes(ctx context.Context, session string) ([]Pane, error) {
	paneInfos, err := c.ListPanes(ctx, session)
	if err != nil {
		return nil, err
	}

	panes := make([]Pane, len(paneInfos))
	for i, info := range paneInfos {
		panes[i] = ConvertPaneInfo(info)
	}
	return panes, nil
}

// GetPanesEnriched returns all panes in a session with full agent metadata
func (c *Client) GetPanesEnriched(ctx context.Context, session string) ([]Pane, error) {
	return c.GetPanes(ctx, session)
}

// SplitPane creates a new pane in the session
func (c *Client) SplitPane(ctx context.Context, session, direction, directory string) (string, error) {
	// Use plugin to create a new pane
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "split_pane",
		Params: map[string]any{
			"direction": direction,
			"cwd":       directory,
		},
	})
	if err != nil {
		return "", err
	}
	if !resp.Success {
		return "", fmt.Errorf("%s", resp.Error)
	}

	// Get the new pane ID from response
	if paneID, ok := resp.Data["pane_id"].(float64); ok {
		return strconv.FormatUint(uint64(paneID), 10), nil
	}
	return "", nil
}

// SetPaneTitle sets the title of a pane via the plugin
func (c *Client) SetPaneTitle(ctx context.Context, session string, paneID string, title string) error {
	id, err := strconv.ParseUint(paneID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid pane ID: %w", err)
	}
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "set_pane_title",
		Params: map[string]any{
			"pane_id": uint32(id),
			"title":   title,
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

// GetPaneTitle returns the title of a pane
func (c *Client) GetPaneTitle(ctx context.Context, session, paneID string) (string, error) {
	id, err := strconv.ParseUint(paneID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid pane ID: %w", err)
	}
	info, err := c.GetPaneInfo(ctx, session, uint32(id))
	if err != nil {
		return "", err
	}
	return info.Title, nil
}

// GetPaneTags returns the tags for a pane parsed from its title.
func (c *Client) GetPaneTags(ctx context.Context, session, paneID string) ([]string, error) {
	title, err := c.GetPaneTitle(ctx, session, paneID)
	if err != nil {
		return nil, err
	}
	_, _, tags := parseAgentFromTitle(title)
	return tags, nil
}

// SetPaneTags sets the tags for a pane by updating its title.
func (c *Client) SetPaneTags(ctx context.Context, session, paneID string, tags []string) error {
	for _, tag := range tags {
		if strings.ContainsAny(tag, "[]") {
			return fmt.Errorf("tag %q contains invalid characters '[' or ']'", tag)
		}
	}

	title, err := c.GetPaneTitle(ctx, session, paneID)
	if err != nil {
		return err
	}

	baseTitle := stripTags(title)
	newTitle := baseTitle + FormatTags(tags)

	return c.SetPaneTitle(ctx, session, paneID, newTitle)
}

// AddPaneTags adds tags to a pane without removing existing ones.
func (c *Client) AddPaneTags(ctx context.Context, session, paneID string, newTags []string) error {
	existing, err := c.GetPaneTags(ctx, session, paneID)
	if err != nil {
		return err
	}

	tagSet := make(map[string]bool)
	for _, t := range existing {
		tagSet[t] = true
	}

	for _, t := range newTags {
		if !tagSet[t] {
			existing = append(existing, t)
			tagSet[t] = true
		}
	}

	return c.SetPaneTags(ctx, session, paneID, existing)
}

// RemovePaneTags removes specific tags from a pane.
func (c *Client) RemovePaneTags(ctx context.Context, session, paneID string, tagsToRemove []string) error {
	existing, err := c.GetPaneTags(ctx, session, paneID)
	if err != nil {
		return err
	}

	removeSet := make(map[string]bool)
	for _, t := range tagsToRemove {
		removeSet[t] = true
	}

	var filtered []string
	for _, t := range existing {
		if !removeSet[t] {
			filtered = append(filtered, t)
		}
	}

	return c.SetPaneTags(ctx, session, paneID, filtered)
}

// HasPaneTag returns true if the pane has the specified tag.
func (c *Client) HasPaneTag(ctx context.Context, session, paneID, tag string) (bool, error) {
	tags, err := c.GetPaneTags(ctx, session, paneID)
	if err != nil {
		return false, err
	}
	for _, t := range tags {
		if t == tag {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyPaneTag returns true if the pane has any of the specified tags.
func (c *Client) HasAnyPaneTag(ctx context.Context, session, paneID string, tags []string) (bool, error) {
	paneTags, err := c.GetPaneTags(ctx, session, paneID)
	if err != nil {
		return false, err
	}
	tagSet := make(map[string]bool)
	for _, t := range paneTags {
		tagSet[t] = true
	}
	for _, t := range tags {
		if tagSet[t] {
			return true, nil
		}
	}
	return false, nil
}

// KillPane kills a pane
func (c *Client) KillPane(ctx context.Context, session, paneID string) error {
	id, err := strconv.ParseUint(paneID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid pane ID: %w", err)
	}
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "close_pane",
		Params: map[string]any{
			"pane_id": uint32(id),
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

// IsAttached checks if a session is currently attached
func (c *Client) IsAttached(ctx context.Context, session string) (bool, error) {
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return false, err
	}

	for _, s := range sessions {
		if s.Name == session {
			return s.Attached, nil
		}
	}
	return false, nil
}

// ApplyTiledLayout applies a tiled layout to the session
func (c *Client) ApplyTiledLayout(ctx context.Context, session string) error {
	// Zellij doesn't have a direct equivalent, but we can use the action
	return c.RunSilent(ctx, "action", "toggle-floating-panes", "--session", session)
}

// ZoomPane zooms/focuses a specific pane
func (c *Client) ZoomPane(ctx context.Context, session string, paneIndex int) error {
	resp, err := c.SendPluginCommand(ctx, session, Request{
		Action: "focus_pane",
		Params: map[string]any{
			"pane_id": uint32(paneIndex),
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

// DisplayMessage shows a message (Zellij doesn't have direct equivalent)
func (c *Client) DisplayMessage(ctx context.Context, session, msg string, durationMs int) error {
	// Zellij doesn't have a direct message display like tmux
	// This is a no-op for compatibility
	return nil
}

// GetPanesWithActivity returns all panes with activity information
func (c *Client) GetPanesWithActivity(ctx context.Context, session string) ([]PaneActivity, error) {
	panes, err := c.GetPanes(ctx, session)
	if err != nil {
		return nil, err
	}

	result := make([]PaneActivity, len(panes))
	for i, pane := range panes {
		result[i] = PaneActivity{
			Pane:         pane,
			LastActivity: time.Now(), // Zellij doesn't track activity time
			IsActive:     pane.Active,
		}
	}
	return result, nil
}

// GetPaneActivity returns activity info for a specific pane
func (c *Client) GetPaneActivityTime(ctx context.Context, session, paneID string) (time.Time, error) {
	// Zellij doesn't track pane activity time like tmux
	// Return current time as approximation
	return time.Now(), nil
}

// GetPaneActivity returns whether a pane is currently active (focused).
// This matches the test API signature.
func (c *Client) GetPaneActivity(ctx context.Context, session string, paneID int) (bool, error) {
	info, err := c.GetPaneInfo(ctx, session, uint32(paneID))
	if err != nil {
		return false, err
	}
	return info.IsFocused, nil
}

// IsRecentlyActive checks if a pane has been recently active
func (c *Client) IsRecentlyActive(ctx context.Context, session, paneID string, threshold time.Duration) (bool, error) {
	// Since Zellij doesn't track activity time, check if pane is focused
	id, err := strconv.ParseUint(paneID, 10, 32)
	if err != nil {
		return false, fmt.Errorf("invalid pane ID: %w", err)
	}
	info, err := c.GetPaneInfo(ctx, session, uint32(id))
	if err != nil {
		return false, err
	}
	return info.IsFocused, nil
}

// GetPaneLastActivityAge returns how long ago the pane was active
func (c *Client) GetPaneLastActivityAge(ctx context.Context, session, paneID string) (time.Duration, error) {
	// Zellij doesn't track this, return 0 for "just now"
	return 0, nil
}

// GetCurrentSession returns the current session name
func (c *Client) GetCurrentSession() string {
	if !InZellij() {
		return ""
	}
	return os.Getenv("ZELLIJ_SESSION_NAME")
}

// validSessionNameRegex matches valid session names
var validSessionNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// ValidateSessionName checks if a session name is valid
func ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	if strings.Contains(name, ":") {
		return fmt.Errorf("session name cannot contain ':'")
	}
	if strings.Contains(name, ".") {
		return fmt.Errorf("session name cannot contain '.'")
	}

	if !validSessionNameRegex.MatchString(name) {
		return fmt.Errorf("invalid session name %q: must start with letter/number and contain only letters, numbers, dashes, underscores", name)
	}

	return nil
}

// ============== Package-level convenience functions ==============

// InZellij returns true if currently inside a Zellij session
func InZellij() bool {
	return os.Getenv("ZELLIJ") != ""
}

// IsInstalled checks if zellij is available
func IsInstalled() bool {
	return DefaultClient.IsInstalled()
}

// EnsureInstalled returns an error if zellij is not installed
func EnsureInstalled() error {
	return DefaultClient.EnsureInstalled()
}

// SessionExists checks if a session exists (default client)
func SessionExists(name string) bool {
	exists, _ := DefaultClient.SessionExists(context.Background(), name)
	return exists
}

// ListSessions returns all sessions (default client)
func ListSessions() ([]Session, error) {
	return DefaultClient.ListSessions(context.Background())
}

// GetSession returns session info (default client)
func GetSession(name string) (*Session, error) {
	return DefaultClient.GetSession(context.Background(), name)
}

// CreateSession creates a new session (default client)
func CreateSession(name, directory string) error {
	return DefaultClient.CreateSessionSimple(context.Background(), name, directory)
}

// GetPanes returns all panes in a session (default client)
func GetPanes(session string) ([]Pane, error) {
	return DefaultClient.GetPanes(context.Background(), session)
}

// GetPanesContext returns all panes with context (default client)
func GetPanesContext(ctx context.Context, session string) ([]Pane, error) {
	return DefaultClient.GetPanes(ctx, session)
}

// SplitWindow creates a new pane (default client) - tmux compatibility alias
func SplitWindow(session, directory string) (string, error) {
	return DefaultClient.SplitPane(context.Background(), session, "down", directory)
}

// SetPaneTitle sets pane title (default client)
func SetPaneTitle(paneID, title string) error {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.SetPaneTitle(context.Background(), session, paneID, title)
}

// GetPaneTitle gets pane title (default client)
func GetPaneTitle(paneID string) (string, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.GetPaneTitle(context.Background(), session, paneID)
}

// GetPaneTags gets pane tags (default client)
func GetPaneTags(paneID string) ([]string, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.GetPaneTags(context.Background(), session, paneID)
}

// SetPaneTags sets pane tags (default client)
func SetPaneTags(paneID string, tags []string) error {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.SetPaneTags(context.Background(), session, paneID, tags)
}

// AddPaneTags adds tags to pane (default client)
func AddPaneTags(paneID string, newTags []string) error {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.AddPaneTags(context.Background(), session, paneID, newTags)
}

// RemovePaneTags removes tags from pane (default client)
func RemovePaneTags(paneID string, tagsToRemove []string) error {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.RemovePaneTags(context.Background(), session, paneID, tagsToRemove)
}

// HasPaneTag checks if pane has tag (default client)
func HasPaneTag(paneID, tag string) (bool, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.HasPaneTag(context.Background(), session, paneID, tag)
}

// HasAnyPaneTag checks if pane has any of the tags (default client)
func HasAnyPaneTag(paneID string, tags []string) (bool, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.HasAnyPaneTag(context.Background(), session, paneID, tags)
}

// SendKeys sends keys to a pane (default client)
func SendKeys(target, keys string, enter bool) error {
	session := DefaultClient.GetCurrentSession()
	id, err := strconv.ParseUint(target, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid pane ID: %w", err)
	}
	return DefaultClient.SendKeys(context.Background(), session, uint32(id), keys, enter)
}

// PasteKeys pastes content to a pane (default client)
func PasteKeys(target, content string, enter bool) error {
	return SendKeys(target, content, enter)
}

// SendInterrupt sends Ctrl+C to a pane (default client)
func SendInterrupt(target string) error {
	session := DefaultClient.GetCurrentSession()
	id, err := strconv.ParseUint(target, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid pane ID: %w", err)
	}
	return DefaultClient.SendInterrupt(context.Background(), session, uint32(id))
}

// DisplayMessage shows a message (default client)
func DisplayMessage(session, msg string, durationMs int) error {
	return DefaultClient.DisplayMessage(context.Background(), session, msg, durationMs)
}

// AttachOrSwitch attaches to a session (default client)
func AttachOrSwitch(session string) error {
	return DefaultClient.AttachSession(context.Background(), session)
}

// KillSession kills a session (default client)
func KillSession(session string) error {
	return DefaultClient.KillSession(context.Background(), session)
}

// KillPane kills a pane (default client)
func KillPane(paneID string) error {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.KillPane(context.Background(), session, paneID)
}

// ApplyTiledLayout applies tiled layout (default client)
func ApplyTiledLayout(session string) error {
	return DefaultClient.ApplyTiledLayout(context.Background(), session)
}

// ZoomPane zooms a pane (default client)
func ZoomPane(session string, paneIndex int) error {
	return DefaultClient.ZoomPane(context.Background(), session, paneIndex)
}

// CapturePaneOutput captures pane output (default client)
func CapturePaneOutput(target string, lines int) (string, error) {
	session := DefaultClient.GetCurrentSession()
	id, err := strconv.ParseUint(target, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid pane ID: %w", err)
	}
	return DefaultClient.CapturePaneOutput(context.Background(), session, uint32(id), lines)
}

// CapturePaneOutputContext captures pane output with context (default client)
func CapturePaneOutputContext(ctx context.Context, target string, lines int) (string, error) {
	session := DefaultClient.GetCurrentSession()
	id, err := strconv.ParseUint(target, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid pane ID: %w", err)
	}
	return DefaultClient.CapturePaneOutput(ctx, session, uint32(id), lines)
}

// GetCurrentSession returns current session name (default client)
func GetCurrentSession() string {
	return DefaultClient.GetCurrentSession()
}

// GetPaneActivity returns pane activity time (default client)
func GetPaneActivity(paneID string) (time.Time, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.GetPaneActivityTime(context.Background(), session, paneID)
}

// GetPanesWithActivity returns panes with activity (default client)
func GetPanesWithActivity(session string) ([]PaneActivity, error) {
	return DefaultClient.GetPanesWithActivity(context.Background(), session)
}

// GetPanesWithActivityContext returns panes with activity and context (default client)
func GetPanesWithActivityContext(ctx context.Context, session string) ([]PaneActivity, error) {
	return DefaultClient.GetPanesWithActivity(ctx, session)
}

// IsRecentlyActive checks if pane is recently active (default client)
func IsRecentlyActive(paneID string, threshold time.Duration) (bool, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.IsRecentlyActive(context.Background(), session, paneID, threshold)
}

// GetPaneLastActivityAge returns pane activity age (default client)
func GetPaneLastActivityAge(paneID string) (time.Duration, error) {
	session := DefaultClient.GetCurrentSession()
	return DefaultClient.GetPaneLastActivityAge(context.Background(), session, paneID)
}

// IsAttached checks if session is attached (default client)
func IsAttached(session string) bool {
	attached, _ := DefaultClient.IsAttached(context.Background(), session)
	return attached
}

// GetFirstWindow returns first window/tab index (compatibility)
func GetFirstWindow(session string) (int, error) {
	// Zellij uses tabs, return 0 as default
	return 0, nil
}

// GetDefaultPaneIndex returns default pane index (compatibility)
func GetDefaultPaneIndex(session string) (int, error) {
	panes, err := GetPanes(session)
	if err != nil {
		return 0, err
	}
	if len(panes) == 0 {
		return 0, fmt.Errorf("no panes found")
	}
	return panes[0].Index, nil
}
