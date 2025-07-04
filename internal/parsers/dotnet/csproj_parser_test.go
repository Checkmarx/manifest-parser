package dotnet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestDotnetCsprojParser_Parse(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "csproj-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		content       string
		expectedPkgs  []models.Package
		expectedError bool
	}{
		{
			name: "exact versions",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Newtonsoft.Json" Version="13.0.1" />
    <PackageReference Include="Microsoft.Extensions.Logging" Version="6.0.0" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Newtonsoft.Json",
					Version:        "13.0.1",
					FilePath:       filepath.Join(tempDir, "test.csproj"),
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 4,
							EndIndex:   67,
						},
					},
				},
				{
					PackageManager: "nuget",
					PackageName:    "Microsoft.Extensions.Logging",
					Version:        "6.0.0",
					FilePath:       filepath.Join(tempDir, "test.csproj"),
					Locations: []models.Location{
						{
							Line:       4,
							StartIndex: 4,
							EndIndex:   79,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "multi-line package reference",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Microsoft.TeamFoundationServer.Client">
      <Version>19.225.1</Version>
    </PackageReference>
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Microsoft.TeamFoundationServer.Client",
					Version:        "19.225.1",
					FilePath:       filepath.Join(tempDir, "test.csproj"),
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 4,
							EndIndex:   70,
						},
						{
							Line:       4,
							StartIndex: 6,
							EndIndex:   33,
						},
						{
							Line:       5,
							StartIndex: 4,
							EndIndex:   23,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "empty version",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Package1" Version="" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "latest",
					FilePath:       filepath.Join(tempDir, "test.csproj"),
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 4,
							EndIndex:   54,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "invalid XML",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Package1" Version="1.0.0"
  </ItemGroup>
</Project>`,
			expectedPkgs:  nil,
			expectedError: true,
		},
		{
			name: "no package references",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
  </ItemGroup>
</Project>`,
			expectedPkgs:  []models.Package{},
			expectedError: false,
		},
		{
			name: "four-part version",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Community.VisualStudio.VSCT" Version="16.0.29.6" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Community.VisualStudio.VSCT",
					Version:        "16.0.29.6",
					FilePath:       filepath.Join(tempDir, "test.csproj"),
					Locations: []models.Location{
						{
							Line:       3,
							StartIndex: 4,
							EndIndex:   82,
						},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, "test.csproj")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Create parser and parse file
			parser := &DotnetCsprojParser{}
			pkgs, err := parser.Parse(testFile)

			// Check error
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check packages
			testdata.ValidatePackages(t, pkgs, tt.expectedPkgs)
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"exact version", "1.2.3", "1.2.3"},
		{"version range", "[1.2.3,2.0.0)", "latest"},
		{"open range", "1.0.0.345", "1.0.0.345"},
		{"wildcard", "*", "latest"},
		{"empty", "", "latest"},
		{"caret", "^1.2.3", "latest"},
		{"tilde", "~1.2.3", "latest"},
		{"greater than", ">1.2.3", "latest"},
		{"less than", "<2.0.0", "latest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.version)
			if result != tt.expected {
				t.Errorf("parseVersion(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestDotnetCsprojParser_ParseNoVersion(t *testing.T) {
	parser := &DotnetCsprojParser{}
	manifestFile := "../../../internal/testdata/ast-visual-studio-extension.csproj"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []models.Package{
		{
			PackageManager: "nuget",
			PackageName:    "Community.VisualStudio.Toolkit.17",
			Version:        "17.0.507",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{
					Line:       30,
					StartIndex: 4,
					EndIndex:   87,
				},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Community.VisualStudio.VSCT",
			Version:        "16.0.29.6",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{
					Line:       31,
					StartIndex: 4,
					EndIndex:   82,
				},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.TeamFoundationServer.Client",
			Version:        "19.225.1",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{Line: 32, StartIndex: 4, EndIndex: 70},
				{Line: 33, StartIndex: 6, EndIndex: 33},
				{Line: 34, StartIndex: 4, EndIndex: 23},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "Microsoft.VisualStudio.SDK",
			Version:        "17.0.32112.339",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{
					Line:       35,
					StartIndex: 4,
					EndIndex:   86,
				},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "System.Json",
			Version:        "4.7.1",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{
					Line:       36,
					StartIndex: 4,
					EndIndex:   62,
				},
			},
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
