package package_json_parser

import (
	"ManifestParser/parsers"
	"testing"
)

func TestNpmPackageJsonParser_ParseParser_Parse(t *testing.T) {
	parser := &NpmPackageJsonParser{}
	manifestFile := "../test/resources/package.json"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []parsers.Package{
		{
			PackageName: "@ant-design/icons",
			Version:     "^2.1.1",
			LineStart:   23,
			LineEnd:     23,
			Filepath:    manifestFile,
		},
		{
			PackageName: "@babel/cli",
			Version:     "^7.12.1",
			LineStart:   26,
			LineEnd:     26,
			Filepath:    manifestFile,
		},
		{
			PackageName: "@babel/core",
			Version:     "^7.19.6",
			LineStart:   27,
			LineEnd:     27,
			Filepath:    manifestFile,
		},
	}

	dotnet_parser.validatePackages(t, packages, expectedPackages)
}
