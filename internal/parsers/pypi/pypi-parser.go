package pypi

import (
	"github.com/Checkmarx/manifest-parser/internal"
	"os"
	"strings"
)

type PypiParser struct{}

func (p *PypiParser) Parse(manifestFile string) ([]internal.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, err
	}

	// parse the content of the requirements.txt file to get the packages
	packages := make([]internal.Package, 0)
	lines := strings.Split(string(content), "\n")
	for l, lineContent := range lines {
		// split the lineContent by '=='
		parts := strings.Split(lineContent, "==")
		if len(parts) != 2 {
			//invalid package lineContent
			continue
		}
		packages = append(packages, internal.Package{
			PackageName: parts[0],
			Version:     parts[1],
			LineStart:   l + 1,
			LineEnd:     l + 1,
			Filepath:    manifestFile,
		})
	}
	return packages, nil
}
