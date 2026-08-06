package composer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestBasicComposerJsonParsing(t *testing.T) {
	composerJSON := `{
		"require": {
			"guzzlehttp/guzzle": "7.4.0",
			"symfony/console": "^5.0"
		},
		"require-dev": {
			"phpunit/phpunit": "^9.5"
		}
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(packages))
	}

	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	expectedPackages := []struct {
		name    string
		version string
		manager string
	}{
		{"guzzlehttp/guzzle", "7.4.0", "packagist"},
		{"symfony/console", "latest", "packagist"},
		{"phpunit/phpunit", "latest", "packagist"},
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

		if len(pkg.Locations) == 0 || pkg.Locations[0].Line == 0 {
			t.Errorf("package %s: missing position info", expected.name)
		}
	}
}

func TestComposerJsonWithLockFile(t *testing.T) {
	composerJSON := `{
		"require": {
			"caret-dep": "^1.0.0",
			"tilde-dep": "~2.0.0",
			"exact-dep": "3.1.0"
		}
	}`

	composerLock := `{
		"packages": [
			{"name": "caret-dep", "version": "1.2.3"},
			{"name": "tilde-dep", "version": "2.0.9"},
			{"name": "exact-dep", "version": "3.1.0"}
		],
		"packages-dev": []
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")
	lockPath := filepath.Join(tempDir, "composer.lock")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	if err := os.WriteFile(lockPath, []byte(composerLock), 0644); err != nil {
		t.Fatalf("failed to write composer.lock: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	expectedVersions := map[string]string{
		"caret-dep": "1.2.3",
		"tilde-dep": "2.0.9",
		"exact-dep": "3.1.0",
	}

	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

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

func TestComposerJsonDevDepsResolvedFromPackagesDev(t *testing.T) {
	composerJSON := `{
		"require": {},
		"require-dev": {
			"phpunit/phpunit": "^9.5"
		}
	}`

	composerLock := `{
		"packages": [],
		"packages-dev": [
			{"name": "phpunit/phpunit", "version": "9.5.28"}
		]
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")
	lockPath := filepath.Join(tempDir, "composer.lock")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	if err := os.WriteFile(lockPath, []byte(composerLock), 0644); err != nil {
		t.Fatalf("failed to write composer.lock: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].Version != "9.5.28" {
		t.Errorf("expected version 9.5.28, got %s", packages[0].Version)
	}
}

func TestComposerJsonSkipsPhpAndExtensions(t *testing.T) {
	composerJSON := `{
		"require": {
			"php": ">=7.4",
			"ext-json": "*",
			"lib-xml": "*",
			"vendor/real-package": "1.0.0"
		}
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].PackageName != "vendor/real-package" {
		t.Errorf("expected vendor/real-package, got %s", packages[0].PackageName)
	}
}

func TestMissingComposerLock(t *testing.T) {
	composerJSON := `{
		"require": {
			"vendor/exact": "1.2.3",
			"vendor/caret": "^2.0"
		}
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	expectedVersions := map[string]string{
		"vendor/exact": "1.2.3",
		"vendor/caret": "latest",
	}

	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

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

func TestCorruptedComposerLock(t *testing.T) {
	composerJSON := `{
		"require": {
			"vendor/pkg": "^1.0"
		}
	}`

	composerLock := `{ invalid json }`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")
	lockPath := filepath.Join(tempDir, "composer.lock")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	if err := os.WriteFile(lockPath, []byte(composerLock), 0644); err != nil {
		t.Fatalf("failed to write composer.lock: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].Version != "latest" {
		t.Errorf("expected version latest, got %s", packages[0].Version)
	}
}

func TestMalformedComposerJson(t *testing.T) {
	composerJSON := `{
		"require": {
			"broken-dep": "1.0.0",`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	_, err := parser.Parse(composerPath)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestPositionTracking(t *testing.T) {
	composerJSON := `{
	"require": {
		"dep1": "1.0.0",
		"dep2": "^2.0.0"
	}
}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	expectedPositions := map[string]struct {
		line int
	}{
		"dep1": {2},
		"dep2": {3},
	}

	for name, expected := range expectedPositions {
		pkg, exists := packageMap[name]
		if !exists {
			t.Errorf("package %s not found", name)
			continue
		}

		if pkg.Locations[0].Line != expected.line {
			t.Errorf("package %s: expected line %d, got %d",
				name, expected.line, pkg.Locations[0].Line)
		}

		if pkg.Locations[0].StartIndex <= 0 || pkg.Locations[0].EndIndex <= 0 {
			t.Errorf("package %s: invalid column indices: start=%d, end=%d",
				name, pkg.Locations[0].StartIndex, pkg.Locations[0].EndIndex)
		}
	}
}

func TestComposerJsonAllVersionConstraints(t *testing.T) {
	composerJSON := `{
		"require": {
			"vendor/caret": "^1.0.0",
			"vendor/tilde": "~1.0",
			"vendor/comparison": ">=1.0.0 <2.0",
			"vendor/wildcard": "1.0.*",
			"vendor/hyphen": "1.0.0 - 2.0.0",
			"vendor/or": ">=1.0 <1.1 || >=1.2",
			"vendor/exact": "1.5.0"
		}
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	expectedVersions := map[string]string{
		"vendor/caret":      "latest",
		"vendor/tilde":      "latest",
		"vendor/comparison": "latest",
		"vendor/wildcard":   "latest",
		"vendor/hyphen":     "latest",
		"vendor/or":         "latest",
		"vendor/exact":      "1.5.0",
	}

	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

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

func TestSamePackageInBothSections(t *testing.T) {
	composerJSON := `{
		"require": {
			"vendor/pkg": "1.0.0"
		},
		"require-dev": {
			"vendor/pkg": "^2.0"
		}
	}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(packages) != 2 {
		t.Fatalf("expected 2 packages (same name in different sections), got %d", len(packages))
	}

	if packages[0].Locations[0].Line >= packages[1].Locations[0].Line {
		t.Error("expected packages sorted by line number")
	}

	if packages[0].Version != "1.0.0" {
		t.Errorf("expected first package version 1.0.0, got %s", packages[0].Version)
	}

	if packages[1].Version != "latest" {
		t.Errorf("expected second package version latest, got %s", packages[1].Version)
	}
}

func TestSectionAwareParsing_Composer(t *testing.T) {
	composerJSON := `{
	"name": "vendor/myapp",
	"description": "My app",
	"require": {
		"vendor/dep1": "1.0.0",
		"myapp": "2.0.0"
	},
	"require-dev": {
		"vendor/dep2": "^1.0"
	}
}`

	tempDir := t.TempDir()
	composerPath := filepath.Join(tempDir, "composer.json")

	if err := os.WriteFile(composerPath, []byte(composerJSON), 0644); err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	parser := &ComposerJsonParser{}
	packages, err := parser.Parse(composerPath)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	packageMap := make(map[string]models.Package)
	for _, pkg := range packages {
		packageMap[pkg.PackageName] = pkg
	}

	if _, exists := packageMap["vendor/dep1"]; !exists {
		t.Error("vendor/dep1 not found")
	}

	if _, exists := packageMap["myapp"]; !exists {
		t.Error("myapp not found")
	}

	if _, exists := packageMap["vendor/dep2"]; !exists {
		t.Error("vendor/dep2 not found")
	}

	if appPkg, exists := packageMap["myapp"]; exists {
		if appPkg.Locations[0].Line < 4 {
			t.Errorf("myapp found at wrong location (line %d), expected >= 4 (in require section)",
				appPkg.Locations[0].Line)
		}
	}

	if len(packages) != 3 {
		t.Errorf("expected 3 packages, got %d", len(packages))
	}
}
