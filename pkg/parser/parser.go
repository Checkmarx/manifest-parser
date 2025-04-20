package parser

import "ManifestParser/parsers"

type Parser interface {
	Parse(manifestFile string) ([]parsers.Package, error)
}
