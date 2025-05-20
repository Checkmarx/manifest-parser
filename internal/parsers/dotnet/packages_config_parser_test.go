package dotnet

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
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
			FilePath:    manifestFile,
		},
		{
			PackageName: "Microsoft.Net.Compilers",
			Version:     "1.0.0",
			LineStart:   4,
			LineEnd:     4,
			StartIndex:  3,
			EndIndex:    112,
			FilePath:    manifestFile,
		},
		{
			PackageName: "Microsoft.Web.Infrastructure",
			Version:     "1.0.0.0",
			LineStart:   5,
			LineEnd:     5,
			StartIndex:  3,
			EndIndex:    90,
			FilePath:    manifestFile,
		},
		{
			PackageName: "Microsoft.Web.Xdt",
			Version:     "2.1.1",
			LineStart:   6,
			LineEnd:     6,
			StartIndex:  3,
			EndIndex:    77,
			FilePath:    manifestFile,
		},
		{
			PackageName: "Newtonsoft.Json",
			Version:     "8.0.3",
			LineStart:   7,
			LineEnd:     7,
			StartIndex:  3,
			EndIndex:    100,
			FilePath:    manifestFile,
		},
		{
			PackageName: "NuGet.Core",
			Version:     "2.11.1",
			LineStart:   8,
			LineEnd:     8,
			StartIndex:  3,
			EndIndex:    71,
			FilePath:    manifestFile,
		},
		{
			PackageName: "NuGet.Server",
			Version:     "2.11.2",
			LineStart:   9,
			LineEnd:     9,
			StartIndex:  3,
			EndIndex:    73,
			FilePath:    manifestFile,
		},
		{
			PackageName: "RouteMagic",
			Version:     "1.3",
			LineStart:   10,
			LineEnd:     10,
			StartIndex:  3,
			EndIndex:    68,
			FilePath:    manifestFile,
		},
		{
			PackageName: "WebActivatorEx",
			Version:     "2.1.0",
			LineStart:   11,
			LineEnd:     11,
			StartIndex:  3,
			EndIndex:    74,
			FilePath:    manifestFile,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
