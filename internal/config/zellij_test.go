package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDefaultZellijConfig(t *testing.T) {
	cfg := DefaultZellijConfig()

	if cfg.DefaultPanes != 10 {
		t.Errorf("DefaultPanes = %d, want 10", cfg.DefaultPanes)
	}
	if cfg.PluginPath != "" {
		t.Errorf("PluginPath = %q, want empty", cfg.PluginPath)
	}
	if cfg.PaletteKey != "F6" {
		t.Errorf("PaletteKey = %q, want F6", cfg.PaletteKey)
	}
	if !cfg.AttachOnCreate {
		t.Error("AttachOnCreate = false, want true")
	}
}

func TestNZMDefaultPath(t *testing.T) {
	// Save and restore env vars
	origConfig := os.Getenv("NZM_CONFIG")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("NZM_CONFIG", origConfig)
		os.Setenv("XDG_CONFIG_HOME", origXDG)
	}()

	t.Run("uses NZM_CONFIG env var", func(t *testing.T) {
		os.Setenv("NZM_CONFIG", "/custom/path/config.toml")
		os.Setenv("XDG_CONFIG_HOME", "")
		path := NZMDefaultPath()
		if path != "/custom/path/config.toml" {
			t.Errorf("path = %q, want /custom/path/config.toml", path)
		}
	})

	t.Run("expands tilde in NZM_CONFIG", func(t *testing.T) {
		os.Setenv("NZM_CONFIG", "~/myconfig.toml")
		os.Setenv("XDG_CONFIG_HOME", "")
		path := NZMDefaultPath()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, "myconfig.toml")
		if path != want {
			t.Errorf("path = %q, want %q", path, want)
		}
	})

	t.Run("uses XDG_CONFIG_HOME", func(t *testing.T) {
		os.Setenv("NZM_CONFIG", "")
		os.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		path := NZMDefaultPath()
		want := "/xdg/config/nzm/config.toml"
		if path != want {
			t.Errorf("path = %q, want %q", path, want)
		}
	})

	t.Run("falls back to ~/.config/nzm", func(t *testing.T) {
		os.Setenv("NZM_CONFIG", "")
		os.Setenv("XDG_CONFIG_HOME", "")
		path := NZMDefaultPath()
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "nzm", "config.toml")
		if path != want {
			t.Errorf("path = %q, want %q", path, want)
		}
	})
}

func TestNZMDefaultProjectsBase(t *testing.T) {
	// Save and restore env var
	origEnv := os.Getenv("NZM_PROJECTS_BASE")
	defer os.Setenv("NZM_PROJECTS_BASE", origEnv)

	t.Run("uses NZM_PROJECTS_BASE env var", func(t *testing.T) {
		os.Setenv("NZM_PROJECTS_BASE", "/custom/projects")
		base := NZMDefaultProjectsBase()
		if base != "/custom/projects" {
			t.Errorf("base = %q, want /custom/projects", base)
		}
	})

	t.Run("falls back to platform default", func(t *testing.T) {
		os.Setenv("NZM_PROJECTS_BASE", "")
		base := NZMDefaultProjectsBase()
		home, _ := os.UserHomeDir()

		var want string
		if runtime.GOOS == "darwin" {
			want = filepath.Join(home, "Developer")
		} else {
			want = "/data/projects"
		}
		if base != want {
			t.Errorf("base = %q, want %q", base, want)
		}
	})
}

func TestNZMDefault(t *testing.T) {
	cfg := NZMDefault()

	if cfg == nil {
		t.Fatal("NZMDefault() returned nil")
	}

	// Check that key fields are populated
	if cfg.ProjectsBase == "" {
		t.Error("ProjectsBase is empty")
	}
	if cfg.Agents.Claude == "" {
		t.Error("Agents.Claude is empty")
	}
	if cfg.Zellij.DefaultPanes != 10 {
		t.Errorf("Zellij.DefaultPanes = %d, want 10", cfg.Zellij.DefaultPanes)
	}
	if cfg.AgentMail.ProgramName != "nzm" {
		t.Errorf("AgentMail.ProgramName = %q, want nzm", cfg.AgentMail.ProgramName)
	}
	if len(cfg.Palette) == 0 {
		t.Error("Palette is empty")
	}
}

func TestNZMConfig_GetProjectDir(t *testing.T) {
	cfg := &NZMConfig{
		ProjectsBase: "/projects",
	}

	dir := cfg.GetProjectDir("myproject")
	want := "/projects/myproject"
	if dir != want {
		t.Errorf("GetProjectDir = %q, want %q", dir, want)
	}
}

func TestNZMConfig_GetProjectDir_ExpandsTilde(t *testing.T) {
	cfg := &NZMConfig{
		ProjectsBase: "~/projects",
	}

	dir := cfg.GetProjectDir("myproject")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, "projects", "myproject")
	if dir != want {
		t.Errorf("GetProjectDir = %q, want %q", dir, want)
	}
}

func TestNZMLoad(t *testing.T) {
	t.Run("returns defaults for nonexistent file", func(t *testing.T) {
		cfg, err := NZMLoad("/nonexistent/path/config.toml")
		if err != nil {
			t.Fatalf("NZMLoad returned error: %v", err)
		}
		if cfg == nil {
			t.Fatal("NZMLoad returned nil config")
		}
		// Should have defaults
		if cfg.Zellij.DefaultPanes != 10 {
			t.Errorf("Zellij.DefaultPanes = %d, want 10", cfg.Zellij.DefaultPanes)
		}
	})

	t.Run("loads config from file", func(t *testing.T) {
		// Create temp config file
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		content := `
projects_base = "/custom/projects"

[zellij]
default_panes = 5
plugin_path = "/custom/plugin.wasm"
`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write temp config: %v", err)
		}

		cfg, err := NZMLoad(path)
		if err != nil {
			t.Fatalf("NZMLoad returned error: %v", err)
		}
		if cfg.ProjectsBase != "/custom/projects" {
			t.Errorf("ProjectsBase = %q, want /custom/projects", cfg.ProjectsBase)
		}
		if cfg.Zellij.DefaultPanes != 5 {
			t.Errorf("Zellij.DefaultPanes = %d, want 5", cfg.Zellij.DefaultPanes)
		}
		if cfg.Zellij.PluginPath != "/custom/plugin.wasm" {
			t.Errorf("Zellij.PluginPath = %q, want /custom/plugin.wasm", cfg.Zellij.PluginPath)
		}
	})

	t.Run("env vars override file config", func(t *testing.T) {
		// Create temp config file
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		content := `
projects_base = "/file/projects"

[zellij]
plugin_path = "/file/plugin.wasm"
`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write temp config: %v", err)
		}

		// Set env vars
		origBase := os.Getenv("NZM_PROJECTS_BASE")
		origPlugin := os.Getenv("NZM_PLUGIN_PATH")
		defer func() {
			os.Setenv("NZM_PROJECTS_BASE", origBase)
			os.Setenv("NZM_PLUGIN_PATH", origPlugin)
		}()
		os.Setenv("NZM_PROJECTS_BASE", "/env/projects")
		os.Setenv("NZM_PLUGIN_PATH", "/env/plugin.wasm")

		cfg, err := NZMLoad(path)
		if err != nil {
			t.Fatalf("NZMLoad returned error: %v", err)
		}
		// Env should override file
		if cfg.ProjectsBase != "/env/projects" {
			t.Errorf("ProjectsBase = %q, want /env/projects", cfg.ProjectsBase)
		}
		if cfg.Zellij.PluginPath != "/env/plugin.wasm" {
			t.Errorf("Zellij.PluginPath = %q, want /env/plugin.wasm", cfg.Zellij.PluginPath)
		}
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		content := `invalid toml [[[`
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write temp config: %v", err)
		}

		_, err := NZMLoad(path)
		if err == nil {
			t.Error("NZMLoad should return error for invalid TOML")
		}
	})
}
