package recipe

import (
	"testing"
)

func TestValidateAgentSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    AgentSpec
		wantErr bool
	}{
		{"valid", AgentSpec{Type: "cc", Count: 1}, false},
		{"valid_custom", AgentSpec{Type: "cursor", Count: 1}, false},
		{"missing_type", AgentSpec{Count: 1}, true},
		{"zero_count", AgentSpec{Type: "cc", Count: 0}, true},
		{"high_count", AgentSpec{Type: "cc", Count: 21}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateAgentSpec(tt.spec); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgentSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRecipeHelpers(t *testing.T) {
	r := Recipe{
		Agents: []AgentSpec{
			{Type: "cc", Count: 2},
			{Type: "cod", Count: 1},
		},
	}

	if total := r.TotalAgents(); total != 3 {
		t.Errorf("TotalAgents() = %d, want 3", total)
	}

	counts := r.AgentCounts()
	if counts["cc"] != 2 {
		t.Errorf("AgentCounts[cc] = %d, want 2", counts["cc"])
	}
	if counts["cod"] != 1 {
		t.Errorf("AgentCounts[cod] = %d, want 1", counts["cod"])
	}
}
