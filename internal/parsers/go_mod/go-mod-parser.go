package go_mod

import (
	"log"
	"os"

	"golang.org/x/mod/modfile"

	"github.com/Checkmarx/manifest-parser/internal"
)

// GoModParser is a parser for Go modules.
type GoModParser struct{}

// Parse parses the Go module file and returns a list of packages.
func (p *GoModParser) Parse(manifest string) ([]internal.Package, error) {
	data, err := os.ReadFile(manifest)
	if err != nil {
		log.Fatal(err)
	}
	mf, err := modfile.Parse(manifest, data, nil)
	if err != nil {
		log.Fatal(err)
	}
	var packages []internal.Package
	for _, req := range mf.Require {
		packages = append(packages, internal.Package{
			PackageName: req.Mod.Path,
			Version:     req.Mod.Version,
			LineStart:   req.Syntax.Start.Line,
			LineEnd:     req.Syntax.End.Line,
			Filepath:    manifest,
		})
	}
	return packages, nil
}
