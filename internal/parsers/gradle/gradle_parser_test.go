package gradle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestGradleParser_Parse(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedPkgs  []models.Package
		expectedError bool
	}{
		{
			name: "basic gradle file",
			content: `plugins {
    id 'java'
}

ext {
    springVersion = '5.3.0'
}

dependencies {
    implementation 'org.springframework:spring-core:5.3.0'
    testImplementation 'junit:junit:4.13'
    api 'com.google.guava:guava:30.1-jre'
    implementation group: 'org.apache.commons', name: 'commons-lang3', version: '3.12.0'
}

buildscript {
    dependencies {
        classpath 'com.android.tools.build:gradle:7.0.0'
    }
}`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "gradle",
					PackageName:    "org.springframework:spring-core",
					Version:        "5.3.0",
					Locations: []models.Location{
						{Line: 10},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "junit:junit",
					Version:        "4.13",
					Locations: []models.Location{
						{Line: 11},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "com.google.guava:guava",
					Version:        "30.1-jre",
					Locations: []models.Location{
						{Line: 12},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "org.apache.commons:commons-lang3",
					Version:        "3.12.0",
					Locations: []models.Location{
						{Line: 13},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "com.android.tools.build:gradle",
					Version:        "7.0.0",
					Locations: []models.Location{
						{Line: 18},
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpFile, err := os.CreateTemp("", "build.gradle")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			// Write content to temp file
			_, err = tmpFile.WriteString(tt.content)
			if err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			// Parse the file
			parser := &GradleParser{}
			pkgs, err := parser.Parse(tmpFile.Name())

			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(pkgs) != len(tt.expectedPkgs) {
				t.Errorf("Expected %d packages, got %d", len(tt.expectedPkgs), len(pkgs))
			}

			for i, pkg := range pkgs {
				if i >= len(tt.expectedPkgs) {
					break
				}
				expected := tt.expectedPkgs[i]
				if pkg.PackageManager != expected.PackageManager ||
					pkg.PackageName != expected.PackageName ||
					pkg.Version != expected.Version {
					t.Errorf("Package %d mismatch: got %+v, expected %+v", i, pkg, expected)
				}
				if len(pkg.Locations) > 0 && len(expected.Locations) > 0 {
					if pkg.Locations[0].Line != expected.Locations[0].Line {
						t.Errorf("Location line mismatch: got %d, expected %d", pkg.Locations[0].Line, expected.Locations[0].Line)
					}
				}
			}
		})
	}
}

func TestGradleParser_ParseFile(t *testing.T) {
	// Test with actual file
	parser := &GradleParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "build.gradle"))
	if err != nil {
		t.Fatalf("Failed to parse build.gradle: %v", err)
	}

	if len(pkgs) == 0 {
		t.Errorf("Expected packages, got none")
	}

	for _, pkg := range pkgs {
		if pkg.PackageManager != "gradle" {
			t.Errorf("Expected package manager 'gradle', got '%s'", pkg.PackageManager)
		}
		if pkg.PackageName == "" {
			t.Errorf("Package name is empty")
		}
		if pkg.Version == "" {
			t.Errorf("Version is empty")
		}
	}
}
