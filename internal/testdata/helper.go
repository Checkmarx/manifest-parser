package testdata

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// CompareLocations is a helper to assert Location equality in tests
func CompareLocations(t *testing.T, got, want []models.Location) {
	if len(got) != len(want) {
		t.Errorf("Locations: got %d locations, want %d", len(got), len(want))
		return
	}

	for i, loc := range got {
		if loc.Line != want[i].Line {
			t.Errorf("Location[%d].Line: got %d, want %d", i, loc.Line, want[i].Line)
		}
		if loc.StartIndex != want[i].StartIndex {
			t.Errorf("Location[%d].StartIndex: got %d, want %d", i, loc.StartIndex, want[i].StartIndex)
		}
		if loc.EndIndex != want[i].EndIndex {
			t.Errorf("Location[%d].EndIndex: got %d, want %d", i, loc.EndIndex, want[i].EndIndex)
		}
	}
}

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
	CompareLocations(t, got.Locations, want.Locations)
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
