package rubygems

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// TestGemfileParser_BasicGem tests parsing a simple gem declaration with version
func TestGemfileParser_BasicGem(t *testing.T) {
	gemfileContent := `source 'https://rubygems.org'

gem 'rails', '6.0.4.7'
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].PackageManager != "rubygems" {
		t.Errorf("expected package manager 'rubygems', got %q", packages[0].PackageManager)
	}

	if packages[0].PackageName != "rails" {
		t.Errorf("expected package name 'rails', got %q", packages[0].PackageName)
	}

	if packages[0].Version != "6.0.4.7" {
		t.Errorf("expected version '6.0.4.7', got %q", packages[0].Version)
	}
}

// TestGemfileParser_NoVersion tests parsing a gem without version specification
func TestGemfileParser_NoVersion(t *testing.T) {
	gemfileContent := `gem 'puma'
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].Version != "latest" {
		t.Errorf("expected version 'latest', got %q", packages[0].Version)
	}
}

// TestGemfileParser_RangeVersion tests parsing a gem with range version specifier
func TestGemfileParser_RangeVersion(t *testing.T) {
	gemfileContent := `gem 'rails', '~> 6.0'
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].Version != "latest" {
		t.Errorf("expected version 'latest' (no lock file present), got %q", packages[0].Version)
	}
}

// TestGemfileParser_MultipleConstraints tests parsing a gem with multiple version constraints
func TestGemfileParser_MultipleConstraints(t *testing.T) {
	gemfileContent := `gem 'pg', '>= 0.18', '< 2.0'
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].PackageName != "pg" {
		t.Errorf("expected package name 'pg', got %q", packages[0].PackageName)
	}

	// First constraint is '>= 0.18'
	if packages[0].Version != "latest" {
		t.Errorf("expected version 'latest', got %q", packages[0].Version)
	}
}

// TestGemfileParser_SkipsComments tests that comment lines are skipped
func TestGemfileParser_SkipsComments(t *testing.T) {
	gemfileContent := `# This is a comment
# Another comment

gem 'rails', '6.0.4.7'
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].PackageName != "rails" {
		t.Errorf("expected package name 'rails', got %q", packages[0].PackageName)
	}
}

// TestGemfileParser_InlineComments tests that inline comments are stripped
func TestGemfileParser_InlineComments(t *testing.T) {
	gemfileContent := `gem 'rails', '6.0.4.7' # pinned for stability
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].PackageName != "rails" {
		t.Errorf("expected package name 'rails', got %q", packages[0].PackageName)
	}
}

// TestGemfileParser_GroupBlock tests that gems in groups are included
func TestGemfileParser_GroupBlock(t *testing.T) {
	gemfileContent := `gem 'rails', '6.0.4.7'

group :development do
  gem 'rspec', '~> 3.9'
end
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	// Both rails and rspec should be present
	names := make(map[string]bool)
	for _, pkg := range packages {
		names[pkg.PackageName] = true
	}

	if !names["rails"] || !names["rspec"] {
		t.Errorf("expected 'rails' and 'rspec' packages")
	}
}

// TestGemfileParser_WithLockFile tests version resolution using Gemfile.lock
func TestGemfileParser_WithLockFile(t *testing.T) {
	gemfileContent := `gem 'rails', '~> 6.0'
gem 'pg', '>= 0.18'
`

	lockfileContent := `GEM
  remote: https://rubygems.org/
  specs:
    rails (6.0.4.7)
    pg (1.1.4)

BUNDLED WITH
   2.2.33
`

	tempDir := t.TempDir()
	gemfilePath := filepath.Join(tempDir, "Gemfile")
	lockfilePath := filepath.Join(tempDir, "Gemfile.lock")

	if err := os.WriteFile(gemfilePath, []byte(gemfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	if err := os.WriteFile(lockfilePath, []byte(lockfileContent), 0644); err != nil {
		t.Fatalf("failed to write Gemfile.lock: %v", err)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	// Check that versions were resolved from lock file
	versionMap := make(map[string]string)
	for _, pkg := range packages {
		versionMap[pkg.PackageName] = pkg.Version
	}

	if versionMap["rails"] != "6.0.4.7" {
		t.Errorf("expected rails version '6.0.4.7', got %q", versionMap["rails"])
	}

	if versionMap["pg"] != "1.1.4" {
		t.Errorf("expected pg version '1.1.4', got %q", versionMap["pg"])
	}
}

// TestGemfileParser_Parse_RealFile tests parsing the real test fixture
func TestGemfileParser_Parse_RealFile(t *testing.T) {
	gemfilePath := "../../../test/resources/Gemfile"

	// Verify file exists
	if _, err := os.Stat(gemfilePath); err != nil {
		t.Skipf("test fixture not found at %s", gemfilePath)
	}

	parser := &GemfileParser{}
	packages, err := parser.Parse(gemfilePath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Should have: rails, pg, puma, byebug, web-console, listen, bundler
	if len(packages) < 5 {
		t.Fatalf("expected at least 5 packages, got %d", len(packages))
	}

	// Verify a few known packages
	nameMap := make(map[string]models.Package)
	for _, pkg := range packages {
		nameMap[pkg.PackageName] = pkg
	}

	expectedPackages := []string{"rails", "pg", "puma"}
	for _, expectedName := range expectedPackages {
		if _, ok := nameMap[expectedName]; !ok {
			t.Errorf("expected package %q not found", expectedName)
		}
	}

	// Check that version was resolved from lock file (Gemfile.lock exists in testdata)
	if railsPkg, ok := nameMap["rails"]; ok {
		if railsPkg.Version == "latest" {
			t.Errorf("rails version should be resolved from Gemfile.lock, got 'latest'")
		}
	}
}
