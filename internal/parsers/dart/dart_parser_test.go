package dart

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func packageMap(pkgs []models.Package) map[string]models.Package {
	m := make(map[string]models.Package, len(pkgs))
	for _, p := range pkgs {
		m[p.PackageName] = p
	}
	return m
}

// TestPubspecParsingWithLock parses the shared pubspec.yaml fixture, which has a
// sibling pubspec.lock, and verifies version resolution across every value shape.
func TestPubspecParsingWithLock(t *testing.T) {
	parser := &DartParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "pubspec.yaml"))
	require.NoError(t, err)

	m := packageMap(pkgs)

	// SDK-sourced deps are not pub.dev packages and must be skipped.
	assert.NotContains(t, m, "flutter")
	assert.NotContains(t, m, "flutter_test")

	// 11 real deps: 8 in dependencies (minus flutter) + 2 dev (minus flutter_test) + 1 override.
	assert.Len(t, pkgs, 11)

	expected := map[string]string{
		"http":           "0.13.6", // ^0.13.5 -> lock
		"provider":       "6.0.5",  // exact
		"path":           "1.8.3",  // '>=1.8.0 <2.0.0' -> lock
		"collection":     "1.17.0", // dependency_overrides exact wins the map lookup below
		"intl":           "1.0.5",  // ^1.0.0 -> lock
		"custom_git_pkg": "0.2.0",  // git source map -> lock
		"local_pkg":      "0.0.1",  // path source map -> lock
		"hosted_pkg":     "2.1.4",  // hosted with ^2.1.0 -> lock
		"test":           "1.24.6", // ^1.21.0 -> lock
		"mockito":        "5.4.2",  // exact
	}
	for name, want := range expected {
		pkg, ok := m[name]
		require.Truef(t, ok, "package %s not found", name)
		assert.Equalf(t, want, pkg.Version, "version for %s", name)
		assert.Equalf(t, "pub", pkg.PackageManager, "package manager for %s", name)
	}

	// collection appears twice (dependencies + dependency_overrides) with distinct lines.
	var collections []models.Package
	for _, p := range pkgs {
		if p.PackageName == "collection" {
			collections = append(collections, p)
		}
	}
	require.Len(t, collections, 2)

	// Location contract: 0-based line, 2-space indent for the http declaration.
	http := m["http"]
	require.Len(t, http.Locations, 1)
	assert.Equal(t, 12, http.Locations[0].Line)
	assert.Equal(t, 2, http.Locations[0].StartIndex)
	assert.Equal(t, len("  http: ^0.13.5"), http.Locations[0].EndIndex)
}

// TestPubspecLockParsing parses pubspec.lock as a standalone manifest.
func TestPubspecLockParsing(t *testing.T) {
	parser := &DartParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "pubspec.lock"))
	require.NoError(t, err)

	m := packageMap(pkgs)
	expected := map[string]string{
		"http":           "0.13.6",
		"path":           "1.8.3",
		"collection":     "1.17.2",
		"intl":           "1.0.5",
		"custom_git_pkg": "0.2.0",
		"local_pkg":      "0.0.1",
		"hosted_pkg":     "2.1.4",
		"test":           "1.24.6",
	}
	assert.Len(t, pkgs, len(expected))
	for name, want := range expected {
		pkg, ok := m[name]
		require.Truef(t, ok, "package %s not found", name)
		assert.Equalf(t, want, pkg.Version, "version for %s", name)
		assert.Equal(t, "pub", pkg.PackageManager)
	}

	// The sdks: section must not leak into the package list.
	assert.NotContains(t, m, "dart")
}

// TestPubspecWithoutLock verifies fallback behaviour when no pubspec.lock exists:
// exact versions are kept and ranged/source-map deps resolve to "latest".
func TestPubspecWithoutLock(t *testing.T) {
	content := `name: no_lock_app
environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter
  http: ^0.13.5
  provider: 6.0.5
  path: '>=1.8.0 <2.0.0'
  collection: any
  my_git:
    git:
      url: https://github.com/example/my_git.git
      ref: main

dev_dependencies:
  mockito: 5.4.2
`
	dir := t.TempDir()
	path := filepath.Join(dir, "pubspec.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	parser := &DartParser{}
	pkgs, err := parser.Parse(path)
	require.NoError(t, err)

	m := packageMap(pkgs)
	assert.NotContains(t, m, "flutter")

	expected := map[string]string{
		"http":       "latest", // ^0.13.5, no lock
		"provider":   "6.0.5",  // exact
		"path":       "latest", // ranged, no lock
		"collection": "latest", // any, no lock
		"my_git":     "latest", // git source map, no lock
		"mockito":    "5.4.2",  // exact
	}
	assert.Len(t, pkgs, len(expected))
	for name, want := range expected {
		pkg, ok := m[name]
		require.Truef(t, ok, "package %s not found", name)
		assert.Equalf(t, want, pkg.Version, "version for %s", name)
	}
}

// TestFourSpaceIndentation ensures non-2-space indentation is still parsed.
func TestFourSpaceIndentation(t *testing.T) {
	content := "name: four_space_app\n\ndependencies:\n    http: ^0.13.5\n    provider: 6.0.5\n    my_git:\n        git:\n            url: https://example.com/g.git\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "pubspec.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	parser := &DartParser{}
	pkgs, err := parser.Parse(path)
	require.NoError(t, err)

	m := packageMap(pkgs)
	require.Len(t, pkgs, 3)
	assert.Equal(t, "latest", m["http"].Version)    // ^, no lock
	assert.Equal(t, "6.0.5", m["provider"].Version) // exact
	assert.Equal(t, "latest", m["my_git"].Version)  // git source map, no lock
	assert.Equal(t, 4, m["http"].Locations[0].StartIndex)
}

// TestVersionConstraintForms exercises every constraint shape (no lock present).
func TestVersionConstraintForms(t *testing.T) {
	content := `name: constraints_app
dependencies:
  exact: 1.2.3
  caret: ^1.2.3
  range: '>=1.2.3 <2.0.0'
  gt: '>1.0.0'
  lt: '<2.0.0'
  anydep: any
  prerelease: 1.0.0-beta.1
`
	dir := t.TempDir()
	path := filepath.Join(dir, "pubspec.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	parser := &DartParser{}
	pkgs, err := parser.Parse(path)
	require.NoError(t, err)

	m := packageMap(pkgs)
	expected := map[string]string{
		"exact":      "1.2.3",
		"caret":      "latest",
		"range":      "latest",
		"gt":         "latest",
		"lt":         "latest",
		"anydep":     "latest",
		"prerelease": "1.0.0-beta.1",
	}
	assert.Len(t, pkgs, len(expected))
	for name, want := range expected {
		assert.Equalf(t, want, m[name].Version, "version for %s", name)
	}
}

// TestDuplicateAcrossSections ensures the same package in multiple sections yields
// one entry per declaration site, each with its own version and line.
func TestDuplicateAcrossSections(t *testing.T) {
	content := `name: dup_app
dependencies:
  dup: 1.0.0
dev_dependencies:
  dup: ^2.0.0
dependency_overrides:
  dup: 3.0.0
`
	dir := t.TempDir()
	path := filepath.Join(dir, "pubspec.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	parser := &DartParser{}
	pkgs, err := parser.Parse(path)
	require.NoError(t, err)

	require.Len(t, pkgs, 3)
	versions := []string{pkgs[0].Version, pkgs[1].Version, pkgs[2].Version}
	assert.Equal(t, []string{"1.0.0", "latest", "3.0.0"}, versions)
	// Distinct, increasing line numbers (file-order scan).
	assert.True(t, pkgs[0].Locations[0].Line < pkgs[1].Locations[0].Line)
	assert.True(t, pkgs[1].Locations[0].Line < pkgs[2].Locations[0].Line)
}

// TestUnsupportedFile ensures a non-Dart file name is rejected.
func TestUnsupportedFile(t *testing.T) {
	parser := &DartParser{}
	_, err := parser.Parse("something.txt")
	assert.Error(t, err)
}

// TestCRLFHandling ensures EndIndex is not inflated by trailing \r on CRLF files.
func TestCRLFHandling(t *testing.T) {
	content := "name: crlf_app\r\ndependencies:\r\n  http: ^0.13.5\r\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "pubspec.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	parser := &DartParser{}
	pkgs, err := parser.Parse(path)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	assert.Equal(t, len("  http: ^0.13.5"), pkgs[0].Locations[0].EndIndex)
}
