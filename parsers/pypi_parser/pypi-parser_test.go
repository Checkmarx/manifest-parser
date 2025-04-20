package pypi_parser

import (
	"ManifestParser/parsers"
	"ManifestParser/parsers/dotnet_parser"
	"testing"
)

func TestPypiParser(t *testing.T) {
	parser := &PypiParser{}
	manifestFile := "../test/resources/requirements.txt"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []parsers.Package{
		{
			PackageName: "awacs",
			Version:     "2.3.0",
			LineStart:   1,
			LineEnd:     1,
			Filepath:    manifestFile,
		},
		{
			PackageName: "awscli",
			Version:     "1.32.70",
			LineStart:   2,
			LineEnd:     2,
			Filepath:    manifestFile,
		},
	}

	dotnet_parser.validatePackages(t, packages, expectedPackages)
}
