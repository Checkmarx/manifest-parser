package golang

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestGoModParser_Parse(t *testing.T) {
	parser := &GoModParser{}
	manifestFile := "../../../internal/testdata/test_go.mod"

	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file:  ", err)
	}

	expectedPackages := []models.Package{
		{
			PackageManager: "go",
			PackageName:    "github.com/Checkmarx/containers-resolver",
			Version:        "v1.0.9",
			FilePath:       manifestFile,
			LineStart:      6,
			LineEnd:        6,
			StartIndex:     8,
			EndIndex:       55,
		},
		{
			PackageManager: "go",
			PackageName:    "github.com/Checkmarx/gen-ai-prompts",
			Version:        "v0.0.0-20240807143411-708ceec12b63",
			FilePath:       manifestFile,
			LineStart:      7,
			LineEnd:        7,
			StartIndex:     8,
			EndIndex:       78,
		},
		{
			PackageManager: "go",
			PackageName:    "gotest.tools",
			Version:        "v2.2.0+incompatible",
			FilePath:       manifestFile,
			LineStart:      8,
			LineEnd:        8,
			StartIndex:     8,
			EndIndex:       40,
		},
		{
			PackageManager: "go",
			PackageName:    "dario.cat/mergo",
			Version:        "v1.0.1",
			FilePath:       manifestFile,
			LineStart:      12,
			LineEnd:        12,
			StartIndex:     8,
			EndIndex:       42,
		},
		{
			PackageManager: "go",
			PackageName:    "k8s.io/kube-openapi",
			Version:        "v0.0.0-20250318190949-c8a335a9a2ff",
			FilePath:       manifestFile,
			LineStart:      13,
			LineEnd:        13,
			StartIndex:     8,
			EndIndex:       74,
		},
		{
			PackageManager: "go",
			PackageName:    "sigs.k8s.io/yaml",
			Version:        "v1.4.0",
			FilePath:       manifestFile,
			LineStart:      14,
			LineEnd:        14,
			StartIndex:     8,
			EndIndex:       43,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
