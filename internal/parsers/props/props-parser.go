package props

import (
	"github.com/Checkmarx/manifest-parser/internal"
)

type PropsParser struct{}

func (p *PropsParser) Parse(manifestFile string) ([]internal.Package, error) {
	return []internal.Package{
		{
			PackageManager: "Maven - pom.xml",
			PackageName:    "Hi, I'm a Maven properties file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
