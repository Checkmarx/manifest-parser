package mod

import (
	"github.com/Checkmarx/manifest-parser/pkg/models"
	"os"
	"path/filepath"

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
	var packages []models.Package
	for _, req := range mf.Require {
		packages = append(packages, models.Package{
			PackageName: req.Mod.Path,
			Version:     req.Mod.Version,
			LineStart:   req.Syntax.Start.Line,
			LineEnd:     req.Syntax.End.Line,
			Filepath:    manifest,
		})
	}
	return packages, nil
}
