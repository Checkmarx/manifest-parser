package golang

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"

	"golang.org/x/mod/modfile"
)

// GoModParser is a parser for Go modules.
type GoModParser struct{}

// Parse parses the Go module file and returns a list of packages.
func (p *GoModParser) Parse(manifest string) ([]models.Package, error) {
	cleanPath := filepath.Clean(manifest)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, err
	}
	mf, err := modfile.Parse(manifest, data, nil)
	if err != nil {
		return nil, err
	}

	// Split file into lines for position calculation
	lines := strings.Split(string(data), "\n")

	var packages []models.Package
	for _, req := range mf.Require {
		// Find the line where the dependency appears
		depName := req.Mod.Path
		depVersion := req.Mod.Version
		lineNum := req.Syntax.Start.Line // 1-based
		if lineNum <= 0 || lineNum > len(lines) {
			continue // skip if out of range
		}
		line := lines[lineNum-1]

		// Find the start index of the dependency name in the line
		startIdx := strings.Index(line, depName)
		if startIdx == -1 {
			startIdx = 0 // fallback
		}
		// End index: end of the line (like csproj_parser.go logic)
		endIdx := len(line)

		packages = append(packages, models.Package{
			PackageManager: "go",
			PackageName:    depName,
			Version:        depVersion,
			FilePath:       manifest,
			Locations: []models.Location{
				{
					Line:       lineNum,
					StartIndex: startIdx,
					EndIndex:   endIdx,
				},
			},
		})
	}
	return packages, nil
}
