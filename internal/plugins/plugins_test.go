package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCommandPlugins(t *testing.T) {
	// Create temp directory for plugins
	tmpDir, err := os.MkdirTemp("", "ntm-plugins-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid plugin file
	validPlugin := `
[plugin]
name = "hello"
command = "echo hello"
description = "Says hello"
usage = "ntm hello"
tags = ["test"]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "hello.toml"), []byte(validPlugin), 0644); err != nil {
		t.Fatalf("failed to write valid plugin: %v", err)
	}

	// Create an invalid plugin file (missing required fields)
	invalidPlugin := `
[plugin]
description = "Missing name and command"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.toml"), []byte(invalidPlugin), 0644); err != nil {
		t.Fatalf("failed to write invalid plugin: %v", err)
	}

	// Load plugins
	plugins, err := LoadCommandPlugins(tmpDir)
	if err != nil {
		t.Fatalf("LoadCommandPlugins failed: %v", err)
	}

	// Verify results
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}

	p := plugins[0]
	if p.Name != "hello" {
		t.Errorf("expected plugin name 'hello', got '%s'", p.Name)
	}
	if p.Command != "echo hello" {
		t.Errorf("expected plugin command 'echo hello', got '%s'", p.Command)
	}
	if len(p.Tags) != 1 || p.Tags[0] != "test" {
		t.Errorf("expected tags ['test'], got %v", p.Tags)
	}
}
