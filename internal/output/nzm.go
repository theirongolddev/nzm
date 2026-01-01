package output

import "os"

// NZMDetectFormat determines the output format for NZM commands.
// Priority: explicit flag > NZM_OUTPUT_FORMAT > NTM_OUTPUT_FORMAT > pipe detection > default text
func NZMDetectFormat(jsonFlag bool) Format {
	// 1. Explicit --json flag takes highest priority
	if jsonFlag {
		return FormatJSON
	}

	// 2. Check NZM_OUTPUT_FORMAT environment variable first
	if envFormat := os.Getenv("NZM_OUTPUT_FORMAT"); envFormat != "" {
		switch envFormat {
		case "json", "JSON":
			return FormatJSON
		case "text", "TEXT":
			return FormatText
		}
	}

	// 3. Fall back to NTM_OUTPUT_FORMAT for compatibility
	if envFormat := os.Getenv("NTM_OUTPUT_FORMAT"); envFormat != "" {
		switch envFormat {
		case "json", "JSON":
			return FormatJSON
		case "text", "TEXT":
			return FormatText
		}
	}

	// 4. Auto-detect: if stdout is not a terminal, use JSON
	if !IsTerminal() {
		return FormatJSON
	}

	// 5. Default to text for interactive terminals
	return FormatText
}

// NZMDefaultFormatter returns a formatter for NZM commands based on the JSON flag
func NZMDefaultFormatter(jsonFlag bool) *Formatter {
	return New(WithFormat(NZMDetectFormat(jsonFlag)))
}
