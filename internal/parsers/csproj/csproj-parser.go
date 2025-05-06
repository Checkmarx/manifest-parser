package csproj

import (
	"github.com/Checkmarx/manifest-parser/internal"
)

type CsprojParser struct{}

func (p *CsprojParser) Parse(manifestFile string) ([]internal.Package, error) {
	return []internal.Package{
		{
			PackageManager: "Dotnet - csproj",
			PackageName:    "Hi, I'm a .NET file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
