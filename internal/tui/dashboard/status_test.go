package dashboard

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Dicklesworthstone/ntm/internal/status"
	"github.com/Dicklesworthstone/ntm/internal/tmux"
	"github.com/Dicklesworthstone/ntm/internal/tui/layout"
)

func TestStatusUpdateSetsPaneStateAndTimestamp(t *testing.T) {
	t.Parallel()

	m := New("session", "")
	m.panes = []tmux.Pane{
		{ID: "%1", Index: 0, Title: "session__cod_1", Type: tmux.AgentCodex},
	}
	m.paneStatus[0] = PaneStatus{}

	now := time.Now()
	msg := StatusUpdateMsg{
		Statuses: []status.AgentStatus{
			{PaneID: "%1", State: status.StateIdle, UpdatedAt: now},
		},
		Time: now,
	}

	updated, _ := m.Update(msg)
	m2 := updated.(Model)

	if got := m2.paneStatus[0].State; got != "idle" {
		t.Fatalf("expected pane state idle, got %q", got)
	}
	if m2.lastRefresh.IsZero() {
		t.Fatalf("expected lastRefresh to be set")
	}
}

func TestViewShowsLoadingWhenSessionFetchInFlight(t *testing.T) {
	t.Parallel()

	m := New("session", "")
	m.width = 80
	m.height = 30
	m.tier = layout.TierForWidth(m.width)

	// Simulate initial load: no panes yet, fetch in flight.
	m.panes = nil
	m.err = nil
	m.fetchingSession = true
	m.lastPaneFetch = time.Now().Add(-750 * time.Millisecond)

	plain := status.StripANSI(m.View())
	if strings.Contains(plain, "No panes found") {
		t.Fatalf("expected loading state, got empty state")
	}
	if !strings.Contains(plain, "Fetching panes") {
		t.Fatalf("expected loading copy to mention fetching panes")
	}
}

func TestViewRendersPanesEvenWhenLastSessionFetchErrored(t *testing.T) {
	t.Parallel()

	m := New("session", "")
	m.width = 120
	m.height = 30
	m.tier = layout.TierForWidth(m.width)

	m.panes = []tmux.Pane{
		{ID: "%1", Index: 0, Title: "session__cod_1", Type: tmux.AgentCodex},
	}
	m.err = errors.New("tmux refresh failed")

	plain := status.StripANSI(m.View())
	if !strings.Contains(plain, "tmux refresh failed") {
		t.Fatalf("expected error to be surfaced in view")
	}
	if !strings.Contains(plain, "session__cod_1") {
		t.Fatalf("expected panes to still render when session fetch errors")
	}
}
