package dotnet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
	"github.com/stretchr/testify/assert"
)

func TestPackagesConfigParser(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedPkgs  []models.Package
		expectedError bool
		errorMessage  string
	}{
		{
			name: "valid single-line packages",
			content: `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Package1" version="1.0.0" />
  <package id="Package2" version="2.0.0" />
</packages>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "1.0.0",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       2,
							StartIndex: 2,
							EndIndex:   43,
						},
					},
				},
				{
					PackageManager: "nuget",
					PackageName:    "Package2",
					Version:        "2.0.0",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 2,
							EndIndex:   43,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "valid multi-line packages",
			content: `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Package1">
    <version>1.0.0</version>
  </package>
  <package id="Package2">
    <version>2.0.0</version>
  </package>
</packages>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "1.0.0",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       2,
							StartIndex: 2,
							EndIndex:   25,
						},
						{
							Line:       3,
							StartIndex: 4,
							EndIndex:   28,
						},
						{
							Line:       4,
							StartIndex: 2,
							EndIndex:   12,
						},
					},
				},
				{
					PackageManager: "nuget",
					PackageName:    "Package2",
					Version:        "2.0.0",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       5,
							StartIndex: 2,
							EndIndex:   25,
						},
						{
							Line:       6,
							StartIndex: 4,
							EndIndex:   28,
						},
						{
							Line:       7,
							StartIndex: 2,
							EndIndex:   12,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "valid packages with version ranges",
			content: `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Package1" version="[1.0.0,2.0.0)" />
  <package id="Package2" version="~1.0.0" />
</packages>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "latest",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       2,
							StartIndex: 2,
							EndIndex:   51,
						},
					},
				},
				{
					PackageManager: "nuget",
					PackageName:    "Package2",
					Version:        "latest",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 2,
							EndIndex:   44,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "valid packages with empty version",
			content: `<?xml version="1.0" encoding="utf-8"?>
<packages>
  <package id="Package1" version="" />
  <package id="Package2" />
</packages>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "latest",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       2,
							StartIndex: 2,
							EndIndex:   38,
						},
					},
				},
				{
					PackageManager: "nuget",
					PackageName:    "Package2",
					Version:        "latest",
					FilePath:       "", // Will be set to manifestPath
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 2,
							EndIndex:   27,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "invalid XML",
			content: `<?xml version="1.0" encoding="utf-8"?>
		<packages>
		  <package id="Package1" version="1.0.0"
		</packages>`,
			expectedPkgs:  nil,
			expectedError: true,
			errorMessage:  "failed to parse XML",
		},
		{
			name:          "empty file",
			content:       "",
			expectedPkgs:  nil,
			expectedError: true,
			errorMessage:  "empty file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpDir := t.TempDir()
			manifestPath := filepath.Join(tmpDir, "packages.config")
			err := os.WriteFile(manifestPath, []byte(tt.content), 0644)
			assert.NoError(t, err)

			// Update expected file paths
			for i := range tt.expectedPkgs {
				tt.expectedPkgs[i].FilePath = manifestPath
			}

			parser := &DotnetPackagesConfigParser{}
			pkgs, err := parser.Parse(manifestPath)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
				assert.Nil(t, pkgs)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPkgs, pkgs)
			}
		})
	}
}

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
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 2, StartIndex: 2, EndIndex: 109}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.Net.Compilers",
			Version:        "1.0.0",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 3, StartIndex: 2, EndIndex: 111}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.Web.Infrastructure",
			Version:        "1.0.0.0",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 4, StartIndex: 2, EndIndex: 89}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.Web.Xdt",
			Version:        "2.1.1",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 5, StartIndex: 2, EndIndex: 76}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Newtonsoft.Json",
			Version:        "8.0.3",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 6, StartIndex: 2, EndIndex: 99}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "NuGet.Core",
			Version:        "2.11.1",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 7, StartIndex: 2, EndIndex: 70}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "NuGet.Server",
			Version:        "2.11.2",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 8, StartIndex: 2, EndIndex: 72}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "RouteMagic",
			Version:        "1.3",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 9, StartIndex: 2, EndIndex: 67}},
		},
		{
			PackageManager: "nuget",
			PackageName:    "WebActivatorEx",
			Version:        "2.1.0",
			FilePath:       manifestFile,
			Locations:      []models.Location{{Line: 10, StartIndex: 2, EndIndex: 73}},
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
