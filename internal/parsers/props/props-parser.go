package props

import (
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type PropsParser struct{}

func (p *PropsParser) Parse(manifestFile string) ([]models.Package, error) {
	return []models.Package{
		{
			PackageManager: "Maven - pom.xml",
			PackageName:    "Hi, I'm a Maven properties file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
