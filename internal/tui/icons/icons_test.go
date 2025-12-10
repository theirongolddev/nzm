package icons

import (
	"os"
	"testing"
)

func TestDetectDefaults(t *testing.T) {
	// Clear env vars
	os.Unsetenv("NTM_ICONS")
	os.Unsetenv("NTM_USE_ICONS")
	os.Unsetenv("NERD_FONTS")
	
	// Should default to ASCII
	icons := Detect()
	if icons.Session != "[]" { // ASCII session
		t.Errorf("Expected ASCII default, got session=%q", icons.Session)
	}
}

func TestDetectExplicit(t *testing.T) {
	os.Setenv("NTM_ICONS", "unicode")
	defer os.Unsetenv("NTM_ICONS")
	
	icons := Detect()
	if icons.Session != "◆" { // Unicode session
		t.Errorf("Expected Unicode, got session=%q", icons.Session)
	}
	
	os.Setenv("NTM_ICONS", "ascii")
	icons = Detect()
	if icons.Session != "[]" {
		t.Errorf("Expected ASCII, got session=%q", icons.Session)
	}
}

func TestDetectAuto(t *testing.T) {
	os.Setenv("NTM_ICONS", "auto")
	defer os.Unsetenv("NTM_ICONS")
	
	// This depends on environment, but should return something valid
	icons := Detect()
	if icons.Session == "" {
		t.Error("Returned empty icons")
	}
}