package csproj

import (
	"ManifestParser/internal"
	"testing"
)

func TestDotnetCsprojParser_ParseNoVersion(t *testing.T) {
	parser := &DotnetCsprojParser{}
	manifestFile := "../test/resources/Bootstrap.csproj"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []internal.Package{
		{
			PackageName: "Autofac",
			Version:     "",
			LineStart:   11,
			LineEnd:     11,
			Filepath:    manifestFile,
		},
		{
			PackageName: "Autofac.Extensions.DependencyInjection",
			Version:     "",
			LineStart:   12,
			LineEnd:     12,
			Filepath:    manifestFile,
		},
	}

	ValidatePackages(t, packages, expectedPackages)
}

func TestDotnetCsprojParser_Parse(t *testing.T) {
	parser := &DotnetCsprojParser{}
	manifestFile := "../test/resources/Gateway.csproj"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []internal.Package{
		{
			PackageName: "Lumo.AwsInfra",
			Version:     "4.0.1",
			LineStart:   16,
			LineEnd:     16,
			Filepath:    manifestFile,
		},
		{
			PackageName: "RestSharp",
			Version:     "106.15.0",
			LineStart:   17,
			LineEnd:     17,
			Filepath:    manifestFile,
		},
	}

	ValidatePackages(t, packages, expectedPackages)
}

func ValidatePackages(t *testing.T, packages []internal.Package, expectedPackages []internal.Package) {
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
