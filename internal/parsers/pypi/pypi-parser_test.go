package pypi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// comparePackages is a helper to assert Package equality in tests.
func comparePackages(t *testing.T, got, want models.Package) {
	if got.PackageManager != want.PackageManager {
		t.Errorf("PackageManager: got %q, want %q", got.PackageManager, want.PackageManager)
	}
	if got.PackageName != want.PackageName {
		t.Errorf("PackageName: got %q, want %q", got.PackageName, want.PackageName)
	}
	if got.Version != want.Version {
		t.Errorf("Version: got %q, want %q", got.Version, want.Version)
	}
	if got.FilePath != want.FilePath {
		t.Errorf("FilePath: got %q, want %q", got.FilePath, want.FilePath)
	}
	if got.LineStart != want.LineStart || got.LineEnd != want.LineEnd {
		t.Errorf("LineStart/LineEnd: got %d/%d, want %d/%d", got.LineStart, got.LineEnd, want.LineStart, want.LineEnd)
	}
	if got.StartIndex != want.StartIndex || got.EndIndex != want.EndIndex {
		t.Errorf("StartIndex/EndIndex: got %d/%d, want %d/%d", got.StartIndex, got.EndIndex, want.StartIndex, want.EndIndex)
	}
}

func TestParseExactVersion(t *testing.T) {
	content := "flask==1.1.2\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "flask",
		Version:        "1.1.2",
		FilePath:       filePath,
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     1,
		EndIndex:       12,
	}
	comparePackages(t, got, want)
}

func TestParseInlineComment(t *testing.T) {
	content := "   requests==2.25.1  # pinned for compatibility\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "requests",
		Version:        "2.25.1",
		FilePath:       filePath,
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     4,
		EndIndex:       19,
	}
	comparePackages(t, got, want)
}

func TestParseRequirementLineEndIndex(t *testing.T) {
	content := "requests==2.25.1  # pinned for compatibility\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "requests",
		Version:        "2.25.1",
		FilePath:       filePath,
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     1,
		EndIndex:       16,
	}
	comparePackages(t, got, want)
}

func TestParseVersionRange(t *testing.T) {
	content := "django>=3.0,<4.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "django",
		Version:        "latest",
		FilePath:       filePath,
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     1,
		EndIndex:       16,
	}
	comparePackages(t, got, want)
}

func TestParseSkipCommentLine(t *testing.T) {
	content := "# just a comment\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseWildcardVersion(t *testing.T) {
	content := "pandas==1.2.*\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "pandas",
		Version:        "latest", // Should treat wildcard as "latest"
		FilePath:       filePath,
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     1,
		EndIndex:       13, // Length of "pandas==1.2.*"
	}
	comparePackages(t, got, want)

	// Additional check to ensure wildcard was handled correctly
	if got.Version != "latest" {
		t.Errorf("Wildcard version not properly handled: got %q, want %q", got.Version, "latest")
	}
}

func TestParseRealRequirementsFile(t *testing.T) {
	filePath := "../../testdata/requirements.txt"
	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 25 {
		t.Fatalf("expected 25 packages, got %d", len(pkgs))
	}

	// Print all packages for inspection
	for _, pkg := range pkgs {
		t.Logf("Found package: %s==%s (line %d)", pkg.PackageName, pkg.Version, pkg.LineStart)
	}
}
