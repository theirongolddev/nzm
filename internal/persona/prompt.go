package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PrepareSystemPrompt writes a persona's system prompt to a file and returns the path.
// If the persona has no system prompt, returns empty string and nil error.
// The prompt file is written to {projectDir}/.ntm/prompts/{personaName}.md
func PrepareSystemPrompt(p *Persona, projectDir string) (string, error) {
	if p == nil || p.SystemPrompt == "" {
		return "", nil
	}

	// Create prompts directory
	promptsDir := filepath.Join(projectDir, ".ntm", "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return "", fmt.Errorf("creating prompts directory: %w", err)
	}

	// Build the prompt content
	content := p.SystemPrompt

	// If persona has context_files, prepend them
	if len(p.ContextFiles) > 0 {
		contextContent, err := PrepareContextFiles(p, projectDir)
		if err != nil {
			// Log warning but continue without context files
			fmt.Fprintf(os.Stderr, "warning: could not load context files for persona %s: %v\n", p.Name, err)
		} else if contextContent != "" {
			content = contextContent + "\n\n---\n\n" + content
		}
	}

	// Expand any template variables in the prompt
	content = expandPromptVars(content, p)

	// Write to file
	promptFile := filepath.Join(promptsDir, p.Name+".md")
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing prompt file: %w", err)
	}

	return promptFile, nil
}

// PrepareContextFiles reads and concatenates all context_files for a persona.
// Returns the concatenated content as a string.
func PrepareContextFiles(p *Persona, projectDir string) (string, error) {
	if p == nil || len(p.ContextFiles) == 0 {
		return "", nil
	}

	var files []string

	// Expand globs and collect file paths
	for _, pattern := range p.ContextFiles {
		// Handle relative patterns
		fullPattern := pattern
		if !filepath.IsAbs(pattern) {
			fullPattern = filepath.Join(projectDir, pattern)
		}

		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return "", fmt.Errorf("expanding glob %q: %w", pattern, err)
		}
		files = append(files, matches...)
	}

	if len(files) == 0 {
		return "", nil
	}

	// Read and concatenate files
	var content strings.Builder
	content.WriteString("# Context Files\n\n")

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			// Skip unreadable files with warning
			fmt.Fprintf(os.Stderr, "warning: could not read context file %s: %v\n", f, err)
			continue
		}

		// Get relative path for display
		relPath := f
		if rel, err := filepath.Rel(projectDir, f); err == nil {
			relPath = rel
		}

		content.WriteString(fmt.Sprintf("## %s\n\n", relPath))
		content.WriteString(string(data))
		content.WriteString("\n\n")
	}

	return content.String(), nil
}

// expandPromptVars replaces template variables in the prompt content.
// Supported variables:
//   - {{.Name}} - persona name
//   - {{.Description}} - persona description
//   - {{.AgentType}} - agent type (claude, codex, gemini)
//   - {{.Model}} - model name
func expandPromptVars(content string, p *Persona) string {
	if p == nil {
		return content
	}

	// Simple string replacements (faster than text/template for simple cases)
	replacements := map[string]string{
		"{{.Name}}":        p.Name,
		"{{.Description}}": p.Description,
		"{{.AgentType}}":   p.AgentType,
		"{{.Model}}":       p.Model,
	}

	for old, new := range replacements {
		content = strings.ReplaceAll(content, old, new)
	}

	return content
}

// CleanupPromptFiles removes prompt files for a session.
// This should be called when a session is killed.
func CleanupPromptFiles(projectDir string) error {
	promptsDir := filepath.Join(projectDir, ".ntm", "prompts")

	// Check if directory exists
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		return nil // Nothing to clean up
	}

	// Remove the entire prompts directory
	return os.RemoveAll(promptsDir)
}
