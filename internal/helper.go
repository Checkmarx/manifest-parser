package internal

import (
	"testing"
)

func ValidatePackages(t *testing.T, packages []Package, expectedPackages []Package) {
	if len(packages) != len(expectedPackages) {
		t.Errorf("Expected %d packages, got %d", len(expectedPackages), len(packages))
	}

	for i, pkg := range packages {
		if pkg.PackageName != expectedPackages[i].PackageName {
			t.Errorf("Expected package name %s, got %s", expectedPackages[i].PackageName, pkg.PackageName)
		}
		if pkg.Version != expectedPackages[i].Version {
			t.Errorf("Expected package version %s, got %s", expectedPackages[i].Version, pkg.Version)
		}
		if pkg.LineStart != expectedPackages[i].LineStart {
			t.Errorf("Expected package line start %d, got %d", expectedPackages[i].LineStart, pkg.LineStart)
		}
		if pkg.LineEnd != expectedPackages[i].LineEnd {
			t.Errorf("Expected package line end %d, got %d", expectedPackages[i].LineEnd, pkg.LineEnd)
		}
		if pkg.Filepath != expectedPackages[i].Filepath {
			t.Errorf("Expected package filepath %s, got %s", expectedPackages[i].Filepath, pkg.Filepath)
		}
	}
}
