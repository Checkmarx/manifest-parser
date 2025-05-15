package dotnet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestDotnetCsprojParser_ParseNoVersion(t *testing.T) {
	parser := &DotnetCsprojParser{}
	manifestFile := "../../../internal/testdata/ast-visual-studio-extension.csproj"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []models.Package{
		{
			PackageManager: "dotnet",
			PackageName:    "Community.VisualStudio.Toolkit.17",
			Version:        "17.0.507",
			LineStart:      31,
			LineEnd:        31,
			Filepath:       manifestFile,
			StartIndex:     5,
			EndIndex:       88,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "Community.VisualStudio.VSCT",
			Version:        "16.0.29.6",
			LineStart:      32,
			LineEnd:        32,
			Filepath:       manifestFile,
			StartIndex:     5,
			EndIndex:       83,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "Microsoft.TeamFoundationServer.Client",
			Version:        "19.225.1",
			LineStart:      33,
			LineEnd:        35,
			Filepath:       manifestFile,
			StartIndex:     5,
			EndIndex:       24,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "Microsoft.VisualStudio.SDK",
			Version:        "17.0.32112.339",
			LineStart:      36,
			LineEnd:        36,
			Filepath:       manifestFile,
			StartIndex:     5,
			EndIndex:       87,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "System.Json",
			Version:        "4.7.1",
			LineStart:      37,
			LineEnd:        37,
			Filepath:       manifestFile,
			StartIndex:     5,
			EndIndex:       63,
		},
	}

	internal.ValidatePackages(t, packages, expectedPackages)
}

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
					PackageManager: "dotnet",
					PackageName:    "Newtonsoft.Json",
					Version:        "13.0.1",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     5,
					EndIndex:       68,
				},
				{
					PackageManager: "dotnet",
					PackageName:    "Microsoft.Extensions.Logging",
					Version:        "6.0.0",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      5,
					LineEnd:        5,
					StartIndex:     5,
					EndIndex:       80,
				},
			},
			expectedError: false,
		},
		{
			name: "version ranges",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project Sdk="Microsoft.NET.Sdk">
  <ItemGroup>
    <PackageReference Include="Package1" Version="[1.2.3,2.0.0)" />
    <PackageReference Include="Package2" Version="[1.0.0,)" />
    <PackageReference Include="Package3" Version="*" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "dotnet",
					PackageName:    "Package1",
					Version:        "1.2.3",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     5,
					EndIndex:       68,
				},
				{
					PackageManager: "dotnet",
					PackageName:    "Package2",
					Version:        "1.0.0",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      5,
					LineEnd:        5,
					StartIndex:     5,
					EndIndex:       63,
				},
				{
					PackageManager: "dotnet",
					PackageName:    "Package3",
					Version:        "latest",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      6,
					LineEnd:        6,
					StartIndex:     5,
					EndIndex:       56,
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
					PackageManager: "dotnet",
					PackageName:    "Package1",
					Version:        "latest",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     5,
					EndIndex:       55,
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
					PackageManager: "dotnet",
					PackageName:    "Community.VisualStudio.VSCT",
					Version:        "16.0.29.6",
					Filepath:       filepath.Join(tempDir, "test.csproj"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     5,
					EndIndex:       83,
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
			internal.ValidatePackages(t, pkgs, tt.expectedPkgs)
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
		{"version range", "[1.2.3,2.0.0)", "1.2.3"},
		{"open range", "[1.0.0,)", "1.0.0"},
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
