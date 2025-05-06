package mod

import (
	"github.com/Checkmarx/manifest-parser/internal"
)

type ModParser struct{}

func (p *ModParser) Parse(manifestFile string) ([]internal.Package, error) {
	return []internal.Package{
		{
			PackageManager: "go - go.mod",
			PackageName:    "Hi, I'm a Go module file",
			Version:        "POC",
			Filepath:       manifestFile,
		},
	}, nil
}
