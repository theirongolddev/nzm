package output

import (
	"os"
	"testing"
)

func TestNZMDetectFormat(t *testing.T) {
	// Save original env vars
	origNZM := os.Getenv("NZM_OUTPUT_FORMAT")
	origNTM := os.Getenv("NTM_OUTPUT_FORMAT")
	defer func() {
		os.Setenv("NZM_OUTPUT_FORMAT", origNZM)
		os.Setenv("NTM_OUTPUT_FORMAT", origNTM)
	}()

	t.Run("explicit json flag takes priority", func(t *testing.T) {
		os.Setenv("NZM_OUTPUT_FORMAT", "text")
		os.Setenv("NTM_OUTPUT_FORMAT", "text")

		format := NZMDetectFormat(true)
		if format != FormatJSON {
			t.Errorf("format = %v, want FormatJSON", format)
		}
	})

	t.Run("NZM_OUTPUT_FORMAT json", func(t *testing.T) {
		os.Setenv("NZM_OUTPUT_FORMAT", "json")
		os.Setenv("NTM_OUTPUT_FORMAT", "text")

		format := NZMDetectFormat(false)
		if format != FormatJSON {
			t.Errorf("format = %v, want FormatJSON", format)
		}
	})

	t.Run("NZM_OUTPUT_FORMAT text", func(t *testing.T) {
		os.Setenv("NZM_OUTPUT_FORMAT", "text")
		os.Setenv("NTM_OUTPUT_FORMAT", "json")

		format := NZMDetectFormat(false)
		if format != FormatText {
			t.Errorf("format = %v, want FormatText", format)
		}
	})

	t.Run("falls back to NTM_OUTPUT_FORMAT", func(t *testing.T) {
		os.Setenv("NZM_OUTPUT_FORMAT", "")
		os.Setenv("NTM_OUTPUT_FORMAT", "json")

		format := NZMDetectFormat(false)
		if format != FormatJSON {
			t.Errorf("format = %v, want FormatJSON", format)
		}
	})

	t.Run("NZM takes precedence over NTM", func(t *testing.T) {
		os.Setenv("NZM_OUTPUT_FORMAT", "text")
		os.Setenv("NTM_OUTPUT_FORMAT", "json")

		format := NZMDetectFormat(false)
		if format != FormatText {
			t.Errorf("format = %v, want FormatText (NZM should take precedence)", format)
		}
	})
}

func TestNZMDefaultFormatter(t *testing.T) {
	t.Run("returns formatter with json format when flag is true", func(t *testing.T) {
		f := NZMDefaultFormatter(true)
		if !f.IsJSON() {
			t.Error("expected JSON format when jsonFlag is true")
		}
	})

	t.Run("returns formatter based on env when flag is false", func(t *testing.T) {
		origNZM := os.Getenv("NZM_OUTPUT_FORMAT")
		origNTM := os.Getenv("NTM_OUTPUT_FORMAT")
		defer func() {
			os.Setenv("NZM_OUTPUT_FORMAT", origNZM)
			os.Setenv("NTM_OUTPUT_FORMAT", origNTM)
		}()

		os.Setenv("NZM_OUTPUT_FORMAT", "json")
		os.Setenv("NTM_OUTPUT_FORMAT", "")

		f := NZMDefaultFormatter(false)
		if !f.IsJSON() {
			t.Error("expected JSON format when NZM_OUTPUT_FORMAT=json")
		}
	})
}
