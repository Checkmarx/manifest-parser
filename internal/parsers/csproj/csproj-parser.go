package csproj

import (
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type CsprojParser struct{}

func (p *CsprojParser) Parse(manifestFile string) ([]models.Package, error) {
	return []models.Package{
		{
			PackageManager: "Dotnet - csproj",
			PackageName:    "Hi, I'm a .NET file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
