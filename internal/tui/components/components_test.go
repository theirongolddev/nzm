package components

import (
	"strings"
	"testing"
)

// Banner tests
func TestRenderBanner(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		result := RenderBanner(false, 0)
		if result == "" {
			t.Error("RenderBanner should return non-empty string")
		}
	})

	t.Run("animated", func(t *testing.T) {
		result := RenderBanner(true, 5)
		if result == "" {
			t.Error("RenderBanner animated should return non-empty string")
		}
	})
}

func TestRenderBannerMedium(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		result := RenderBannerMedium(false, 0)
		if result == "" {
			t.Error("RenderBannerMedium should return non-empty string")
		}
	})

	t.Run("animated", func(t *testing.T) {
		result := RenderBannerMedium(true, 5)
		if result == "" {
			t.Error("RenderBannerMedium animated should return non-empty string")
		}
	})
}

func TestRenderInlineLogo(t *testing.T) {
	result := RenderInlineLogo()
	if result == "" {
		t.Error("RenderInlineLogo should return non-empty string")
	}
}

func TestRenderSubtitle(t *testing.T) {
	result := RenderSubtitle("Test subtitle")
	if result == "" {
		t.Error("RenderSubtitle should return non-empty string")
	}
}

func TestRenderVersion(t *testing.T) {
	result := RenderVersion("1.0.0")
	if !strings.Contains(result, "v") {
		t.Error("RenderVersion should contain version prefix")
	}
}

func TestRenderHeaderBar(t *testing.T) {
	result := RenderHeaderBar("Test Title", 40)
	if result == "" {
		t.Error("RenderHeaderBar should return non-empty string")
	}
}

func TestRenderSection(t *testing.T) {
	result := RenderSection("Section", 40)
	if result == "" {
		t.Error("RenderSection should return non-empty string")
	}

	// Test short width
	result = RenderSection("Very Long Section Title Here", 10)
	if result == "" {
		t.Error("RenderSection should handle short width")
	}
}

func TestRenderAgentBadge(t *testing.T) {
	tests := []string{"claude", "cc", "codex", "cod", "gemini", "gmi", "unknown"}

	for _, agent := range tests {
		t.Run(agent, func(t *testing.T) {
			result := RenderAgentBadge(agent)
			if result == "" {
				t.Errorf("RenderAgentBadge(%q) should return non-empty string", agent)
			}
		})
	}
}

func TestRenderStatusBadge(t *testing.T) {
	tests := []string{"running", "active", "idle", "error", "failed", "success", "done", "unknown"}

	for _, status := range tests {
		t.Run(status, func(t *testing.T) {
			result := RenderStatusBadge(status)
			if result == "" {
				t.Errorf("RenderStatusBadge(%q) should return non-empty string", status)
			}
		})
	}
}

func TestRenderKeyMap(t *testing.T) {
	keys := map[string]string{
		"q": "quit",
		"?": "help",
	}
	result := RenderKeyMap(keys, 60)
	if result == "" {
		t.Error("RenderKeyMap should return non-empty string")
	}
}

func TestRenderFooter(t *testing.T) {
	result := RenderFooter("Footer text", 40)
	if result == "" {
		t.Error("RenderFooter should return non-empty string")
	}
}

func TestRenderHint(t *testing.T) {
	result := RenderHint("Hint text")
	if result == "" {
		t.Error("RenderHint should return non-empty string")
	}
}

func TestRenderHighlight(t *testing.T) {
	result := RenderHighlight("Highlighted")
	if result == "" {
		t.Error("RenderHighlight should return non-empty string")
	}
}

func TestRenderCommand(t *testing.T) {
	result := RenderCommand("ntm")
	if result == "" {
		t.Error("RenderCommand should return non-empty string")
	}
}

func TestRenderArg(t *testing.T) {
	result := RenderArg("session")
	if !strings.Contains(result, "<") {
		t.Error("RenderArg should wrap in angle brackets")
	}
}

func TestRenderFlag(t *testing.T) {
	result := RenderFlag("--help")
	if result == "" {
		t.Error("RenderFlag should return non-empty string")
	}
}

func TestRenderExample(t *testing.T) {
	result := RenderExample("ntm spawn myproject")
	if result == "" {
		t.Error("RenderExample should return non-empty string")
	}
}

// Box tests
func TestNewBox(t *testing.T) {
	box := NewBox()
	if box == nil {
		t.Error("NewBox should return non-nil box")
	}
	if box.Style != BoxRounded {
		t.Error("NewBox should default to BoxRounded")
	}
	if box.Padding != 1 {
		t.Error("NewBox should default padding to 1")
	}
}

func TestBoxBuilderPattern(t *testing.T) {
	box := NewBox().
		WithTitle("Test").
		WithContent("Content").
		WithSize(40, 10).
		WithStyle(BoxDouble).
		WithPadding(2)

	if box.Title != "Test" {
		t.Error("WithTitle should set title")
	}
	if box.Content != "Content" {
		t.Error("WithContent should set content")
	}
	if box.Width != 40 {
		t.Error("WithSize should set width")
	}
	if box.Height != 10 {
		t.Error("WithSize should set height")
	}
	if box.Style != BoxDouble {
		t.Error("WithStyle should set style")
	}
	if box.Padding != 2 {
		t.Error("WithPadding should set padding")
	}
}

func TestBoxRender(t *testing.T) {
	t.Run("simple box", func(t *testing.T) {
		box := NewBox().WithContent("Hello")
		result := box.Render()
		if result == "" {
			t.Error("Box.Render should return non-empty string")
		}
	})

	t.Run("box with title", func(t *testing.T) {
		box := NewBox().
			WithTitle("Title").
			WithContent("Content").
			WithSize(30, 0)
		result := box.Render()
		if result == "" {
			t.Error("Box with title should render")
		}
	})

	t.Run("all styles", func(t *testing.T) {
		styles := []BoxStyle{BoxRounded, BoxDouble, BoxThick, BoxNormal, BoxHidden}
		for _, style := range styles {
			box := NewBox().WithContent("Test").WithStyle(style)
			result := box.Render()
			if result == "" {
				t.Errorf("BoxStyle %d should render", style)
			}
		}
	})
}

func TestBoxString(t *testing.T) {
	box := NewBox().WithContent("Test")
	if box.String() != box.Render() {
		t.Error("Box.String() should equal Box.Render()")
	}
}

func TestSimpleBox(t *testing.T) {
	result := SimpleBox("Title", "Content", 40)
	if result == "" {
		t.Error("SimpleBox should return non-empty string")
	}
}

func TestInfoBox(t *testing.T) {
	result := InfoBox("Info", "Information", 40)
	if result == "" {
		t.Error("InfoBox should return non-empty string")
	}
}

func TestSuccessBox(t *testing.T) {
	result := SuccessBox("Success", "Done!", 40)
	if result == "" {
		t.Error("SuccessBox should return non-empty string")
	}
}

func TestErrorBox(t *testing.T) {
	result := ErrorBox("Error", "Failed!", 40)
	if result == "" {
		t.Error("ErrorBox should return non-empty string")
	}
}

func TestWarningBox(t *testing.T) {
	result := WarningBox("Warning", "Be careful!", 40)
	if result == "" {
		t.Error("WarningBox should return non-empty string")
	}
}

func TestDivider(t *testing.T) {
	result := Divider(20)
	if result == "" {
		t.Error("Divider should return non-empty string")
	}
}

func TestDoubleDivider(t *testing.T) {
	result := DoubleDivider(20)
	if result == "" {
		t.Error("DoubleDivider should return non-empty string")
	}
}

func TestThinDivider(t *testing.T) {
	result := ThinDivider(20)
	if result == "" {
		t.Error("ThinDivider should return non-empty string")
	}
}

func TestLabeledDivider(t *testing.T) {
	t.Run("normal width", func(t *testing.T) {
		result := LabeledDivider("Section", 40)
		if result == "" {
			t.Error("LabeledDivider should return non-empty string")
		}
	})

	t.Run("narrow width", func(t *testing.T) {
		result := LabeledDivider("Very Long Label", 10)
		if result == "" {
			t.Error("LabeledDivider should handle narrow width")
		}
	})
}

// Spinner tests
func TestNewSpinner(t *testing.T) {
	s := NewSpinner()
	if s.Style != SpinnerDots {
		t.Error("NewSpinner should default to SpinnerDots")
	}
	if len(s.GradientColors) == 0 {
		t.Error("NewSpinner should have gradient colors")
	}
}

func TestSpinnerView(t *testing.T) {
	s := NewSpinner()
	result := s.View()
	if result == "" {
		t.Error("Spinner.View should return non-empty string")
	}
}

func TestSpinnerViewWithLabel(t *testing.T) {
	s := NewSpinner()
	s.Label = "Loading..."
	result := s.View()
	if !strings.Contains(result, "Loading") {
		t.Error("Spinner with label should include label")
	}
}

func TestSpinnerViewWithGradient(t *testing.T) {
	s := NewSpinner()
	s.Gradient = true
	result := s.View()
	if result == "" {
		t.Error("Spinner with gradient should render")
	}
}

func TestSpinnerStyles(t *testing.T) {
	styles := []SpinnerStyle{
		SpinnerDots,
		SpinnerLine,
		SpinnerBounce,
		SpinnerPoints,
		SpinnerGlobe,
		SpinnerMoon,
		SpinnerMonkey,
		SpinnerMeter,
		SpinnerHamburger,
	}

	for _, style := range styles {
		t.Run("", func(t *testing.T) {
			s := NewSpinner()
			s.Style = style
			result := s.View()
			if result == "" {
				t.Errorf("SpinnerStyle %d should render", style)
			}
		})
	}
}

func TestSpinnerUpdate(t *testing.T) {
	s := NewSpinner()
	initialFrame := s.Frame

	updated, cmd := s.Update(SpinnerTickMsg{})
	if updated.Frame == initialFrame {
		// Frame should advance
		if updated.Frame != 1 {
			t.Error("Spinner frame should advance on tick")
		}
	}
	if cmd == nil {
		t.Error("Spinner.Update should return a command")
	}
}

func TestSpinnerUpdateUnknownMsg(t *testing.T) {
	s := NewSpinner()
	initialFrame := s.Frame

	updated, cmd := s.Update("unknown message")
	if updated.Frame != initialFrame {
		t.Error("Spinner should not update on unknown message")
	}
	if cmd != nil {
		t.Error("Spinner should not return command on unknown message")
	}
}

func TestSpinnerTickCmd(t *testing.T) {
	s := NewSpinner()
	cmd := s.TickCmd()
	if cmd == nil {
		t.Error("TickCmd should return a command")
	}
}

func TestSpinnerInit(t *testing.T) {
	s := NewSpinner()
	cmd := s.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

// Logo constants tests
func TestLogoConstants(t *testing.T) {
	if len(LogoLarge) == 0 {
		t.Error("LogoLarge should not be empty")
	}
	if len(LogoMedium) == 0 {
		t.Error("LogoMedium should not be empty")
	}
	if LogoSmall == "" {
		t.Error("LogoSmall should not be empty")
	}
	if LogoIcon == "" {
		t.Error("LogoIcon should not be empty")
	}
	if LogoIconPlain == "" {
		t.Error("LogoIconPlain should not be empty")
	}
}

// List tests
func TestNewList(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Description: "First item"},
		{Title: "Item 2", Description: "Second item"},
	}
	list := NewList(items)
	if list == nil {
		t.Error("NewList should return non-nil list")
	}
	if len(list.Items) != 2 {
		t.Error("NewList should have 2 items")
	}
}

func TestListBuilders(t *testing.T) {
	items := []ListItem{{Title: "Item"}}
	list := NewList(items).
		WithCursor("→").
		WithMaxVisible(5).
		WithWidth(40).
		WithNumbers(true).
		WithIcons(true)

	if list.Cursor != "→" {
		t.Error("WithCursor should set cursor")
	}
	if list.MaxVisible != 5 {
		t.Error("WithMaxVisible should set max visible")
	}
	if list.Width != 40 {
		t.Error("WithWidth should set width")
	}
	if !list.ShowNumbers {
		t.Error("WithNumbers should set show numbers")
	}
	if !list.ShowIcons {
		t.Error("WithIcons should set show icons")
	}
}

func TestListNavigation(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	}
	list := NewList(items)

	// Move down
	list.MoveDown()
	if list.Selected != 1 {
		t.Error("MoveDown should move selection")
	}

	// Move up
	list.MoveUp()
	if list.Selected != 0 {
		t.Error("MoveUp should move selection")
	}

	// Move up at top (should wrap or stay)
	list.MoveUp()
	// Depending on implementation, check behavior
}

func TestListSelectByNumber(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
		{Title: "Item 3"},
	}
	list := NewList(items)

	list.SelectByNumber(2)
	if list.Selected != 1 {
		t.Errorf("SelectByNumber(2) should select index 1, got %d", list.Selected)
	}

	// Out of range
	list.SelectByNumber(99)
	// Should not change
}

func TestListSelectedItem(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1"},
		{Title: "Item 2"},
	}
	list := NewList(items)

	item := list.SelectedItem()
	if item == nil || item.Title != "Item 1" {
		t.Error("SelectedItem should return first item")
	}
}

func TestListRender(t *testing.T) {
	items := []ListItem{
		{Title: "Item 1", Description: "Desc 1"},
		{Title: "Item 2", Description: "Desc 2"},
	}
	list := NewList(items)
	result := list.Render()
	if result == "" {
		t.Error("List.Render should return non-empty string")
	}
}

func TestListString(t *testing.T) {
	items := []ListItem{{Title: "Item"}}
	list := NewList(items)
	if list.String() != list.Render() {
		t.Error("List.String should equal List.Render")
	}
}

// Preview tests
func TestNewPreview(t *testing.T) {
	p := NewPreview()
	if p == nil {
		t.Error("NewPreview should return non-nil preview")
	}
}

func TestPreviewBuilders(t *testing.T) {
	p := NewPreview().
		WithTitle("Title").
		WithContent("Content").
		WithSize(40, 10).
		WithBorder(true)

	if p.Title != "Title" {
		t.Error("WithTitle should set title")
	}
	if p.Content != "Content" {
		t.Error("WithContent should set content")
	}
	if p.Width != 40 {
		t.Error("WithSize should set width")
	}
	if p.Height != 10 {
		t.Error("WithSize should set height")
	}
	if !p.ShowBorder {
		t.Error("WithBorder should set border")
	}
}

func TestPreviewRender(t *testing.T) {
	p := NewPreview().
		WithTitle("Preview").
		WithContent("Some content").
		WithSize(40, 10)

	result := p.Render()
	if result == "" {
		t.Error("Preview.Render should return non-empty string")
	}
}

func TestPreviewString(t *testing.T) {
	p := NewPreview().WithContent("Test")
	if p.String() != p.Render() {
		t.Error("Preview.String should equal Preview.Render")
	}
}

// StatusBar tests
func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar(80)
	if sb == nil {
		t.Error("NewStatusBar should return non-nil")
	}
}

func TestStatusBarBuilders(t *testing.T) {
	sb := NewStatusBar(80).
		WithLeft("Left").
		WithCenter("Center").
		WithRight("Right")

	if sb.Left != "Left" {
		t.Error("WithLeft should set left")
	}
	if sb.Center != "Center" {
		t.Error("WithCenter should set center")
	}
	if sb.Right != "Right" {
		t.Error("WithRight should set right")
	}
}

func TestStatusBarRender(t *testing.T) {
	sb := NewStatusBar(80).
		WithLeft("File: test.go").
		WithCenter("Line 42").
		WithRight("UTF-8")

	result := sb.Render()
	if result == "" {
		t.Error("StatusBar.Render should return non-empty string")
	}
}

// HelpBar tests
func TestNewHelpBar(t *testing.T) {
	hb := NewHelpBar(80)
	if hb == nil {
		t.Error("NewHelpBar should return non-nil")
	}
}

func TestHelpBarAdd(t *testing.T) {
	hb := NewHelpBar(80)
	hb.Add("q", "quit")
	hb.Add("?", "help")

	if len(hb.Items) != 2 {
		t.Error("Add should add items")
	}
}

func TestHelpBarRender(t *testing.T) {
	hb := NewHelpBar(80)
	hb.Add("q", "quit")
	hb.Add("?", "help")

	result := hb.Render()
	if result == "" {
		t.Error("HelpBar.Render should return non-empty string")
	}
}

func TestHelpBarString(t *testing.T) {
	hb := NewHelpBar(80)
	hb.Add("q", "quit")
	if hb.String() != hb.Render() {
		t.Error("HelpBar.String should equal HelpBar.Render")
	}
}

// Header tests
func TestNewHeader(t *testing.T) {
	h := NewHeader("Title", 80)
	if h == nil {
		t.Error("NewHeader should return non-nil")
	}
}

func TestHeaderBuilders(t *testing.T) {
	h := NewHeader("Title", 80).
		WithSubtitle("Subtitle").
		WithIcon("󰆍")

	if h.Subtitle != "Subtitle" {
		t.Error("WithSubtitle should set subtitle")
	}
	if h.Icon != "󰆍" {
		t.Error("WithIcon should set icon")
	}
}

func TestHeaderRender(t *testing.T) {
	h := NewHeader("Test Header", 80).
		WithSubtitle("A subtitle")

	result := h.Render()
	if result == "" {
		t.Error("Header.Render should return non-empty string")
	}
}

func TestHeaderString(t *testing.T) {
	h := NewHeader("Title", 80)
	if h.String() != h.Render() {
		t.Error("Header.String should equal Header.Render")
	}
}

// Progress tests
func TestNewProgressBar(t *testing.T) {
	pb := NewProgressBar(40)
	if pb.Width != 40 {
		t.Error("NewProgressBar should set width")
	}
}

func TestProgressBarSetPercent(t *testing.T) {
	pb := NewProgressBar(40)
	pb.SetPercent(0.5)
	if pb.Percent != 0.5 {
		t.Error("SetPercent should set percent")
	}

	// Test bounds
	pb.SetPercent(-0.5)
	if pb.Percent != 0 {
		t.Error("SetPercent should clamp to 0")
	}

	pb.SetPercent(1.5)
	if pb.Percent != 1.0 {
		t.Error("SetPercent should clamp to 1")
	}
}

func TestProgressBarView(t *testing.T) {
	pb := NewProgressBar(40)
	pb.SetPercent(0.5)
	result := pb.View()
	if result == "" {
		t.Error("ProgressBar.View should return non-empty string")
	}
}

func TestNewIndeterminateBar(t *testing.T) {
	ib := NewIndeterminateBar(40)
	if ib.Width != 40 {
		t.Error("NewIndeterminateBar should set width")
	}
}

func TestIndeterminateBarView(t *testing.T) {
	ib := NewIndeterminateBar(40)
	result := ib.View()
	if result == "" {
		t.Error("IndeterminateBar.View should return non-empty string")
	}
}

// GroupedList tests
func TestNewGroupedList(t *testing.T) {
	groups := []ListGroup{
		{
			Name: "Group 1",
			Items: []ListItem{
				{Title: "Item 1"},
				{Title: "Item 2"},
			},
		},
	}
	gl := NewGroupedList(groups)
	if gl == nil {
		t.Error("NewGroupedList should return non-nil")
	}
}

func TestGroupedListNavigation(t *testing.T) {
	groups := []ListGroup{
		{
			Name: "Group 1",
			Items: []ListItem{
				{Title: "Item 1"},
				{Title: "Item 2"},
			},
		},
	}
	gl := NewGroupedList(groups)

	gl.MoveDown()
	gl.MoveUp()
	gl.SelectByNumber(1)
	// Just verify no panic
}

func TestGroupedListSelectedItem(t *testing.T) {
	groups := []ListGroup{
		{
			Name: "Group 1",
			Items: []ListItem{
				{Title: "Item 1"},
				{Title: "Item 2"},
			},
		},
	}
	gl := NewGroupedList(groups)

	item := gl.SelectedItem()
	if item == nil {
		t.Error("SelectedItem should return an item")
	}

	total := gl.TotalItems()
	if total != 2 {
		t.Errorf("TotalItems should return 2, got %d", total)
	}
}

func TestGroupedListRender(t *testing.T) {
	groups := []ListGroup{
		{
			Name: "Group 1",
			Items: []ListItem{
				{Title: "Item 1"},
			},
		},
	}
	gl := NewGroupedList(groups)
	result := gl.Render()
	if result == "" {
		t.Error("GroupedList.Render should return non-empty string")
	}
}

func TestGroupedListString(t *testing.T) {
	groups := []ListGroup{
		{
			Name: "Group 1",
			Items: []ListItem{
				{Title: "Item 1"},
			},
		},
	}
	gl := NewGroupedList(groups)
	if gl.String() != gl.Render() {
		t.Error("GroupedList.String should equal GroupedList.Render")
	}
}

// Gradient function tests
func TestGradientFunctions(t *testing.T) {
	t.Run("gradientPrimary", func(t *testing.T) {
		colors := gradientPrimary()
		if len(colors) == 0 {
			t.Error("gradientPrimary should return colors")
		}
	})

	t.Run("gradientSecondary", func(t *testing.T) {
		colors := gradientSecondary()
		if len(colors) == 0 {
			t.Error("gradientSecondary should return colors")
		}
	})

	t.Run("gradientSuccess", func(t *testing.T) {
		colors := gradientSuccess()
		if len(colors) == 0 {
			t.Error("gradientSuccess should return colors")
		}
	})

	t.Run("gradientRainbow", func(t *testing.T) {
		colors := gradientRainbow()
		if len(colors) < 7 {
			t.Error("gradientRainbow should return rainbow colors")
		}
	})

	t.Run("gradientAgent", func(t *testing.T) {
		agents := []string{"claude", "codex", "gemini", "unknown"}
		for _, agent := range agents {
			colors := gradientAgent(agent)
			if len(colors) == 0 {
				t.Errorf("gradientAgent(%q) should return colors", agent)
			}
		}
	})
}
