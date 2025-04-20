package csproj

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/internal"
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

	internal.ValidatePackages(t, packages, expectedPackages)
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

	internal.ValidatePackages(t, packages, expectedPackages)
}
