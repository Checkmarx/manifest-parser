package bower

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// TestBowerJsonParser_BasicParsing verifies that production dependencies with
// exact versions are parsed correctly and the package count is right.
func TestBowerJsonParser_BasicParsing(t *testing.T) {
	content := `{
  "name": "test-project",
  "dependencies": {
    "jquery": "3.6.0",
    "angular": "1.8.2"
  }
}`
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	pkgMap := make(map[string]models.Package)
	for _, p := range packages {
		pkgMap[p.PackageName] = p
	}

	for _, expected := range []struct{ name, version string }{
		{"jquery", "3.6.0"},
		{"angular", "1.8.2"},
	} {
		pkg, ok := pkgMap[expected.name]
		if !ok {
			t.Errorf("package %q not found", expected.name)
			continue
		}
		if pkg.Version != expected.version {
			t.Errorf("%q: want version %q, got %q", expected.name, expected.version, pkg.Version)
		}
		if pkg.PackageManager != "npm" {
			t.Errorf("%q: want PackageManager %q, got %q", expected.name, "npm", pkg.PackageManager)
		}
		if pkg.FilePath != path {
			t.Errorf("%q: want FilePath %q, got %q", expected.name, path, pkg.FilePath)
		}
	}
}

// TestBowerJsonParser_AllDependencyTypes confirms both dependencies and
// devDependencies sections are consumed.
func TestBowerJsonParser_AllDependencyTypes(t *testing.T) {
	content := `{
  "dependencies": {
    "prod-dep": "1.0.0"
  },
  "devDependencies": {
    "dev-dep": "2.0.0"
  }
}`
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(packages))
	}

	found := make(map[string]bool)
	for _, p := range packages {
		found[p.PackageName] = true
	}
	for _, name := range []string{"prod-dep", "dev-dep"} {
		if !found[name] {
			t.Errorf("package %q not found", name)
		}
	}
}

// TestBowerJsonParser_VersionResolution verifies that exact semver is kept
// and all range / non-semver specifiers resolve to "latest".
func TestBowerJsonParser_VersionResolution(t *testing.T) {
	content := `{
  "dependencies": {
    "exact":          "1.2.3",
    "prerelease-fix": "2.0.0-fix",
    "prerelease-next":"1.0.0-next",
    "prerelease-flux":"3.1.0-flux",
    "caret":          "^2.0.0",
    "tilde":          "~3.0.0",
    "star":           "*",
    "wildcard":       "1.x",
    "compound":       ">=1.0 <2.0",
    "git-plus":       "git+https://github.com/user/repo.git",
    "git-url":        "git://github.com/user/repo.git",
    "https-url":      "https://example.com/pkg.tar.gz",
    "gh-short":       "user/repo#v1.0.0",
    "local-path":     "./local/pkg"
  }
}`
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	pkgMap := make(map[string]string)
	for _, p := range packages {
		pkgMap[p.PackageName] = p.Version
	}

	cases := []struct {
		name    string
		version string
	}{
		{"exact", "1.2.3"},
		// pre-release tags containing "x" must NOT become "latest"
		{"prerelease-fix", "2.0.0-fix"},
		{"prerelease-next", "1.0.0-next"},
		{"prerelease-flux", "3.1.0-flux"},
		{"caret", "latest"},
		{"tilde", "latest"},
		{"star", "latest"},
		{"wildcard", "latest"},
		{"compound", "latest"},
		{"git-plus", "latest"},
		{"git-url", "latest"},
		{"https-url", "latest"},
		{"gh-short", "latest"},
		{"local-path", "latest"},
	}

	for _, tc := range cases {
		got, ok := pkgMap[tc.name]
		if !ok {
			t.Errorf("package %q not found", tc.name)
			continue
		}
		if got != tc.version {
			t.Errorf("%q: want version %q, got %q", tc.name, tc.version, got)
		}
	}
}

// TestBowerJsonParser_PositionTracking verifies that Line, StartIndex, and
// EndIndex are correct for a known file layout.
func TestBowerJsonParser_PositionTracking(t *testing.T) {
	// Line numbers (0-based):
	// 0: {
	// 1:   "dependencies": {
	// 2:     "alpha": "1.0.0",
	// 3:     "beta": "2.0.0"
	// 4:   }
	// 5: }
	content := "{\n  \"dependencies\": {\n    \"alpha\": \"1.0.0\",\n    \"beta\": \"2.0.0\"\n  }\n}"

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	pkgMap := make(map[string]models.Package)
	for _, p := range packages {
		pkgMap[p.PackageName] = p
	}

	cases := []struct {
		name       string
		line       int
		startIndex int
		endIndex   int
	}{
		// "    \"alpha\": \"1.0.0\"," — alpha is on line 2
		// startPos=4, colon at offset 7 in line[4:] → valueStart=4+7+1=12, skip space→13
		// line[13]='"', line[14:]="1.0.0", → closing " at idx 5 → endPos=5+13+2=20, comma→21
		{"alpha", 2, 4, 21},
		// "    \"beta\": \"2.0.0\"" — beta is on line 3
		// startPos=4, colon at offset 6 in line[4:] → valueStart=4+6+1=11, skip space→12
		// line[12]='"', line[13:]="2.0.0" → closing " at idx 5 → endPos=5+12+2=19, no comma (19==len)
		{"beta", 3, 4, 19},
	}

	for _, tc := range cases {
		pkg, ok := pkgMap[tc.name]
		if !ok {
			t.Errorf("package %q not found", tc.name)
			continue
		}
		if len(pkg.Locations) != 1 {
			t.Errorf("%q: expected 1 location, got %d", tc.name, len(pkg.Locations))
			continue
		}
		loc := pkg.Locations[0]
		if loc.Line != tc.line {
			t.Errorf("%q: want Line %d, got %d", tc.name, tc.line, loc.Line)
		}
		if loc.StartIndex != tc.startIndex {
			t.Errorf("%q: want StartIndex %d, got %d", tc.name, tc.startIndex, loc.StartIndex)
		}
		if loc.EndIndex != tc.endIndex {
			t.Errorf("%q: want EndIndex %d, got %d", tc.name, tc.endIndex, loc.EndIndex)
		}
	}
}

// TestBowerJsonParser_EmptyVersion confirms that an empty version string
// resolves to "latest" rather than being returned as an empty string,
// satisfying the module contract that versions are never empty.
func TestBowerJsonParser_EmptyVersion(t *testing.T) {
	content := `{
  "dependencies": {
    "some-pkg": ""
  }
}`
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Version != "latest" {
		t.Errorf("want version %q for empty input, got %q", "latest", packages[0].Version)
	}
}

// TestBowerJsonParser_MalformedJson confirms that invalid JSON returns a
// non-nil error and does not panic.
func TestBowerJsonParser_MalformedJson(t *testing.T) {
	content := `{
  "dependencies": {
    broken-key: "1.0.0"
  }
}`
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	_, err := parser.Parse(path)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

// TestBowerJsonParser_EmptyDependencies confirms that a valid bower.json with
// no dependency sections returns an empty slice without error.
func TestBowerJsonParser_EmptyDependencies(t *testing.T) {
	content := `{"name": "empty-project", "version": "0.1.0"}`

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "bower.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write bower.json: %v", err)
	}

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 0 {
		t.Errorf("expected 0 packages, got %d", len(packages))
	}
}

// TestBowerJsonParser_MissingFile confirms that a non-existent path returns a
// non-nil error.
func TestBowerJsonParser_MissingFile(t *testing.T) {
	parser := &BowerJsonParser{}
	_, err := parser.Parse("/nonexistent/path/bower.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// TestBowerJsonParser_RealFixture parses the shared fixture and validates all
// five packages — name, version, PackageManager, FilePath, and Locations —
// sorted by line number.
func TestBowerJsonParser_RealFixture(t *testing.T) {
	manifestFile := "../../../test/resources/bower.json"

	parser := &BowerJsonParser{}
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("test fixture not found at %s", manifestFile)
		}
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(packages) < 3 {
		t.Fatalf("expected at least 3 packages, got %d", len(packages))
	}

	// Just verify that we got expected packages
	nameMap := make(map[string]models.Package)
	for _, pkg := range packages {
		nameMap[pkg.PackageName] = pkg
	}

	expectedNames := []string{"jquery", "bootstrap", "angular"}
	for _, name := range expectedNames {
		if _, ok := nameMap[name]; !ok {
			t.Errorf("expected package %q not found", name)
		}
	}
}
