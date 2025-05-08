package parser

import (
	"github.com/Checkmarx/manifest-parser/pkg/models"
)

type Parser interface {
	Parse(manifestFile string) ([]models.Package, error)
}
