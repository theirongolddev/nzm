package persona

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersonaValidation(t *testing.T) {
	tests := []struct {
		name    string
		persona Persona
		wantErr bool
	}{
		{
			name: "valid persona",
			persona: Persona{
				Name:        "test",
				AgentType:   "claude",
				Model:       "sonnet",
				Description: "Test persona",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			persona: Persona{
				AgentType: "claude",
			},
			wantErr: true,
		},
		{
			name: "missing agent_type",
			persona: Persona{
				Name: "test",
			},
			wantErr: true,
		},
		{
			name: "invalid agent_type",
			persona: Persona{
				Name:      "test",
				AgentType: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid claude short name",
			persona: Persona{
				Name:      "test",
				AgentType: "cc",
			},
			wantErr: false,
		},
		{
			name: "valid codex",
			persona: Persona{
				Name:      "test",
				AgentType: "codex",
			},
			wantErr: false,
		},
		{
			name: "valid gemini short name",
			persona: Persona{
				Name:      "test",
				AgentType: "gmi",
			},
			wantErr: false,
		},
		{
			name: "invalid temperature - too high",
			persona: Persona{
				Name:        "test",
				AgentType:   "claude",
				Temperature: ptrFloat64(2.5),
			},
			wantErr: true,
		},
		{
			name: "invalid temperature - negative",
			persona: Persona{
				Name:        "test",
				AgentType:   "claude",
				Temperature: ptrFloat64(-0.1),
			},
			wantErr: true,
		},
		{
			name: "valid temperature",
			persona: Persona{
				Name:        "test",
				AgentType:   "claude",
				Temperature: ptrFloat64(0.7),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.persona.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentTypeFlag(t *testing.T) {
	tests := []struct {
		agentType string
		want      string
	}{
		{"claude", "cc"},
		{"Claude", "cc"},
		{"cc", "cc"},
		{"codex", "cod"},
		{"Codex", "cod"},
		{"cod", "cod"},
		{"gemini", "gmi"},
		{"Gemini", "gmi"},
		{"gmi", "gmi"},
		{"unknown", "cc"}, // defaults to cc
	}

	for _, tt := range tests {
		t.Run(tt.agentType, func(t *testing.T) {
			p := &Persona{AgentType: tt.agentType}
			if got := p.AgentTypeFlag(); got != tt.want {
				t.Errorf("AgentTypeFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	p1 := &Persona{Name: "architect", AgentType: "claude"}
	p2 := &Persona{Name: "implementer", AgentType: "claude"}
	p3 := &Persona{Name: "Architect", AgentType: "codex"} // Override with different case

	r.Add(p1)
	r.Add(p2)
	r.Add(p3) // Should override p1

	// Test Get with case insensitivity
	got, ok := r.Get("architect")
	if !ok {
		t.Error("expected to find architect")
	}
	if got.AgentType != "codex" {
		t.Errorf("expected architect to be codex (overwritten), got %s", got.AgentType)
	}

	// Test List
	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 personas, got %d", len(list))
	}

	// Test Get not found
	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent persona")
	}
}

func TestBuiltinPersonas(t *testing.T) {
	personas := BuiltinPersonas()

	if len(personas) < 5 {
		t.Errorf("expected at least 5 builtin personas, got %d", len(personas))
	}

	// Verify all builtin personas are valid
	for _, p := range personas {
		if err := p.Validate(); err != nil {
			t.Errorf("builtin persona %q is invalid: %v", p.Name, err)
		}
	}

	// Check expected personas exist
	names := make(map[string]bool)
	for _, p := range personas {
		names[p.Name] = true
	}

	expected := []string{"architect", "implementer", "reviewer", "tester", "documenter"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected builtin persona %q not found", name)
		}
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	personasFile := filepath.Join(tmpDir, "personas.toml")

	content := `
[[personas]]
name = "custom"
description = "Custom test persona"
agent_type = "claude"
model = "opus"
system_prompt = "You are a custom agent."
temperature = 0.5
tags = ["custom", "test"]

[[personas]]
name = "another"
description = "Another persona"
agent_type = "codex"
model = "gpt-4"
`

	if err := os.WriteFile(personasFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := LoadFromFile(personasFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if len(cfg.Personas) != 2 {
		t.Errorf("expected 2 personas, got %d", len(cfg.Personas))
	}

	// Check first persona
	p := cfg.Personas[0]
	if p.Name != "custom" {
		t.Errorf("expected name 'custom', got %q", p.Name)
	}
	if p.Model != "opus" {
		t.Errorf("expected model 'opus', got %q", p.Model)
	}
	if p.Temperature == nil || *p.Temperature != 0.5 {
		t.Error("expected temperature 0.5")
	}
	if len(p.Tags) != 2 || p.Tags[0] != "custom" {
		t.Errorf("unexpected tags: %v", p.Tags)
	}
}

func TestLoadRegistry(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(filepath.Join(projectDir, ".ntm"), 0755); err != nil {
		t.Fatalf("failed to create project dir: %v", err)
	}

	// Create project personas file
	projectPersonas := `
[[personas]]
name = "architect"
description = "Project-specific architect"
agent_type = "codex"
model = "gpt-4"
`
	projectPath := filepath.Join(projectDir, ".ntm", "personas.toml")
	if err := os.WriteFile(projectPath, []byte(projectPersonas), 0644); err != nil {
		t.Fatalf("failed to write project personas: %v", err)
	}

	registry, err := LoadRegistry(projectDir)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}

	// Should have builtins + project override
	architect, ok := registry.Get("architect")
	if !ok {
		t.Error("expected to find architect")
	}
	// Project should override builtin
	if architect.AgentType != "codex" {
		t.Errorf("expected project architect with codex, got %s", architect.AgentType)
	}

	// Builtin-only personas should still exist
	if _, ok := registry.Get("implementer"); !ok {
		t.Error("expected to find builtin implementer")
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}
