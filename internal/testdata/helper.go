package testdata

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// ComparePackages is a helper to assert Package equality in tests
func ComparePackages(t *testing.T, got, want models.Package) {
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

func ValidatePackages(t *testing.T, packages []models.Package, expectedPackages []models.Package) {
	if len(packages) != len(expectedPackages) {
		t.Errorf("Expected %d packages, got %d", len(expectedPackages), len(packages))
		return
	}

	for i, pkg := range packages {
		ComparePackages(t, pkg, expectedPackages[i])
	}
}
