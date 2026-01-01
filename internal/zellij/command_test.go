package zellij

import (
	"testing"
)

func TestSanitizePaneCommand_Valid(t *testing.T) {
	tests := []string{
		"echo hello",
		"cd /path && ./run.sh",
		"python script.py --arg=value",
		"ls -la | grep test",
	}

	for _, cmd := range tests {
		result, err := SanitizePaneCommand(cmd)
		if err != nil {
			t.Errorf("SanitizePaneCommand(%q) returned error: %v", cmd, err)
		}
		if result != cmd {
			t.Errorf("SanitizePaneCommand(%q) = %q, want same", cmd, result)
		}
	}
}

func TestSanitizePaneCommand_Invalid(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"newline", "echo\nhello"},
		{"carriage_return", "echo\rhello"},
		{"null_byte", "echo\x00hello"},
		{"escape", "echo\x1bhello"},
		{"bell", "echo\x07hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SanitizePaneCommand(tt.cmd)
			if err == nil {
				t.Errorf("SanitizePaneCommand(%q) should return error", tt.cmd)
			}
		})
	}
}

func TestBuildPaneCommand(t *testing.T) {
	cmd, err := BuildPaneCommand("/path/to/project", "python main.py")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "cd '/path/to/project' && python main.py"
	if cmd != expected {
		t.Errorf("BuildPaneCommand = %q, want %q", cmd, expected)
	}
}

func TestBuildPaneCommand_WithSpaces(t *testing.T) {
	cmd, err := BuildPaneCommand("/path/with spaces/project", "python main.py")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "cd '/path/with spaces/project' && python main.py"
	if cmd != expected {
		t.Errorf("BuildPaneCommand = %q, want %q", cmd, expected)
	}
}

func TestBuildPaneCommand_InvalidCommand(t *testing.T) {
	_, err := BuildPaneCommand("/path", "echo\nhello")
	if err == nil {
		t.Error("expected error for invalid command")
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "''"},
		{"hello", "'hello'"},
		{"hello world", "'hello world'"},
		{"it's", `'it'\''s'`},
		{"/path/to/file", "'/path/to/file'"},
	}

	for _, tt := range tests {
		got := ShellQuote(tt.input)
		if got != tt.want {
			t.Errorf("ShellQuote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatPaneName(t *testing.T) {
	tests := []struct {
		session   string
		agentType string
		index     int
		variant   string
		want      string
	}{
		{"project", "cc", 1, "", "project__cc_1"},
		{"project", "cc", 1, "opus", "project__cc_1_opus"},
		{"myapp", "cod", 2, "", "myapp__cod_2"},
		{"test", "gmi", 3, "pro", "test__gmi_3_pro"},
	}

	for _, tt := range tests {
		got := FormatPaneName(tt.session, tt.agentType, tt.index, tt.variant)
		if got != tt.want {
			t.Errorf("FormatPaneName(%q, %q, %d, %q) = %q, want %q",
				tt.session, tt.agentType, tt.index, tt.variant, got, tt.want)
		}
	}
}
