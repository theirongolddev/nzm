package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadMerged loads the global config and merges any project-specific config found starting from cwd.
func LoadMerged(cwd, globalPath string) (*Config, error) {
	// Load global
	cfg, err := Load(globalPath)
	if err != nil {
		// If global missing, use defaults
		cfg = Default()
	}

	// Find project config
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	projectDir, projectCfg, err := FindProjectConfig(cwd)
	if err != nil {
		// Return error if project config is invalid, so user knows
		return cfg, fmt.Errorf("loading project config: %w", err)
	}

	if projectCfg != nil {
		cfg = MergeConfig(cfg, projectCfg, projectDir)
	}

	return cfg, nil
}

// MergeConfig merges project config into global config.
func MergeConfig(global *Config, project *ProjectConfig, projectDir string) *Config {
	// Merge Agents
	if project.Agents.Claude != "" {
		global.Agents.Claude = project.Agents.Claude
	}
	if project.Agents.Codex != "" {
		global.Agents.Codex = project.Agents.Codex
	}
	if project.Agents.Gemini != "" {
		global.Agents.Gemini = project.Agents.Gemini
	}

	// Merge Defaults
	if len(project.Defaults.Agents) > 0 {
		global.ProjectDefaults = project.Defaults.Agents
	}

	// Merge Palette File
	if project.Palette.File != "" {
		palettePath := filepath.Join(projectDir, ".ntm", project.Palette.File)
		if cmds, err := LoadPaletteFromMarkdown(palettePath); err == nil && len(cmds) > 0 {
			// Append project commands to global palette
			// Deduplicate based on key? Or just append?
			// Task says "Arrays are merged".
			// Simple append for now.
			global.Palette = append(global.Palette, cmds...)
		}
	}

	return global
}
