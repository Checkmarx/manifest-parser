package mod

import (
	"github.com/Checkmarx/manifest-parser/pkg/models"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal"
)

func TestGoModParser_Parse(t *testing.T) {
	parser := &GoModParser{}
	manifestFile := "../../../test/resources/test_go.mod"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file:  ", err)
	}

	expectedPackages := []models.Package{
		{
			PackageName: "github.com/gomarkdown/markdown",
			Version:     "v0.0.0-20241102151059-6bc1ffdc6e8c",
			LineStart:   6,
			LineEnd:     6,
			Filepath:    manifestFile,
		},
		{
			PackageName: "github.com/google/shlex",
			Version:     "v0.0.0-20191202100458-e7afc7fbc510",
			LineStart:   7,
			LineEnd:     7,
			Filepath:    manifestFile,
		},
		{
			PackageName: "dario.cat/mergo",
			Version:     "v1.0.1",
			LineStart:   11,
			LineEnd:     11,
			Filepath:    manifestFile,
		},
		{
			PackageName: "github.com/AdamKorcz/go-118-fuzz-build",
			Version:     "v0.0.0-20240914100643-eb91380d8434",
			LineStart:   12,
			LineEnd:     12,
			Filepath:    manifestFile,
		},
	}

	internal.ValidatePackages(t, packages, expectedPackages)
}
