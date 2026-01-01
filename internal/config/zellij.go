package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// ZellijConfig holds Zellij-specific settings (parallel to TmuxConfig for NZM)
type ZellijConfig struct {
	DefaultPanes   int    `toml:"default_panes"`    // Default number of panes to spawn
	PluginPath     string `toml:"plugin_path"`      // Path to nzm-agent.wasm plugin
	PaletteKey     string `toml:"palette_key"`      // Keybinding for command palette
	AttachOnCreate bool   `toml:"attach_on_create"` // Auto-attach when creating session
}

// DefaultZellijConfig returns sensible Zellij defaults
func DefaultZellijConfig() ZellijConfig {
	return ZellijConfig{
		DefaultPanes:   10,
		PluginPath:     "", // Auto-detect from installation
		PaletteKey:     "F6",
		AttachOnCreate: true,
	}
}

// configPath returns the config file path for the given tool name and env var
func configPath(toolName, envVar string) string {
	if env := os.Getenv(envVar); env != "" {
		return ExpandHome(env)
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, toolName, "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", toolName, "config.toml")
}

// projectsBase returns the projects base directory, checking the given env var first
func projectsBase(envVar string) string {
	if envBase := os.Getenv(envVar); envBase != "" {
		return envBase
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Developer")
	}
	return "/data/projects"
}

// NZMDefaultPath returns the default config file path for NZM
func NZMDefaultPath() string {
	return configPath("nzm", "NZM_CONFIG")
}

// NZMDefaultProjectsBase returns the default projects directory for NZM
func NZMDefaultProjectsBase() string {
	return projectsBase("NZM_PROJECTS_BASE")
}

// NZMConfig represents the main configuration for NZM (Zellij version)
type NZMConfig struct {
	ProjectsBase  string            `toml:"projects_base"`
	Theme         string            `toml:"theme"`        // UI Theme (mocha, macchiato, nord, latte, auto)
	PaletteFile   string            `toml:"palette_file"` // Path to command_palette.md (optional)
	Agents        AgentConfig       `toml:"agents"`
	Palette       []PaletteCmd      `toml:"palette"`
	PaletteState  PaletteState      `toml:"palette_state"`
	Zellij        ZellijConfig      `toml:"zellij"`
	AgentMail     AgentMailConfig   `toml:"agent_mail"`
	Models        ModelsConfig      `toml:"models"`
	Alerts        AlertsConfig      `toml:"alerts"`
	Checkpoints   CheckpointsConfig `toml:"checkpoints"`
	Resilience    ResilienceConfig  `toml:"resilience"`
	Scanner       ScannerConfig     `toml:"scanner"`
	CASS          CASSConfig        `toml:"cass"`
	Accounts      AccountsConfig    `toml:"accounts"`
	Rotation      RotationConfig    `toml:"rotation"`
	GeminiSetup   GeminiSetupConfig `toml:"gemini_setup"`

	// Runtime-only fields (populated by project config merging)
	ProjectDefaults map[string]int `toml:"-"`
}

// NZMDefault returns the default NZM configuration
func NZMDefault() *NZMConfig {
	// Determine projects base: env var takes precedence
	projectsBase := NZMDefaultProjectsBase()

	return &NZMConfig{
		ProjectsBase: projectsBase,
		Agents:       DefaultAgentTemplates(),
		Zellij:       DefaultZellijConfig(),
		AgentMail: AgentMailConfig{
			Enabled:      true,
			URL:          DefaultAgentMailURL,
			Token:        "",
			AutoRegister: true,
			ProgramName:  "nzm",
		},
		Models:      DefaultModels(),
		Alerts:      DefaultAlertsConfig(),
		Checkpoints: DefaultCheckpointsConfig(),
		Resilience:  DefaultResilienceConfig(),
		Scanner:     DefaultScannerConfig(),
		CASS:        DefaultCASSConfig(),
		Accounts:    DefaultAccountsConfig(),
		Rotation:    DefaultRotationConfig(),
		GeminiSetup: DefaultGeminiSetupConfig(),
		Palette:     defaultPaletteCommands(),
	}
}

// NZMLoad loads NZM configuration from a file.
// If path is empty, uses NZMDefaultPath().
func NZMLoad(path string) (*NZMConfig, error) {
	if path == "" {
		path = NZMDefaultPath()
	}

	// 1. Initialize with defaults
	cfg := NZMDefault()

	// 2. Read and unmarshal TOML over defaults
	if data, err := os.ReadFile(path); err == nil {
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// 3. Apply Environment Variable Overrides (Env > TOML > Default)
	if envBase := os.Getenv("NZM_PROJECTS_BASE"); envBase != "" {
		cfg.ProjectsBase = envBase
	}

	// AgentMail Env Overrides
	if url := os.Getenv("AGENT_MAIL_URL"); url != "" {
		cfg.AgentMail.URL = url
	}
	if token := os.Getenv("AGENT_MAIL_TOKEN"); token != "" {
		cfg.AgentMail.Token = token
	}
	if enabled := os.Getenv("AGENT_MAIL_ENABLED"); enabled != "" {
		cfg.AgentMail.Enabled = enabled == "1" || enabled == "true"
	}

	// Zellij Env Overrides
	if pluginPath := os.Getenv("NZM_PLUGIN_PATH"); pluginPath != "" {
		cfg.Zellij.PluginPath = pluginPath
	}

	return cfg, nil
}

// GetProjectDir returns the project directory for a session
func (c *NZMConfig) GetProjectDir(session string) string {
	base := ExpandHome(c.ProjectsBase)
	return filepath.Join(base, session)
}
