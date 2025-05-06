package xml

import (
	"github.com/Checkmarx/manifest-parser/internal"
)

type PackagesConfigParser struct{}

func (p *PackagesConfigParser) Parse(manifestFile string) ([]internal.Package, error) {
	return []internal.Package{
		{
			PackageManager: "Dotnet - nuget",
			PackageName:    "Hi, I'm a NuGet packages.config file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
