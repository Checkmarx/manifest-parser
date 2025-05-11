package config

import (
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type ConfigParser struct{}

func (p *ConfigParser) Parse(manifestFile string) ([]models.Package, error) {
	return []models.Package{
		{
			PackageManager: "Dotnet - config",
			PackageName:    "Hi, I'm a configuration file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
