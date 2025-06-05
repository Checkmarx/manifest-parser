package maven

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestMavenPomParser_Parse(t *testing.T) {

	tests := []struct {
		name          string
		content       string
		expectedPkgs  []models.Package
		expectedError bool
	}{
		{
			name: "basic pom file",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>test-project</artifactId>
    <version>1.0.0</version>
    <dependencies>
        <dependency>
            <groupId>org.springframework</groupId>
            <artifactId>spring-core</artifactId>
            <version>5.3.0</version>
        </dependency>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>4.13</version>
            <scope>test</scope>
        </dependency>
    </dependencies>
</project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "maven",
					PackageName:    "org.springframework:spring-core",
					Version:        "5.3.0",
					Locations: []models.Location{{
						Line:       8,
						StartIndex: 12,
						EndIndex:   48,
					}},
				},
				{
					PackageManager: "maven",
					PackageName:    "junit:junit",
					Version:        "4.13",
					Locations: []models.Location{{
						Line:       13,
						StartIndex: 12,
						EndIndex:   42,
					}},
				},
			},
			expectedError: false,
		},
		{
			name: "pom with version ranges",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>org.example</groupId>
            <artifactId>test-lib</artifactId>
            <version>[1.0.0,2.0.0)</version>
        </dependency>
    </dependencies>
</project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "maven",
					PackageName:    "org.example:test-lib",
					Version:        "latest",
					Locations: []models.Location{{
						Line:       5,
						StartIndex: 12,
						EndIndex:   45,
					}},
				},
			},
			expectedError: false,
		},
		{
			name: "pom with properties",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <properties>
        <spring.version>5.3.0</spring.version>
        <junit.version>4.13</junit.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework</groupId>
            <artifactId>spring-core</artifactId>
            <version>${spring.version}</version>
        </dependency>
        <dependency>
            <groupId>junit</groupId>
            <artifactId>junit</artifactId>
            <version>${junit.version}</version>
        </dependency>
    </dependencies>
</project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "maven",
					PackageName:    "org.springframework:spring-core",
					Version:        "5.3.0",
					Locations: []models.Location{{
						Line:       9,
						StartIndex: 12,
						EndIndex:   48,
					}},
				},
				{
					PackageManager: "maven",
					PackageName:    "junit:junit",
					Version:        "4.13",
					Locations: []models.Location{{
						Line:       14,
						StartIndex: 12,
						EndIndex:   42,
					}},
				},
			},
			expectedError: false,
		},
		{
			name: "pom with dependency management",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>org.springframework</groupId>
                <artifactId>spring-core</artifactId>
                <version>5.3.0</version>
            </dependency>
        </dependencies>
    </dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>org.springframework</groupId>
            <artifactId>spring-core</artifactId>
        </dependency>
    </dependencies>
</project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "maven",
					PackageName:    "org.springframework:spring-core",
					Version:        "5.3.0",
					Locations: []models.Location{{
						Line:       6,
						StartIndex: 16,
						EndIndex:   52,
					}},
				},
			},
			expectedError: false,
		},
		{
			name: "pom with nested properties",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <properties>
        <spring.version>5.3.0</spring.version>
        <version.suffix>.RELEASE</version.suffix>
        <full.version>${spring.version}${version.suffix}</full.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework</groupId>
            <artifactId>spring-core</artifactId>
            <version>${full.version}</version>
        </dependency>
    </dependencies>
</project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "maven",
					PackageName:    "org.springframework:spring-core",
					Version:        "${spring.version}${version.suffix}",
					Locations: []models.Location{{
						Line:       10,
						StartIndex: 12,
						EndIndex:   48,
					}},
				},
			},
			expectedError: false,
		},
		{
			name: "pom with version ranges in dependency management",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>org.example</groupId>
                <artifactId>test-lib</artifactId>
                <version>[1.0.0,2.0.0)</version>
            </dependency>
        </dependencies>
    </dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>org.example</groupId>
            <artifactId>test-lib</artifactId>
        </dependency>
    </dependencies>
</project>`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "maven",
					PackageName:    "org.example:test-lib",
					Version:        "latest",
					Locations: []models.Location{{
						Line:       6,
						StartIndex: 16,
						EndIndex:   49,
					}},
				},
			},
			expectedError: false,
		},
		{
			name: "malformed pom file",
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>org.example</groupId>
            <artifactId>test-lib</artifactId>
            <version>1.0.0</version>
        </dependency>
    </dependencies>
</project`,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempDir := t.TempDir()
			testFile := filepath.Join(tempDir, "pom.xml")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Create parser and parse file
			parser := &MavenPomParser{}
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

			// Update file paths in expected packages
			for i := range tt.expectedPkgs {
				tt.expectedPkgs[i].FilePath = testFile
			}

			// Check packages
			testdata.ValidatePackages(t, pkgs, tt.expectedPkgs)
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		props       map[string]string
		managedDeps []MavenDependency
		groupId     string
		artifactId  string
		expected    string
	}{
		{
			name:     "exact version",
			version:  "1.2.3",
			props:    make(map[string]string),
			expected: "1.2.3",
		},
		{
			name:     "version range",
			version:  "[1.2.3,2.0.0)",
			props:    make(map[string]string),
			expected: "latest",
		},
		{
			name:     "property substitution",
			version:  "${spring.version}",
			props:    map[string]string{"spring.version": "5.3.0"},
			expected: "5.3.0",
		},
		{
			name:     "missing property",
			version:  "${missing.prop}",
			props:    make(map[string]string),
			expected: "${missing.prop}",
		},
		{
			name:    "managed dependency version",
			version: "",
			props:   make(map[string]string),
			managedDeps: []MavenDependency{
				{
					GroupId:    "org.example",
					ArtifactId: "test-lib",
					Version:    "1.0.0",
				},
			},
			groupId:    "org.example",
			artifactId: "test-lib",
			expected:   "1.0.0",
		},
		{
			name:    "managed dependency with range",
			version: "",
			props:   make(map[string]string),
			managedDeps: []MavenDependency{
				{
					GroupId:    "org.example",
					ArtifactId: "test-lib",
					Version:    "[1.0.0,2.0.0)",
				},
			},
			groupId:    "org.example",
			artifactId: "test-lib",
			expected:   "latest",
		},
		{
			name:    "managed dependency with property",
			version: "",
			props:   map[string]string{"lib.version": "1.0.0"},
			managedDeps: []MavenDependency{
				{
					GroupId:    "org.example",
					ArtifactId: "test-lib",
					Version:    "${lib.version}",
				},
			},
			groupId:    "org.example",
			artifactId: "test-lib",
			expected:   "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveVersion(tt.version, tt.props, tt.managedDeps, tt.groupId, tt.artifactId)
			if result != tt.expected {
				t.Errorf("resolveVersion(%q, %v, %v, %q, %q) = %q, want %q",
					tt.version, tt.props, tt.managedDeps, tt.groupId, tt.artifactId,
					result, tt.expected)
			}
		})
	}
}

func TestMavenPomParser_ParseRealFile(t *testing.T) {
	parser := &MavenPomParser{}
	manifestFile := "../../../internal/testdata/pom.xml"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expectedPackages := []models.Package{
		{
			PackageManager: "maven",
			PackageName:    "org.mockito:mockito-core",
			Version:        "latest",
			Locations: []models.Location{{
				Line:       68,
				StartIndex: 12,
				EndIndex:   49,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "com.checkmarx:cx-integrations-common",
			Version:        "0.0.319",
			Locations: []models.Location{{
				Line:       74,
				StartIndex: 12,
				EndIndex:   59,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "com.checkmarx:cx-interceptors-lib",
			Version:        "0.1.58",
			Locations: []models.Location{{
				Line:       80,
				StartIndex: 12,
				EndIndex:   56,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "org.apache.httpcomponents.client5:httpclient5",
			Version:        "5.4.3",
			Locations: []models.Location{{
				Line:       27,
				StartIndex: 16,
				EndIndex:   52,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "org.apache.httpcomponents.client5:httpclient5-fluent",
			Version:        "5.4.3",
			Locations: []models.Location{{
				Line:       32,
				StartIndex: 16,
				EndIndex:   59,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "org.projectlombok:lombok",
			Version:        "latest",
			Locations: []models.Location{{
				Line:       93,
				StartIndex: 12,
				EndIndex:   43,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "org.yaml:snakeyaml",
			Version:        "latest",
			Locations: []models.Location{{
				Line:       97,
				StartIndex: 12,
				EndIndex:   46,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "org.apache.tomcat.embed:tomcat-embed-core",
			Version:        "latest",
			Locations: []models.Location{{
				Line:       101,
				StartIndex: 12,
				EndIndex:   54,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "org.springframework.boot:spring-boot-starter-web",
			Version:        "latest",
			Locations: []models.Location{{
				Line:       105,
				StartIndex: 12,
				EndIndex:   60,
			}},
			FilePath: manifestFile,
		},
		{
			PackageManager: "maven",
			PackageName:    "com.fasterxml.jackson.dataformat:jackson-dataformat-smile",
			Version:        "2.18.2",
			Locations: []models.Location{{
				Line:       119,
				StartIndex: 12,
				EndIndex:   61,
			}},
			FilePath: manifestFile,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
