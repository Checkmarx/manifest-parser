package xml

import (
	"github.com/Checkmarx/manifest-parser/pkg/models"
)

type PackagesConfigParser struct{}

func (p *PackagesConfigParser) Parse(manifestFile string) ([]models.Package, error) {
	return []models.Package{
		{
			PackageManager: "java  Maven",
			PackageName:    "Hi, I'm a NuGet packages.config file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
