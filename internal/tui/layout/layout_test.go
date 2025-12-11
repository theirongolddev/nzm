package layout

import "testing"

func TestTierForWidth(t *testing.T) {
	tests := []struct {
		width int
		want  Tier
	}{
		{0, TierNarrow},
		{119, TierNarrow},
		{120, TierSplit},
		{121, TierSplit},
		{199, TierSplit},
		{200, TierWide},
		{239, TierWide},
		{240, TierUltra},
		{319, TierUltra},
		{320, TierMega},
		{400, TierMega},
	}

	for _, tt := range tests {
		if got := TierForWidth(tt.width); got != tt.want {
			t.Errorf("TierForWidth(%d) = %v, want %v", tt.width, got, tt.want)
		}
	}
}

func TestSplitProportions(t *testing.T) {
	// Below threshold: should return total,0
	l, r := SplitProportions(100)
	if l != 100 || r != 0 {
		t.Fatalf("SplitProportions(100) = %d,%d want 100,0", l, r)
	}

	// At threshold: ensure budget applied and non-zero panes
	l, r = SplitProportions(140) // avail = 132 -> ~52/80
	if l <= 0 || r <= 0 {
		t.Fatalf("SplitProportions(140) returned zero widths: %d,%d", l, r)
	}
	if l+r > 140 {
		t.Fatalf("SplitProportions(140) sum %d exceeds total 140", l+r)
	}
}

func TestUltraProportions(t *testing.T) {
	// Below threshold falls back to center-only
	l, c, r := UltraProportions(239)
	if l != 0 || r != 0 || c != 239 {
		t.Fatalf("UltraProportions(239) = %d,%d,%d want 0,239,0", l, c, r)
	}

	width := 300 // Ultra tier
	l, c, r = UltraProportions(width)

	total := l + c + r
	expectedTotal := width - 6 // padding budget

	if total != expectedTotal {
		t.Errorf("UltraProportions(%d) total width = %d, want %d", width, total, expectedTotal)
	}

	if l == 0 || c == 0 || r == 0 {
		t.Errorf("UltraProportions(%d) returned zero width panel: %d/%d/%d", width, l, c, r)
	}
}

func TestMegaProportions(t *testing.T) {
	// Below threshold should return center-only
	p1, p2, p3, p4, p5 := MegaProportions(300)
	if p1 != 0 || p3 != 0 || p4 != 0 || p5 != 0 || p2 != 300 {
		t.Fatalf("MegaProportions(300) unexpected: %d,%d,%d,%d,%d", p1, p2, p3, p4, p5)
	}

	width := 400 // Mega tier
	p1, p2, p3, p4, p5 = MegaProportions(width)

	total := p1 + p2 + p3 + p4 + p5
	expectedTotal := width - 10 // padding budget

	if total != expectedTotal {
		t.Errorf("MegaProportions(%d) total width = %d, want %d", width, total, expectedTotal)
	}

	if p1 == 0 || p2 == 0 || p3 == 0 || p4 == 0 || p5 == 0 {
		t.Errorf("MegaProportions(%d) returned zero width panel", width)
	}
}
