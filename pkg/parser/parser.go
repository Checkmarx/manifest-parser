package parser

import (
	"ManifestParser/internal/parsers"
)

type Parser interface {
	Parse(manifestFile string) ([]parsers.Package, error)
}
