package dotnet

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestParseVersionProps(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"exact version", "1.2.3", "1.2.3"},
		{"open range", "[1.0.0,)", "latest"},
		{"wildcard", "*", "latest"},
		{"empty", "", "latest"},
		{"caret", "^1.2.3", "latest"},
		{"tilde", "~1.2.3", "latest"},
		{"greater than", ">1.2.3", "latest"},
		{"less than", "<2.0.0", "latest"},
		{"complex range", "1.2.30.1220", "1.2.30.1220"},
		{"complex range with parentheses", "(1.2.3,2.0.0]", "latest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionProps(tt.version)
			if result != tt.expected {
				t.Errorf("parseVersionProps(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestComputePackageVersionLocations(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		packageName string
		wantStart   int
		wantEnd     int
	}{
		{
			name:        "single-line format",
			content:     `<PackageVersion Include="Package1" Version="1.0.0" />`,
			packageName: "Package1",
			wantStart:   0,
			wantEnd:     len(`<PackageVersion Include="Package1" Version="1.0.0" />`),
		},
		{
			name:        "multi-line format",
			content:     `<PackageVersion Include="Package1">\n  <Version>1.0.0</Version>\n</PackageVersion>`,
			packageName: "Package1",
			wantStart:   0,
			wantEnd:     len(`<PackageVersion Include="Package1">`),
		},
		{
			name:        "package not found",
			content:     `<PackageVersion Include="Package1" Version="1.0.0" />`,
			packageName: "Package2",
			wantStart:   0,
			wantEnd:     0,
		},
		{
			name:        "empty content",
			content:     "",
			packageName: "Package1",
			wantStart:   0,
			wantEnd:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.content, "\n")
			lineNum := -1
			packagePattern := fmt.Sprintf(`PackageVersion.*Include="%s"`, tt.packageName)
			re := regexp.MustCompile(packagePattern)

			for i, line := range lines {
				if re.MatchString(line) {
					lineNum = i
					break
				}
			}

			locations := computePackageVersionLocations(lines, lineNum)
			if len(locations) == 0 {
				if tt.wantStart != 0 || tt.wantEnd != 0 {
					t.Errorf("computePackageVersionLocations() returned no locations, want (%v, %v)",
						tt.wantStart, tt.wantEnd)
				}
				return
			}

			firstLocation := locations[0]
			if firstLocation.StartIndex != tt.wantStart || firstLocation.EndIndex != tt.wantEnd {
				t.Errorf("computePackageVersionLocations() = (%v, %v), want (%v, %v)",
					firstLocation.StartIndex, firstLocation.EndIndex, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestDotnetDirectoryPackagesPropsParser_Parse(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "directory-packages-props-test")
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
			name: "single-line format",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project>
  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
  </PropertyGroup>
  <ItemGroup>
    <PackageVersion Include="Package1" Version="1.0.0" />
    <PackageVersion Include="Package2" Version="2.0.0" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "1.0.0",
					FilePath:       filepath.Join(tempDir, "Directory.Packages.props"),
					Locations: []models.Location{
						{
							Line:       6,
							StartIndex: 4,
							EndIndex:   57,
						},
					},
				},
				{
					PackageManager: "nuget",
					PackageName:    "Package2",
					Version:        "2.0.0",
					FilePath:       filepath.Join(tempDir, "Directory.Packages.props"),
					Locations: []models.Location{
						{
							Line:       7,
							StartIndex: 4,
							EndIndex:   57,
						},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "multi-line format",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project>
  <PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
  </PropertyGroup>
  <ItemGroup>
    <PackageVersion Include="Package1">
      <Version>1.0.0</Version>
    </PackageVersion>
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "nuget",
					PackageName:    "Package1",
					Version:        "1.0.0",
					FilePath:       filepath.Join(tempDir, "Directory.Packages.props"),
					Locations: []models.Location{
						{
							Line:       6,
							StartIndex: 4,
							EndIndex:   39,
						},
						{
							Line:       7,
							StartIndex: 6,
							EndIndex:   30,
						},
						{
							Line:       8,
							StartIndex: 4,
							EndIndex:   21,
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
			testFile := filepath.Join(tempDir, "Directory.Packages.props")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Create parser and parse file
			parser := &DotnetDirectoryPackagesPropsParser{}
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

func TestDotnetDirectoryPackagesPropsParser_ParseActualFile(t *testing.T) {
	parser := &DotnetDirectoryPackagesPropsParser{}
	manifestFile := "../../../internal/testdata/Directory.Packages.props"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expectedPackages := []models.Package{
		{
			PackageManager: "nuget",
			PackageName:    "AwesomeAssertions",
			Version:        "8.1.0",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{Line: 14, StartIndex: 4, EndIndex: 66},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "ILMerge",
			Version:        "3.0.41.22",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{Line: 15, StartIndex: 4, EndIndex: 60},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "MSTest.TestAdapter",
			Version:        "latest",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{Line: 16, StartIndex: 4, EndIndex: 85},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "MSTest.TestFramework",
			Version:        "latest",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{Line: 17, StartIndex: 4, EndIndex: 64},
			},
		},
		{
			PackageManager: "nuget",
			PackageName:    "System.Text.Json",
			Version:        "latest",
			FilePath:       manifestFile,
			Locations: []models.Location{
				{Line: 19, StartIndex: 4, EndIndex: 73},
			},
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
