package packages_config

import (
	"github.com/Checkmarx/manifest-parser/internal"
	"github.com/Checkmarx/manifest-parser/parsers/dotnet_parser/csproj_parser"
	"testing"
)

func TestDotnetPackagesConfigParser_Parse(t *testing.T) {

	// Initialize parser
	parser := &DotnetPackagesConfigParser{}

	manifestFile := "../test/resources/packages.config"

	// Parse the test file
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify results
	expected := []internal.Package{
		{
			PackageName: "Newtonsoft.Json",
			Version:     "13.0.1",
			LineStart:   3,
			LineEnd:     3,
			Filepath:    manifestFile,
		},
		{
			PackageName: "System.Runtime",
			Version:     "4.3.0",
			LineStart:   4,
			LineEnd:     4,
			Filepath:    manifestFile,
		},
	}

	csproj_parser.validatePackages(t, packages, expected)
}
