package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Plugin represents a loaded command plugin
type Plugin struct {
	Name        string   `toml:"name"`
	Command     string   `toml:"command"`
	Description string   `toml:"description"`
	Usage       string   `toml:"usage"`
	Hooks       []string `toml:"hooks"`
	Tags        []string `toml:"tags"`
	Path        string   `toml:"-"` // Source file path
}

// Config represents the TOML structure for a plugin file
type Config struct {
	Plugin Plugin `toml:"plugin"`
}

// LoadCommandPlugins scans a directory for TOML plugin definitions
func LoadCommandPlugins(dir string) ([]Plugin, error) {
	var plugins []Plugin

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading plugins directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		plugin, err := loadPluginFile(path)
		if err != nil {
			// Log error but continue loading other plugins
			// TODO: Use a proper logger
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", entry.Name(), err)
			continue
		}

		plugins = append(plugins, *plugin)
	}

	return plugins, nil
}

func loadPluginFile(path string) (*Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing TOML: %w", err)
	}

	if cfg.Plugin.Name == "" {
		return nil, fmt.Errorf("missing plugin name")
	}
	if cfg.Plugin.Command == "" {
		return nil, fmt.Errorf("missing plugin command")
	}

	cfg.Plugin.Path = path
	return &cfg.Plugin, nil
}

// Execute runs the plugin command
// This is a placeholder for the actual execution logic which will be implemented later
// likely using os/exec and passing environment variables
func (p *Plugin) Execute(args []string, env map[string]string) error {
	// TODO: Implement execution logic
	// 1. Construct command (replace placeholders?)
	// 2. Set up environment
	// 3. Run command attached to stdout/stderr
	return fmt.Errorf("plugin execution not yet implemented: %s", p.Name)
}
