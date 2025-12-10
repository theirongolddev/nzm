// Package persona provides configuration and management for AI agent personas.
// Personas define agent characteristics including agent type, model, system prompts,
// and behavioral settings.
package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Persona defines the configuration for an agent persona.
type Persona struct {
	Name         string   `toml:"name"`
	Description  string   `toml:"description"`
	AgentType    string   `toml:"agent_type"`    // claude, codex, gemini
	Model        string   `toml:"model"`         // Model alias or full name
	SystemPrompt string   `toml:"system_prompt"` // System prompt to inject
	Temperature  *float64 `toml:"temperature,omitempty"`
	ContextFiles []string `toml:"context_files,omitempty"` // Globs of files to include in context
	Tags         []string `toml:"tags,omitempty"`
}

// PersonasConfig holds a collection of persona definitions.
type PersonasConfig struct {
	Personas []Persona `toml:"personas"`
}

// AgentTypeFlag returns the NTM flag for this persona's agent type.
// e.g., "claude" -> "cc", "codex" -> "cod", "gemini" -> "gmi"
func (p *Persona) AgentTypeFlag() string {
	switch strings.ToLower(p.AgentType) {
	case "claude", "cc":
		return "cc"
	case "codex", "cod":
		return "cod"
	case "gemini", "gmi":
		return "gmi"
	default:
		return "cc" // Default to Claude
	}
}

// Validate checks if the persona configuration is valid.
func (p *Persona) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("persona name is required")
	}
	if p.AgentType == "" {
		return fmt.Errorf("persona %q: agent_type is required", p.Name)
	}

	// Validate agent type
	switch strings.ToLower(p.AgentType) {
	case "claude", "cc", "codex", "cod", "gemini", "gmi":
		// valid
	default:
		return fmt.Errorf("persona %q: invalid agent_type %q (must be claude, codex, or gemini)", p.Name, p.AgentType)
	}

	// Validate temperature if set
	if p.Temperature != nil {
		if *p.Temperature < 0 || *p.Temperature > 2 {
			return fmt.Errorf("persona %q: temperature must be between 0 and 2", p.Name)
		}
	}

	return nil
}

// Registry holds loaded personas and provides lookup functionality.
type Registry struct {
	personas map[string]*Persona
}

// NewRegistry creates a new empty persona registry.
func NewRegistry() *Registry {
	return &Registry{
		personas: make(map[string]*Persona),
	}
}

// Add adds a persona to the registry, overwriting any existing persona with the same name.
func (r *Registry) Add(p *Persona) {
	r.personas[strings.ToLower(p.Name)] = p
}

// Get retrieves a persona by name (case-insensitive).
func (r *Registry) Get(name string) (*Persona, bool) {
	p, ok := r.personas[strings.ToLower(name)]
	return p, ok
}

// List returns all personas in the registry.
func (r *Registry) List() []*Persona {
	result := make([]*Persona, 0, len(r.personas))
	for _, p := range r.personas {
		result = append(result, p)
	}
	return result
}

// LoadFromFile loads personas from a TOML file.
func LoadFromFile(path string) (*PersonasConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg PersonasConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing personas file %s: %w", path, err)
	}

	// Validate all personas
	for i := range cfg.Personas {
		if err := cfg.Personas[i].Validate(); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

// DefaultUserPath returns the default user personas file path.
func DefaultUserPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ntm", "personas.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ntm", "personas.toml")
}

// DefaultProjectPath returns the default project personas file path.
func DefaultProjectPath() string {
	return ".ntm/personas.toml"
}

// LoadRegistry loads personas from all sources and returns a registry.
// Loading precedence: built-in -> user -> project (later sources override earlier).
func LoadRegistry(projectDir string) (*Registry, error) {
	registry := NewRegistry()

	// 1. Load built-in personas
	for _, p := range BuiltinPersonas() {
		registry.Add(&p)
	}

	// 2. Load user personas
	userPath := DefaultUserPath()
	if cfg, err := LoadFromFile(userPath); err == nil {
		for i := range cfg.Personas {
			registry.Add(&cfg.Personas[i])
		}
	}
	// Ignore file not found errors for user personas

	// 3. Load project personas
	if projectDir != "" {
		projectPath := filepath.Join(projectDir, DefaultProjectPath())
		if cfg, err := LoadFromFile(projectPath); err == nil {
			for i := range cfg.Personas {
				registry.Add(&cfg.Personas[i])
			}
		}
		// Ignore file not found errors for project personas
	}

	return registry, nil
}

// BuiltinPersonas returns the default built-in persona definitions.
func BuiltinPersonas() []Persona {
	return []Persona{
		{
			Name:        "architect",
			Description: "Senior software architect focused on system design and high-level decisions",
			AgentType:   "claude",
			Model:       "opus",
			SystemPrompt: `You are a senior software architect. Your focus is on:
- System design and architecture decisions
- Code organization and module boundaries
- Design patterns and best practices
- Performance and scalability considerations
- Security architecture

When reviewing code, focus on structural issues rather than implementation details.
Propose refactoring strategies and architectural improvements.
Consider long-term maintainability and extensibility.`,
			Tags: []string{"design", "review", "architecture"},
		},
		{
			Name:        "implementer",
			Description: "Fast implementation engineer focused on writing working code quickly",
			AgentType:   "claude",
			Model:       "sonnet",
			SystemPrompt: `You are a fast implementation engineer. Your focus is on:
- Writing working code quickly and efficiently
- Following existing patterns in the codebase
- Implementing features to specification
- Writing clean, readable code
- Adding appropriate error handling

Prioritize getting things working over perfection.
Follow the existing code style and conventions.
Ask clarifying questions if requirements are unclear.`,
			Tags: []string{"implementation", "coding", "fast"},
		},
		{
			Name:        "reviewer",
			Description: "Code reviewer focused on quality, bugs, and best practices",
			AgentType:   "claude",
			Model:       "sonnet",
			SystemPrompt: `You are a thorough code reviewer. Your focus is on:
- Finding bugs and logical errors
- Identifying security vulnerabilities
- Checking for edge cases and error handling
- Ensuring code follows best practices
- Reviewing test coverage

Be constructive in your feedback.
Explain the reasoning behind your suggestions.
Prioritize issues by severity.`,
			Tags: []string{"review", "quality", "bugs"},
		},
		{
			Name:        "tester",
			Description: "QA engineer focused on testing and test coverage",
			AgentType:   "claude",
			Model:       "sonnet",
			SystemPrompt: `You are a QA engineer focused on testing. Your focus is on:
- Writing comprehensive test cases
- Identifying edge cases and boundary conditions
- Creating unit, integration, and e2e tests
- Improving test coverage
- Finding bugs through systematic testing

Think about what could go wrong.
Test both happy paths and error cases.
Consider performance and stress testing where appropriate.`,
			Tags: []string{"testing", "qa", "quality"},
		},
		{
			Name:        "documenter",
			Description: "Technical writer focused on documentation and clarity",
			AgentType:   "claude",
			Model:       "sonnet",
			SystemPrompt: `You are a technical writer focused on documentation. Your focus is on:
- Writing clear, concise documentation
- Adding helpful code comments
- Creating README files and guides
- Documenting APIs and interfaces
- Explaining complex concepts simply

Write for your audience - other developers.
Include examples where helpful.
Keep documentation up to date with code changes.`,
			Tags: []string{"documentation", "writing", "clarity"},
		},
	}
}
