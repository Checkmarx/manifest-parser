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
			PackageManager: "nuget",
			PackageName:    "Microsoft.CodeDom.Providers.DotNetCompilerPlatform",
			Version:        "1.0.0",
			LineStart:      2,
			LineEnd:        2,
			StartIndex:     2,
			EndIndex:       109,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.Net.Compilers",
			Version:        "1.0.0",
			LineStart:      3,
			LineEnd:        3,
			StartIndex:     2,
			EndIndex:       111,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.Web.Infrastructure",
			Version:        "1.0.0.0",
			LineStart:      4,
			LineEnd:        4,
			StartIndex:     2,
			EndIndex:       89,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.Web.Xdt",
			Version:        "2.1.1",
			LineStart:      5,
			LineEnd:        5,
			StartIndex:     2,
			EndIndex:       76,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "Newtonsoft.Json",
			Version:        "8.0.3",
			LineStart:      6,
			LineEnd:        6,
			StartIndex:     2,
			EndIndex:       99,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "NuGet.Core",
			Version:        "2.11.1",
			LineStart:      7,
			LineEnd:        7,
			StartIndex:     2,
			EndIndex:       70,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "NuGet.Server",
			Version:        "2.11.2",
			LineStart:      8,
			LineEnd:        8,
			StartIndex:     2,
			EndIndex:       72,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "RouteMagic",
			Version:        "1.3",
			LineStart:      9,
			LineEnd:        9,
			StartIndex:     2,
			EndIndex:       67,
			FilePath:       manifestFile,
		},
		{
			PackageManager: "nuget",
			PackageName:    "WebActivatorEx",
			Version:        "2.1.0",
			LineStart:      10,
			LineEnd:        10,
			StartIndex:     2,
			EndIndex:       73,
			FilePath:       manifestFile,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
