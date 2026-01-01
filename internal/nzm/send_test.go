package nzm

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/theirongolddev/nzm/internal/zellij"
)

// mockPluginClient implements the PluginClient interface for testing
type mockPluginClient struct {
	panes       []zellij.PaneInfo
	listErr     error
	sendKeysErr error
	interruptErr error
	sentText    string
	sentPaneID  uint32
	sentEnter   bool
}

func (m *mockPluginClient) ListPanes(ctx context.Context, session string) ([]zellij.PaneInfo, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.panes, nil
}

func (m *mockPluginClient) SendKeys(ctx context.Context, session string, paneID uint32, text string, enter bool) error {
	m.sentPaneID = paneID
	m.sentText = text
	m.sentEnter = enter
	return m.sendKeysErr
}

func (m *mockPluginClient) SendInterrupt(ctx context.Context, session string, paneID uint32) error {
	m.sentPaneID = paneID
	return m.interruptErr
}

func TestSendOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    SendOptions
		wantErr bool
	}{
		{
			name: "valid with target and text",
			opts: SendOptions{
				Session: "proj",
				Target:  "cc_1",
				Text:    "hello",
			},
			wantErr: false,
		},
		{
			name: "empty session",
			opts: SendOptions{
				Session: "",
				Target:  "cc_1",
				Text:    "hello",
			},
			wantErr: true,
		},
		{
			name: "empty target",
			opts: SendOptions{
				Session: "proj",
				Target:  "",
				Text:    "hello",
			},
			wantErr: true,
		},
		{
			name: "empty text",
			opts: SendOptions{
				Session: "proj",
				Target:  "cc_1",
				Text:    "",
			},
			wantErr: true,
		},
		{
			name: "interrupt without text is valid",
			opts: SendOptions{
				Session:   "proj",
				Target:    "cc_1",
				Interrupt: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSend_ByPaneName(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
			{ID: 2, Title: "proj__cc_2"},
		},
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session: "proj",
		Target:  "cc_1",
		Text:    "hello world",
		Enter:   true,
	}

	err := sender.Send(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.sentPaneID != 1 {
		t.Errorf("expected pane ID 1, got %d", mock.sentPaneID)
	}
	if mock.sentText != "hello world" {
		t.Errorf("expected text 'hello world', got %q", mock.sentText)
	}
	if !mock.sentEnter {
		t.Error("expected enter=true")
	}
}

func TestSend_ByAgentType(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
			{ID: 2, Title: "proj__gmi_1"},
		},
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session: "proj",
		Target:  "gmi",  // Should match gmi_1
		Text:    "test",
	}

	err := sender.Send(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.sentPaneID != 2 {
		t.Errorf("expected pane ID 2 (gmi_1), got %d", mock.sentPaneID)
	}
}

func TestSend_TargetNotFound(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
		},
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session: "proj",
		Target:  "nonexistent",
		Text:    "test",
	}

	err := sender.Send(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for nonexistent target")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestSend_Interrupt(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
		},
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session:   "proj",
		Target:    "cc_1",
		Interrupt: true,
	}

	err := sender.Send(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.sentPaneID != 1 {
		t.Errorf("expected pane ID 1, got %d", mock.sentPaneID)
	}
}

func TestSend_ListPanesError(t *testing.T) {
	mock := &mockPluginClient{
		listErr: errors.New("plugin not responding"),
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session: "proj",
		Target:  "cc_1",
		Text:    "test",
	}

	err := sender.Send(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSend_SendKeysError(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
		},
		sendKeysErr: errors.New("pane busy"),
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session: "proj",
		Target:  "cc_1",
		Text:    "test",
	}

	err := sender.Send(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSend_ByFullPaneName(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
			{ID: 2, Title: "proj__cc_2"},
		},
	}
	sender := NewSender(mock)

	opts := SendOptions{
		Session: "proj",
		Target:  "proj__cc_2",  // Full pane name
		Text:    "test",
	}

	err := sender.Send(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.sentPaneID != 2 {
		t.Errorf("expected pane ID 2, got %d", mock.sentPaneID)
	}
}

func TestSend_MultipleAgentsOfSameType(t *testing.T) {
	mock := &mockPluginClient{
		panes: []zellij.PaneInfo{
			{ID: 1, Title: "proj__cc_1"},
			{ID: 2, Title: "proj__cc_2"},
			{ID: 3, Title: "proj__cc_3"},
		},
	}
	sender := NewSender(mock)

	// Just "cc" should match first cc pane
	opts := SendOptions{
		Session: "proj",
		Target:  "cc",
		Text:    "test",
	}

	err := sender.Send(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.sentPaneID != 1 {
		t.Errorf("expected first cc pane (ID 1), got %d", mock.sentPaneID)
	}
}
