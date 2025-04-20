package parser

import (
	"ManifestParser/internal"
)

type Parser interface {
	Parse(manifestFile string) ([]internal.Package, error)
}
