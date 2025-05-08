package props

import (
	"github.com/Checkmarx/manifest-parser/pkg/models"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal"
)

func TestDotnetDirectoryPackagesPropsParser_Parse(t *testing.T) {
	parser := &DotnetDirectoryPackagesPropsParser{}
	manifestFile := "../../../test/resources/Directory.Packages.props"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []models.Package{
		{
			PackageName: "Autofac",
			Version:     "8.1.0",
			LineStart:   7,
			LineEnd:     7,
			Filepath:    manifestFile,
		},
		{
			PackageName: "Autofac.Extensions.DependencyInjection",
			Version:     "10.0.0",
			LineStart:   8,
			LineEnd:     8,
			Filepath:    manifestFile,
		},
		{
			PackageName: "coverlet.collector",
			Version:     "6.0.2",
			LineStart:   9,
			LineEnd:     9,
			Filepath:    manifestFile,
		},
	}

	internal.ValidatePackages(t, packages, expectedPackages)
}
