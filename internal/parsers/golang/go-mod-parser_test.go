package golang

import (
	"os"
	"strings"
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
			StartIndex:     9,
			EndIndex:       56,
		},
		{
			PackageManager: "go",
			PackageName:    "github.com/Checkmarx/gen-ai-prompts",
			Version:        "v0.0.0-20240807143411-708ceec12b63",
			FilePath:       manifestFile,
			LineStart:      7,
			LineEnd:        7,
			StartIndex:     9,
			EndIndex:       79,
		},
		{
			PackageManager: "go",
			PackageName:    "gotest.tools",
			Version:        "v2.2.0+incompatible",
			FilePath:       manifestFile,
			LineStart:      8,
			LineEnd:        8,
			StartIndex:     9,
			EndIndex:       41,
		},
		{
			PackageManager: "go",
			PackageName:    "dario.cat/mergo",
			Version:        "v1.0.1",
			FilePath:       manifestFile,
			LineStart:      12,
			LineEnd:        12,
			StartIndex:     9,
			EndIndex:       43,
		},
		{
			PackageManager: "go",
			PackageName:    "k8s.io/kube-openapi",
			Version:        "v0.0.0-20250318190949-c8a335a9a2ff",
			FilePath:       manifestFile,
			LineStart:      13,
			LineEnd:        13,
			StartIndex:     9,
			EndIndex:       75,
		},
		{
			PackageManager: "go",
			PackageName:    "sigs.k8s.io/yaml",
			Version:        "v1.4.0",
			FilePath:       manifestFile,
			LineStart:      14,
			LineEnd:        14,
			StartIndex:     9,
			EndIndex:       44,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}

func TestGoModParser_Parse_WithPositions(t *testing.T) {
	parser := &GoModParser{}
	manifestFile := "../../../test/resources/test_go.mod"
	data, err := os.ReadFile(manifestFile)
	if err != nil {
		t.Fatalf("Failed to read test go.mod: %v", err)
	}
	lines := strings.Split(string(data), "\n")

	expectedPackages := []models.Package{
		{
			PackageManager: "go",
			PackageName:    "github.com/gomarkdown/markdown",
			Version:        "v0.0.0-20241102151059-6bc1ffdc6e8c",
			FilePath:       manifestFile,
			LineStart:      6,
			LineEnd:        6,
			StartIndex:     strings.Index(lines[5], "github.com/gomarkdown/markdown") + 1,
			EndIndex:       len(lines[5]) + 1,
		},
		{
			PackageManager: "go",
			PackageName:    "github.com/google/shlex",
			Version:        "v0.0.0-20191202100458-e7afc7fbc510",
			FilePath:       manifestFile,
			LineStart:      7,
			LineEnd:        7,
			StartIndex:     strings.Index(lines[6], "github.com/google/shlex") + 1,
			EndIndex:       len(lines[6]) + 1,
		},
		{
			PackageManager: "go",
			PackageName:    "dario.cat/mergo",
			Version:        "v1.0.1",
			FilePath:       manifestFile,
			LineStart:      11,
			LineEnd:        11,
			StartIndex:     strings.Index(lines[10], "dario.cat/mergo") + 1,
			EndIndex:       len(lines[10]) + 1,
		},
		{
			PackageManager: "go",
			PackageName:    "github.com/AdamKorcz/go-118-fuzz-build",
			Version:        "v0.0.0-20240914100643-eb91380d8434",
			FilePath:       manifestFile,
			LineStart:      12,
			LineEnd:        12,
			StartIndex:     strings.Index(lines[11], "github.com/AdamKorcz/go-118-fuzz-build") + 1,
			EndIndex:       len(lines[11]) + 1,
		},
	}

	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	testdata.ValidatePackages(t, packages, expectedPackages)
}
