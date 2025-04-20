package parser

import (
	"github.com/Checkmarx/manifest-parser/internal"
)

type Parser interface {
	Parse(manifestFile string) ([]internal.Package, error)
}
