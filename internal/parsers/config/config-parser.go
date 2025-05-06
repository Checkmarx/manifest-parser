package config

import (
	"github.com/Checkmarx/manifest-parser/internal"
)

type ConfigParser struct{}

func (p *ConfigParser) Parse(manifestFile string) ([]internal.Package, error) {
	return []internal.Package{
		{
			PackageManager: "Dotnet - config",
			PackageName:    "Hi, I'm a configuration file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
