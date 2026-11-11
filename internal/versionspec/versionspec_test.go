package versionspec

import "testing"

func TestSpec(t *testing.T) {
	tests := []struct {
		name     string
		declared string
		version  string
		report   bool
		want     string
	}{
		{"reported for latest when asked", "^1.0.0", Latest, true, "^1.0.0"},
		{"trimmed", "  ^1.0.0 ", Latest, true, "^1.0.0"},
		{"not reported by default", "^1.0.0", Latest, false, ""},
		{"never next to an exact version", "1.0.0", "1.0.0", true, ""},
		{"nothing declared", "", Latest, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Spec(tt.declared, tt.version, tt.report); got != tt.want {
				t.Errorf("Spec(%q, %q, %v) = %q, want %q", tt.declared, tt.version, tt.report, got, tt.want)
			}
		})
	}
}

func TestRequirementSpec(t *testing.T) {
	tests := []struct {
		requirement string
		want        string
	}{
		{"pkg==1.0", "==1.0"},
		{"pkg[extra]>=1.0,<2", ">=1.0,<2"},
		{"pkg (>=1.0)", ">=1.0"},
		{"pkg>=1.0; python_version < '3.8'", ">=1.0"},
		{"pkg", ""},
		{"pkg @ https://example.com/pkg.whl", ""},
	}

	for _, tt := range tests {
		if got := RequirementSpec(tt.requirement); got != tt.want {
			t.Errorf("RequirementSpec(%q) = %q, want %q", tt.requirement, got, tt.want)
		}
	}
}
