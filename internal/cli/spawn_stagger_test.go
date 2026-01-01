package cli

import (
	"testing"
	"time"
)

func TestOptionalDurationValue_Set(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantDuration    time.Duration
		wantEnabled     bool
		wantErr         bool
	}{
		{
			name:         "empty string uses default",
			input:        "",
			wantDuration: 90 * time.Second,
			wantEnabled:  true,
		},
		{
			name:         "explicit duration",
			input:        "2m",
			wantDuration: 2 * time.Minute,
			wantEnabled:  true,
		},
		{
			name:         "zero disables",
			input:        "0",
			wantDuration: 0,
			wantEnabled:  false,
		},
		{
			name:         "30 seconds",
			input:        "30s",
			wantDuration: 30 * time.Second,
			wantEnabled:  true,
		},
		{
			name:         "5 minutes",
			input:        "5m",
			wantDuration: 5 * time.Minute,
			wantEnabled:  true,
		},
		{
			name:    "invalid duration",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "negative duration rejected",
			input:   "-1m",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var duration time.Duration
			var enabled bool
			v := newOptionalDurationValue(90*time.Second, &duration, &enabled)

			err := v.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if duration != tt.wantDuration {
				t.Errorf("Set(%q) duration = %v, want %v", tt.input, duration, tt.wantDuration)
			}
			if enabled != tt.wantEnabled {
				t.Errorf("Set(%q) enabled = %v, want %v", tt.input, enabled, tt.wantEnabled)
			}
		})
	}
}

func TestOptionalDurationValue_String(t *testing.T) {
	var duration time.Duration
	var enabled bool

	v := newOptionalDurationValue(90*time.Second, &duration, &enabled)

	// Before Set, should return empty
	if got := v.String(); got != "" {
		t.Errorf("String() before Set = %q, want empty", got)
	}

	// After Set, should return the duration
	_ = v.Set("2m")
	if got := v.String(); got != "2m0s" {
		t.Errorf("String() after Set = %q, want %q", got, "2m0s")
	}
}

func TestOptionalDurationValue_NoOptDefVal(t *testing.T) {
	var duration time.Duration
	var enabled bool

	v := newOptionalDurationValue(90*time.Second, &duration, &enabled)

	if got := v.NoOptDefVal(); got != "1m30s" {
		t.Errorf("NoOptDefVal() = %q, want %q", got, "1m30s")
	}
}

func TestOptionalDurationValue_Type(t *testing.T) {
	var duration time.Duration
	var enabled bool

	v := newOptionalDurationValue(90*time.Second, &duration, &enabled)

	if got := v.Type(); got != "duration" {
		t.Errorf("Type() = %q, want %q", got, "duration")
	}
}

func TestStaggerDelayCalculation(t *testing.T) {
	// Test the stagger delay calculation logic
	stagger := 90 * time.Second

	tests := []struct {
		agentIdx int
		want     time.Duration
	}{
		{0, 0},                    // First agent: no delay
		{1, 90 * time.Second},     // Second: 90s
		{2, 180 * time.Second},    // Third: 180s (3m)
		{3, 270 * time.Second},    // Fourth: 270s (4.5m)
		{4, 360 * time.Second},    // Fifth: 360s (6m)
	}

	for _, tt := range tests {
		got := time.Duration(tt.agentIdx) * stagger
		if got != tt.want {
			t.Errorf("agent %d delay = %v, want %v", tt.agentIdx, got, tt.want)
		}
	}
}
