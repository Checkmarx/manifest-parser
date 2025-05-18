package dotnet

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"

	"github.com/Checkmarx/manifest-parser/internal"
)

func TestDotnetPackagesConfigParser_ParseRealFile_Actual(t *testing.T) {
	parser := &DotnetPackagesConfigParser{}
	manifestFile := "../../../internal/testdata/packages.config"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expectedPackages := []models.Package{
		{
			PackageName: "Microsoft.CodeDom.Providers.DotNetCompilerPlatform",
			Version:     "1.0.0",
			LineStart:   3,
			LineEnd:     3,
			StartIndex:  3,
			EndIndex:    110,
			Filepath:    manifestFile,
		},
		{
			PackageName: "Microsoft.Net.Compilers",
			Version:     "1.0.0",
			LineStart:   4,
			LineEnd:     4,
			StartIndex:  3,
			EndIndex:    112,
			Filepath:    manifestFile,
		},
		{
			PackageName: "Microsoft.Web.Infrastructure",
			Version:     "1.0.0.0",
			LineStart:   5,
			LineEnd:     5,
			StartIndex:  3,
			EndIndex:    90,
			Filepath:    manifestFile,
		},
		{
			PackageName: "Microsoft.Web.Xdt",
			Version:     "2.1.1",
			LineStart:   6,
			LineEnd:     6,
			StartIndex:  3,
			EndIndex:    77,
			Filepath:    manifestFile,
		},
		{
			PackageName: "Newtonsoft.Json",
			Version:     "8.0.3",
			LineStart:   7,
			LineEnd:     7,
			StartIndex:  3,
			EndIndex:    100,
			Filepath:    manifestFile,
		},
		{
			PackageName: "NuGet.Core",
			Version:     "2.11.1",
			LineStart:   8,
			LineEnd:     8,
			StartIndex:  3,
			EndIndex:    71,
			Filepath:    manifestFile,
		},
		{
			PackageName: "NuGet.Server",
			Version:     "2.11.2",
			LineStart:   9,
			LineEnd:     9,
			StartIndex:  3,
			EndIndex:    73,
			Filepath:    manifestFile,
		},
		{
			PackageName: "RouteMagic",
			Version:     "1.3",
			LineStart:   10,
			LineEnd:     10,
			StartIndex:  3,
			EndIndex:    68,
			Filepath:    manifestFile,
		},
		{
			PackageName: "WebActivatorEx",
			Version:     "2.1.0",
			LineStart:   11,
			LineEnd:     11,
			StartIndex:  3,
			EndIndex:    74,
			Filepath:    manifestFile,
		},
	}

	internal.ValidatePackages(t, packages, expectedPackages)
}
