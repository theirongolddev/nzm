package zellij

import (
	"fmt"
	"strings"
)

// SanitizePaneCommand rejects control characters that could inject unintended
// key sequences when sending commands into panes.
func SanitizePaneCommand(cmd string) (string, error) {
	for _, r := range cmd {
		switch {
		case r == '\n', r == '\r', r == 0:
			return "", fmt.Errorf("command contains disallowed control characters")
		case r < 0x20 && r != ' ' && r != '\t':
			return "", fmt.Errorf("command contains disallowed control character 0x%02x", r)
		}
	}
	return cmd, nil
}

// BuildPaneCommand constructs a safe cd+command string for execution inside a
// pane, rejecting commands with unsafe control characters.
func BuildPaneCommand(projectDir, agentCommand string) (string, error) {
	safeCommand, err := SanitizePaneCommand(agentCommand)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cd %s && %s", ShellQuote(projectDir), safeCommand), nil
}

// ShellQuote returns a POSIX-shell-safe single-quoted string.
func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	// Close-quote, escape single quote, reopen: ' -> '\''.
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// FormatPaneName formats a pane title according to NZM convention
func FormatPaneName(session string, agentType string, index int, variant string) string {
	base := fmt.Sprintf("%s__%s_%d", session, agentType, index)
	if variant != "" {
		return fmt.Sprintf("%s_%s", base, variant)
	}
	return base
}
