package mod

import (
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type ModParser struct{}

func (p *ModParser) Parse(manifestFile string) ([]models.Package, error) {
	return []models.Package{
		{
			PackageManager: "go - go.mod",
			PackageName:    "Hi, I'm a Go module file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
