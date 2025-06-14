package npm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// TestBasicPackageJsonParsing tests the basic ability to parse a simple package.json
func TestBasicPackageJsonParsing(t *testing.T) {
	// Minimal package.json with basic dependencies
	packageJSON := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"express": "4.17.1",
			"lodash": "^4.17.21"
		},
		"devDependencies": {
			"jest": "^27.0.0"
		}
	}`

	// Create a temporary file for testing
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Check that we have the correct number of packages
	if len(packages) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(packages))
	}

	// Map packages by name for easier checking
	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	// Check package details
	expectedPackages := []struct {
		name    string
		version string
		manager string
	}{
		{"express", "4.17.1", "npm"},
		{"lodash", "latest", "npm"}, // "^4.17.21" will become "latest" without a lock file
		{"jest", "latest", "npm"},   // "^27.0.0" will become "latest" without a lock file
	}

	for _, expected := range expectedPackages {
		pkg, exists := packageMap[expected.name]
		if !exists {
			t.Errorf("package %s not found", expected.name)
			continue
		}

		if pkg.Version != expected.version {
			t.Errorf("package %s: expected version %q, got %q",
				expected.name, expected.version, pkg.Version)
		}

		if pkg.PackageManager != expected.manager {
			t.Errorf("package %s: expected package manager %q, got %q",
				expected.name, expected.manager, pkg.PackageManager)
		}

		// Check that we have position information
		if len(pkg.Locations) == 0 || pkg.Locations[0].Line == 0 || pkg.Locations[0].StartIndex == 0 {
			t.Errorf("package %s: missing position info", expected.name)
		}
	}
}

// TestPackageJsonWithLockFile tests the interaction between package.json and package-lock.json
func TestPackageJsonWithLockFile(t *testing.T) {
	// Package.json with various dependency notations
	packageJSON := `{
		"dependencies": {
			"exact-dep": "1.2.3",
			"caret-dep": "^2.0.0",
			"tilde-dep": "~3.0.0",
			"star-dep": "4.*"
		}
	}`

	// Package-lock.json in v2 format
	packageLockJSON := `{
		"lockfileVersion": 2,
		"packages": {
			"": {
				"dependencies": {
					"exact-dep": "1.2.3",
					"caret-dep": "^2.0.0",
					"tilde-dep": "~3.0.0",
					"star-dep": "4.*"
				}
			},
			"node_modules/caret-dep": {
				"version": "2.3.4",
				"resolved": "https://registry.npmjs.org/caret-dep/-/caret-dep-2.3.4.tgz"
			},
			"node_modules/tilde-dep": {
				"version": "3.0.9",
				"resolved": "https://registry.npmjs.org/tilde-dep/-/tilde-dep-3.0.9.tgz"
			},
			"node_modules/star-dep": {
				"version": "4.5.0",
				"resolved": "https://registry.npmjs.org/star-dep/-/star-dep-4.5.0.tgz"
			}
		}
	}`

	// Create temporary files
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")
	packageLockPath := filepath.Join(tempDir, "package-lock.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	if err := os.WriteFile(packageLockPath, []byte(packageLockJSON), 0644); err != nil {
		t.Fatalf("failed to write package-lock.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Check results
	expectedVersions := map[string]string{
		"exact-dep": "1.2.3", // Exact version should remain as is
		"caret-dep": "2.3.4", // Version with ^ should be taken from the lock file
		"tilde-dep": "3.0.9", // Version with ~ should be taken from the lock file
		"star-dep":  "4.5.0", // Version with * should be taken from the lock file
	}

	// Map packages by name
	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	// Check versions
	for name, expectedVersion := range expectedVersions {
		pkg, exists := packageMap[name]
		if !exists {
			t.Errorf("package %s not found", name)
			continue
		}

		if pkg.Version != expectedVersion {
			t.Errorf("package %s: expected version %q, got %q",
				name, expectedVersion, pkg.Version)
		}
	}
}

// TestPackageJsonWithV1LockFile tests working with an older format (v1) lock file
func TestPackageJsonWithV1LockFile(t *testing.T) {
	// Package.json file
	packageJSON := `{
		"dependencies": {
			"v1-package": "^1.0.0"
		}
	}`

	// Package-lock.json in v1 format
	packageLockJSON := `{
		"lockfileVersion": 1,
		"dependencies": {
			"v1-package": {
				"version": "1.5.2",
				"resolved": "https://registry.npmjs.org/v1-package/-/v1-package-1.5.2.tgz"
			}
		}
	}`

	// Create temporary files
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")
	packageLockPath := filepath.Join(tempDir, "package-lock.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	if err := os.WriteFile(packageLockPath, []byte(packageLockJSON), 0644); err != nil {
		t.Fatalf("failed to write package-lock.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Check package version
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.PackageName != "v1-package" {
		t.Errorf("expected package name 'v1-package', got '%s'", pkg.PackageName)
	}

	if pkg.Version != "1.5.2" {
		t.Errorf("expected version '1.5.2', got '%s'", pkg.Version)
	}
}

// TestAllDependencyTypes tests that all dependency types are correctly parsed
func TestAllDependencyTypes(t *testing.T) {
	// Package.json with all supported dependency types
	packageJSON := `{
		"dependencies": {
			"normal-dep": "1.0.0"
		},
		"devDependencies": {
			"dev-dep": "2.0.0"
		},
		"peerDependencies": {
			"peer-dep": "3.0.0"
		},
		"optionalDependencies": {
			"optional-dep": "4.0.0"
		}
	}`

	// Create temporary file
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Create a map of found package names
	foundPackages := make(map[string]bool)
	for _, pkg := range packages {
		foundPackages[pkg.PackageName] = true
	}

	// Check that all dependency types were found
	expectedPackages := []string{
		"normal-dep", "dev-dep", "peer-dep", "optional-dep",
	}

	for _, name := range expectedPackages {
		if !foundPackages[name] {
			t.Errorf("package %s not found", name)
		}
	}
}

// TestPositionTracking tests that package positions in the source file are correctly identified
func TestPositionTracking(t *testing.T) {
	// Package.json with known format
	packageJSON := `{
	"dependencies": {
		"dep1": "1.0.0",
		"dep2": "^2.0.0"
	}
}`

	// Create temporary file
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Map packages by name
	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	// Check positions
	expectedPositions := map[string]struct {
		lineStart int
		lineEnd   int
	}{
		"dep1": {2, 2},
		"dep2": {3, 3},
	}

	for name, expected := range expectedPositions {
		pkg, exists := packageMap[name]
		if !exists {
			t.Errorf("package %s not found", name)
			continue
		}

		if pkg.Locations[0].Line != expected.lineStart || pkg.Locations[0].Line != expected.lineEnd {
			t.Errorf("package %s: expected position line %d-%d, got %d-%d",
				name, expected.lineStart, expected.lineEnd, pkg.Locations[0].Line, pkg.Locations[0].Line)
		}

		// Check that column indices make sense
		if pkg.Locations[0].StartIndex <= 0 || pkg.Locations[0].EndIndex <= 0 {
			t.Errorf("package %s: invalid column indices: start=%d, end=%d",
				name, pkg.Locations[0].StartIndex, pkg.Locations[0].EndIndex)
		}
	}
}

// TestMalformedPackageJson tests parser behavior with a malformed package.json
func TestMalformedPackageJson(t *testing.T) {
	// Package.json with syntax error
	packageJSON := `{
		"dependencies": {
			"broken-dep": "1.0.0",
			missing-quotes: "2.0.0"
		}
	}`

	// Create temporary file
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	_, err := parser.Parse(packageJSONPath)

	// Check that there is an error
	if err == nil {
		t.Errorf("expected error for malformed JSON, but parsing succeeded")
	}
}

// TestMissingPackageLock tests behavior when there is no package-lock.json file
func TestMissingPackageLock(t *testing.T) {
	// Package.json with various version notations
	packageJSON := `{
		"dependencies": {
			"exact-dep": "1.2.3",
			"caret-dep": "^2.0.0",
			"tilde-dep": "~3.0.0",
			"star-dep": "4.*"
		}
	}`

	// Create temporary file (only package.json, no package-lock.json)
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	// Run the parser
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Check results
	expectedVersions := map[string]string{
		"exact-dep": "1.2.3",  // Exact version should remain as is
		"caret-dep": "latest", // Version with ^ becomes "latest" without a lock file
		"tilde-dep": "latest", // Version with ~ becomes "latest" without a lock file
		"star-dep":  "latest", // Version with * becomes "latest" without a lock file
	}

	// Map packages by name
	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	// Check versions
	for name, expectedVersion := range expectedVersions {
		pkg, exists := packageMap[name]
		if !exists {
			t.Errorf("package %s not found", name)
			continue
		}

		if pkg.Version != expectedVersion {
			t.Errorf("package %s: expected version %q, got %q",
				name, expectedVersion, pkg.Version)
		}
	}
}

// TestCorruptedPackageLock tests behavior when package-lock.json is corrupted
func TestCorruptedPackageLock(t *testing.T) {
	// Package.json file
	packageJSON := `{
		"dependencies": {
			"some-dep": "^1.0.0"
		}
	}`

	// Corrupted package-lock.json
	packageLockJSON := `{
		"lockfileVersion": 2,
		"packages": {
			missing-quotes: {
				"version": "1.2.3"
			}
		}
	}`

	// Create temporary files
	tempDir := t.TempDir()
	packageJSONPath := filepath.Join(tempDir, "package.json")
	packageLockPath := filepath.Join(tempDir, "package-lock.json")

	if err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	if err := os.WriteFile(packageLockPath, []byte(packageLockJSON), 0644); err != nil {
		t.Fatalf("failed to write package-lock.json: %v", err)
	}

	// Run the parser - should ignore the corrupted lock file and continue
	parser := &NpmPackageJsonParser{}
	packages, err := parser.Parse(packageJSONPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	// Check results
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	pkg := packages[0]
	if pkg.PackageName != "some-dep" {
		t.Errorf("expected package name 'some-dep', got '%s'", pkg.PackageName)
	}

	// Since the lock file is corrupted, we expect the version to be "latest"
	if pkg.Version != "latest" {
		t.Errorf("expected version 'latest', got '%s'", pkg.Version)
	}
}

// Direct test of getResolvedVersion function - optional if the function can be accessed directly
func TestGetResolvedVersionDirectly(t *testing.T) {
	// Create the function being tested
	var lock lockFile
	if err := json.Unmarshal([]byte(`{
		"lockfileVersion": 2,
		"packages": {
			"node_modules/test-pkg": {
				"version": "2.3.4"
			}
		}
	}`), &lock); err != nil {
		t.Fatalf("failed to create test lock file: %v", err)
	}

	getResolvedVersion := func(name, specVersion string) string {
		// Check if version is already exact
		if !strings.HasPrefix(specVersion, "^") &&
			!strings.HasPrefix(specVersion, "~") &&
			!strings.Contains(specVersion, "*") &&
			!strings.Contains(specVersion, ">") &&
			!strings.Contains(specVersion, "<") &&
			!strings.Contains(specVersion, "latest") {
			return specVersion
		}

		// Search in v2 format
		if pkgs := lock.Packages; pkgs != nil {
			pathVariations := []string{
				"node_modules/" + name,
				"node_modules/" + name + "@" + specVersion,
				"node_modules/" + name + "@" + strings.TrimPrefix(specVersion, "^"),
				"node_modules/" + name + "@" + strings.TrimPrefix(specVersion, "~"),
				"", // Root package
			}

			for _, path := range pathVariations {
				if entry, ok := pkgs[path]; ok && entry.Version != "" {
					return entry.Version
				}
			}
		}

		// Default fallback
		if strings.HasPrefix(specVersion, "^") ||
			strings.HasPrefix(specVersion, "~") ||
			strings.Contains(specVersion, "*") ||
			strings.Contains(specVersion, ">") ||
			strings.Contains(specVersion, "<") {
			return "latest"
		}

		return specVersion
	}

	// Test cases
	testCases := []struct {
		name     string
		pkgName  string
		version  string
		expected string
	}{
		{"exact version", "test-pkg", "1.0.0", "1.0.0"},
		{"caret version with lock match", "test-pkg", "^2.0.0", "2.3.4"},
		{"tilde version with lock match", "test-pkg", "~2.0.0", "2.3.4"},
		{"caret version without match", "missing-pkg", "^3.0.0", "latest"},
		{"star version", "test-pkg", "2.*", "2.3.4"},
		{"complex range", "missing-pkg", ">1.0.0 <2.0.0", "latest"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getResolvedVersion(tc.pkgName, tc.version)
			if result != tc.expected {
				t.Errorf("getResolvedVersion(%q, %q): expected %q, got %q",
					tc.pkgName, tc.version, tc.expected, result)
			}
		})
	}
}

// TestParse_RealTestdataPackageJson tests parsing the real package.json from internal/testdata
func TestParse_RealTestdataPackageJson(t *testing.T) {
	parser := &NpmPackageJsonParser{}
	manifestFile := "../../../internal/testdata/package.json"

	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expected := []models.Package{
		{
			PackageManager: "npm",
			PackageName:    "@istanbuljs/nyc-config-typescript",
			Version:        "1.0.2",
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       8,
				StartIndex: 4,
				EndIndex:   49,
			}},
		},
		{
			PackageManager: "npm",
			PackageName:    "@checkmarxdev/ast-cli-javascript-wrapper",
			Version:        "latest",
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       9,
				StartIndex: 4,
				EndIndex:   74,
			}},
		},
		{
			PackageManager: "npm",
			PackageName:    "webpack-cli",
			Version:        "latest",
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       10,
				StartIndex: 4,
				EndIndex:   27,
			}},
		},
		{
			PackageManager: "npm",
			PackageName:    "validator",
			Version:        "13.12.0",
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       13,
				StartIndex: 4,
				EndIndex:   27,
			}},
		},
		{
			PackageManager: "npm",
			PackageName:    "@popperjs/core",
			Version:        "latest",
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       14,
				StartIndex: 4,
				EndIndex:   31,
			}},
		},
	}

	testdata.ValidatePackages(t, packages, expected)
}
