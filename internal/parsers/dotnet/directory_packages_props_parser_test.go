package dotnet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestDotnetDirectoryPackagesPropsParser_Parse(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "props-test")
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
<Project>
  <ItemGroup>
    <PackageVersion Include="Autofac" Version="8.1.0" />
    <PackageVersion Include="coverlet.collector" Version="6.0.2" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "dotnet",
					PackageName:    "Autofac",
					Version:        "8.1.0",
					Filepath:       filepath.Join(tempDir, "test.props"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     1,
					EndIndex:       1,
				},
				{
					PackageManager: "dotnet",
					PackageName:    "coverlet.collector",
					Version:        "6.0.2",
					Filepath:       filepath.Join(tempDir, "test.props"),
					LineStart:      5,
					LineEnd:        5,
					StartIndex:     1,
					EndIndex:       1,
				},
			},
			expectedError: false,
		},
		{
			name: "missing version",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project>
  <ItemGroup>
    <PackageVersion Include="Autofac" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "dotnet",
					PackageName:    "Autofac",
					Version:        "latest",
					Filepath:       filepath.Join(tempDir, "test.props"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     1,
					EndIndex:       1,
				},
			},
			expectedError: false,
		},
		{
			name: "special version specifiers",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project>
  <ItemGroup>
    <PackageVersion Include="Autofac" Version=">8.0.0" />
    <PackageVersion Include="coverlet.collector" Version="[6.0.0,7.0.0)" />
  </ItemGroup>
</Project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "dotnet",
					PackageName:    "Autofac",
					Version:        "latest",
					Filepath:       filepath.Join(tempDir, "test.props"),
					LineStart:      4,
					LineEnd:        4,
					StartIndex:     1,
					EndIndex:       1,
				},
				{
					PackageManager: "dotnet",
					PackageName:    "coverlet.collector",
					Version:        "latest",
					Filepath:       filepath.Join(tempDir, "test.props"),
					LineStart:      5,
					LineEnd:        5,
					StartIndex:     1,
					EndIndex:       1,
				},
			},
			expectedError: false,
		},
		{
			name: "invalid XML",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project>
  <ItemGroup>
    <PackageVersion Include="Autofac" Version="8.1.0"
  </ItemGroup>
</Project>`,
			expectedPkgs:  nil,
			expectedError: true,
		},
		{
			name:          "empty file",
			content:       ``,
			expectedPkgs:  nil,
			expectedError: true,
		},
		{
			name: "no package references",
			content: `<?xml version="1.0" encoding="utf-8"?>
<Project>
  <ItemGroup>
  </ItemGroup>
</Project>`,
			expectedPkgs:  []models.Package{},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tempDir, "test.props")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			parser := &DotnetDirectoryPackagesPropsParser{}
			pkgs, err := parser.Parse(testFile)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			internal.ValidatePackages(t, pkgs, tt.expectedPkgs)
		})
	}
}
