package dotnet

import (
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

func TestFindPackageVersionPosition(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		packageName string
		wantStart   int
		wantEnd     int
	}{
		{
			name:        "simple package",
			content:     `<PackageVersion Include="Package1" Version="1.0.0" />`,
			packageName: "Package1",
			wantStart:   0,
			wantEnd:     len(`<PackageVersion Include="Package1" Version="1.0.0" />`),
		},
		{
			name:        "package with special characters",
			content:     `<PackageVersion Include="Package.1.2" Version="1.0.0" />`,
			packageName: "Package.1.2",
			wantStart:   0,
			wantEnd:     len(`<PackageVersion Include="Package.1.2" Version="1.0.0" />`),
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
			_, start, end := findPackageVersionPosition(tt.content, tt.packageName)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("findPackageVersionPosition() = (%v, %v), want (%v, %v)",
					start, end, tt.wantStart, tt.wantEnd)
			}
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
			PackageManager: "dotnet",
			PackageName:    "AwesomeAssertions",
			Version:        "8.1.0",
			FilePath:       manifestFile,
			LineStart:      14,
			LineEnd:        14,
			StartIndex:     4,
			EndIndex:       66,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "ILMerge",
			Version:        "3.0.41.22",
			FilePath:       manifestFile,
			LineStart:      15,
			LineEnd:        15,
			StartIndex:     4,
			EndIndex:       60,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "MSTest.TestAdapter",
			Version:        "latest",
			FilePath:       manifestFile,
			LineStart:      16,
			LineEnd:        16,
			StartIndex:     4,
			EndIndex:       85,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "MSTest.TestFramework",
			Version:        "latest",
			FilePath:       manifestFile,
			LineStart:      17,
			LineEnd:        17,
			StartIndex:     4,
			EndIndex:       64,
		},
		{
			PackageManager: "dotnet",
			PackageName:    "System.Text.Json",
			Version:        "latest",
			FilePath:       manifestFile,
			LineStart:      19,
			LineEnd:        19,
			StartIndex:     4,
			EndIndex:       73,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
