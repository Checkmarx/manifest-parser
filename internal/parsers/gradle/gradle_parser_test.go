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
						{Line: 9},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "junit:junit",
					Version:        "4.13",
					Locations: []models.Location{
						{Line: 10},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "com.google.guava:guava",
					Version:        "30.1-jre",
					Locations: []models.Location{
						{Line: 11},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "org.apache.commons:commons-lang3",
					Version:        "3.12.0",
					Locations: []models.Location{
						{Line: 12},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "com.android.tools.build:gradle",
					Version:        "7.0.0",
					Locations: []models.Location{
						{Line: 17},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "kotlin dsl dependency syntax",
			content: `val kotlinVersion = "1.4.32"

dependencies {
    implementation("org.springframework:spring-core:$kotlinVersion")
    implementation(
        "org.apache.commons:commons-lang3:3.12.0"
    )
    implementation(group = "com.google.guava", name = "guava", version = "30.1-jre")
    if (project.hasProperty("feature")) {
        testImplementation("junit:junit:$kotlinVersion")
    }
}`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "gradle",
					PackageName:    "org.springframework:spring-core",
					Version:        "1.4.32",
					Locations: []models.Location{
						{Line: 3},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "org.apache.commons:commons-lang3",
					Version:        "3.12.0",
					Locations: []models.Location{
						{Line: 4},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "com.google.guava:guava",
					Version:        "30.1-jre",
					Locations: []models.Location{
						{Line: 7},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "junit:junit",
					Version:        "1.4.32",
					Locations: []models.Location{
						{Line: 9},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "multi-line and conditional dependencies",
			content: `ext {
    featureVersion = '1.0.0'
}

dependencies {
    implementation(
        'org.springframework:spring-core:5.3.0'
    )
    implementation group: 'org.apache.commons',
        name: 'commons-lang3',
        version: '3.12.0'
    if (project.hasProperty('feature')) {
        testImplementation 'junit:junit:$featureVersion'
    }
    if (useRedux) {
        api group: 'com.google.guava',
            name: 'guava',
            version: '30.1-jre'
    }
}`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "gradle",
					PackageName:    "org.springframework:spring-core",
					Version:        "5.3.0",
					Locations: []models.Location{
						{Line: 5},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "org.apache.commons:commons-lang3",
					Version:        "3.12.0",
					Locations: []models.Location{
						{Line: 8},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "junit:junit",
					Version:        "1.0.0",
					Locations: []models.Location{
						{Line: 12},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "com.google.guava:guava",
					Version:        "30.1-jre",
					Locations: []models.Location{
						{Line: 15},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "gradle with version ranges",
			content: `dependencies {
    implementation 'org.springframework:spring-core:[1.0.0,2.0.0)'
    implementation 'org.junit:junit:(1.0,2.0]'
}`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "gradle",
					PackageName:    "org.springframework:spring-core",
					Version:        "latest",
					Locations: []models.Location{
						{Line: 1},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "org.junit:junit",
					Version:        "latest",
					Locations: []models.Location{
						{Line: 2},
					},
				},
			},
			expectedError: false,
		},
		{
			name: "gradle with prefix wildcards",
			content: `dependencies {
    implementation 'org.springframework:spring-core:1.0.+'
    implementation 'org.junit:junit:4.12.*'
}`,
			expectedPkgs: []models.Package{
				{
					PackageManager: "gradle",
					PackageName:    "org.springframework:spring-core",
					Version:        "latest",
					Locations: []models.Location{
						{Line: 1},
					},
				},
				{
					PackageManager: "gradle",
					PackageName:    "org.junit:junit",
					Version:        "latest",
					Locations: []models.Location{
						{Line: 2},
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
			t.Errorf("Version is empty for %s", pkg.PackageName)
		}
	}
}

func TestGradleParser_ParseFile_NoProjectReferences(t *testing.T) {
	parser := &GradleParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "build.gradle"))
	if err != nil {
		t.Fatalf("Failed to parse build.gradle: %v", err)
	}

	for _, pkg := range pkgs {
		if pkg.PackageName == ":core" || pkg.PackageName == ":app" || pkg.PackageName == ":security" {
			t.Errorf("Project reference should not be extracted as a package: %s", pkg.PackageName)
		}
	}
}

func TestGradleParser_ParseFile_VariableResolution(t *testing.T) {
	parser := &GradleParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "build.gradle"))
	if err != nil {
		t.Fatalf("Failed to parse build.gradle: %v", err)
	}

	for _, pkg := range pkgs {
		if pkg.PackageName == "org.springframework.boot:spring-boot-starter-web" {
			if pkg.Version != "2.5.0" {
				t.Errorf("Expected spring-boot-starter-web version '2.5.0', got '%s'", pkg.Version)
			}
			return
		}
	}
	t.Errorf("Expected to find org.springframework.boot:spring-boot-starter-web in packages")
}

func TestGradleParser_ProjectReferencesSkipped(t *testing.T) {
	content := `dependencies {
    implementation project(':core')
    implementation(project(':lib'))
    implementation 'org.apache.commons:commons-lang3:3.8'
    api project(":shared")
}`
	tmpFile, err := os.CreateTemp("", "build.gradle")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	parser := &GradleParser{}
	pkgs, err := parser.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkgs) != 1 {
		t.Fatalf("Expected 1 package, got %d: %+v", len(pkgs), pkgs)
	}
	if pkgs[0].PackageName != "org.apache.commons:commons-lang3" {
		t.Errorf("Expected commons-lang3, got %s", pkgs[0].PackageName)
	}
}

func TestGradleParser_PlatformDependencies(t *testing.T) {
	content := `dependencies {
    implementation platform('org.springframework.boot:spring-boot-dependencies:2.5.0')
    implementation enforcedPlatform('com.google.cloud:libraries-bom:26.1.0')
    implementation(platform("org.junit:junit-bom:5.9.0"))
    implementation 'org.springframework:spring-core:5.3.0'
}`
	tmpFile, err := os.CreateTemp("", "build.gradle")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	parser := &GradleParser{}
	pkgs, err := parser.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPkgs := map[string]string{
		"org.springframework.boot:spring-boot-dependencies": "2.5.0",
		"com.google.cloud:libraries-bom":                    "26.1.0",
		"org.junit:junit-bom":                               "5.9.0",
		"org.springframework:spring-core":                    "5.3.0",
	}

	if len(pkgs) != len(expectedPkgs) {
		t.Fatalf("Expected %d packages, got %d: %+v", len(expectedPkgs), len(pkgs), pkgs)
	}

	for _, pkg := range pkgs {
		expectedVersion, ok := expectedPkgs[pkg.PackageName]
		if !ok {
			t.Errorf("Unexpected package: %s", pkg.PackageName)
			continue
		}
		if pkg.Version != expectedVersion {
			t.Errorf("Package %s: expected version %s, got %s", pkg.PackageName, expectedVersion, pkg.Version)
		}
	}
}

func TestGradleParser_FileReferencesSkipped(t *testing.T) {
	content := `dependencies {
    implementation files('libs/local.jar')
    implementation fileTree(dir: 'libs', include: ['*.jar'])
    implementation 'org.apache.commons:commons-lang3:3.8'
}`
	tmpFile, err := os.CreateTemp("", "build.gradle")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	parser := &GradleParser{}
	pkgs, err := parser.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkgs) != 1 {
		t.Fatalf("Expected 1 package, got %d: %+v", len(pkgs), pkgs)
	}
	if pkgs[0].PackageName != "org.apache.commons:commons-lang3" {
		t.Errorf("Expected commons-lang3, got %s", pkgs[0].PackageName)
	}
}

func TestGradleParser_ExtendedConfigurations(t *testing.T) {
	content := `dependencies {
    debugImplementation 'com.facebook.stetho:stetho:1.6.0'
    releaseImplementation 'com.google.firebase:firebase-crashlytics:18.0.0'
    ksp 'com.google.dagger:dagger-compiler:2.44'
    compileOnlyApi 'org.projectlombok:lombok:1.18.24'
    testCompileOnly 'org.mockito:mockito-core:4.0.0'
    lintChecks 'com.android.tools.lint:lint-checks:30.0.0'
}`
	tmpFile, err := os.CreateTemp("", "build.gradle")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	parser := &GradleParser{}
	pkgs, err := parser.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedNames := []string{
		"com.facebook.stetho:stetho",
		"com.google.firebase:firebase-crashlytics",
		"com.google.dagger:dagger-compiler",
		"org.projectlombok:lombok",
		"org.mockito:mockito-core",
		"com.android.tools.lint:lint-checks",
	}

	if len(pkgs) != len(expectedNames) {
		t.Fatalf("Expected %d packages, got %d: %+v", len(expectedNames), len(pkgs), pkgs)
	}

	for i, pkg := range pkgs {
		if pkg.PackageName != expectedNames[i] {
			t.Errorf("Package %d: expected %s, got %s", i, expectedNames[i], pkg.PackageName)
		}
	}
}

func TestGradleParser_CommentedExtBlocksIgnored(t *testing.T) {
	content := `
// ext {
//     badVar = '0.0.0'
// }

ext {
    goodVar = '1.0.0'
}

dependencies {
    implementation "org.example:lib:$goodVar"
}`
	tmpFile, err := os.CreateTemp("", "build.gradle")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	parser := &GradleParser{}
	pkgs, err := parser.Parse(tmpFile.Name())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkgs) != 1 {
		t.Fatalf("Expected 1 package, got %d: %+v", len(pkgs), pkgs)
	}
	if pkgs[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0' from non-commented ext block, got '%s'", pkgs[0].Version)
	}
}

func TestGradleParser_ParentGradleProperties(t *testing.T) {
	// Create a directory structure: parent/child/
	parentDir, err := os.MkdirTemp("", "gradle-parent")
	if err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}
	defer os.RemoveAll(parentDir)

	childDir := filepath.Join(parentDir, "child")
	os.Mkdir(childDir, 0755)

	// Create settings.gradle in parent to mark it as project root
	os.WriteFile(filepath.Join(parentDir, "settings.gradle"), []byte("include ':child'"), 0644)

	// Create parent gradle.properties
	os.WriteFile(filepath.Join(parentDir, "gradle.properties"), []byte("parentVersion=3.0.0\nsharedVersion=1.0.0"), 0644)

	// Create child gradle.properties (overrides sharedVersion)
	os.WriteFile(filepath.Join(childDir, "gradle.properties"), []byte("sharedVersion=2.0.0"), 0644)

	// Create child build.gradle
	buildContent := `dependencies {
    implementation "org.example:parent-lib:$parentVersion"
    implementation "org.example:shared-lib:$sharedVersion"
}`
	buildFile := filepath.Join(childDir, "build.gradle")
	os.WriteFile(buildFile, []byte(buildContent), 0644)

	parser := &GradleParser{}
	pkgs, err := parser.Parse(buildFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("Expected 2 packages, got %d: %+v", len(pkgs), pkgs)
	}

	// Parent property should be resolved
	if pkgs[0].Version != "3.0.0" {
		t.Errorf("Expected parent-lib version '3.0.0', got '%s'", pkgs[0].Version)
	}
	// Child property should take precedence over parent
	if pkgs[1].Version != "2.0.0" {
		t.Errorf("Expected shared-lib version '2.0.0' (child overrides parent), got '%s'", pkgs[1].Version)
	}
}

func TestVersionCatalog_Parse(t *testing.T) {
	catalogContent := `[versions]
spring = "5.3.0"
guava = "30.1-jre"

[libraries]
spring-core = { module = "org.springframework:spring-core", version.ref = "spring" }
spring-web = { module = "org.springframework:spring-web", version = "5.2.0" }
guava = "com.google.guava:guava:30.1-jre"
commons = { group = "org.apache.commons", name = "commons-lang3", version.ref = "spring" }
`
	tmpFile, err := os.CreateTemp("", "libs.versions.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(catalogContent)
	tmpFile.Close()

	catalog := parseVersionCatalog(tmpFile.Name())
	if catalog == nil {
		t.Fatalf("Failed to parse version catalog")
	}

	// Check versions
	if catalog.Versions["spring"] != "5.3.0" {
		t.Errorf("Expected spring version '5.3.0', got '%s'", catalog.Versions["spring"])
	}
	if catalog.Versions["guava"] != "30.1-jre" {
		t.Errorf("Expected guava version '30.1-jre', got '%s'", catalog.Versions["guava"])
	}

	// Check libraries
	tests := []struct {
		key     string
		group   string
		name    string
		version string
	}{
		{"spring-core", "org.springframework", "spring-core", "5.3.0"},
		{"spring-web", "org.springframework", "spring-web", "5.2.0"},
		{"guava", "com.google.guava", "guava", "30.1-jre"},
		{"commons", "org.apache.commons", "commons-lang3", "5.3.0"},
	}

	for _, tt := range tests {
		lib, ok := catalog.Libraries[tt.key]
		if !ok {
			t.Errorf("Library '%s' not found in catalog", tt.key)
			continue
		}
		if lib.Group != tt.group {
			t.Errorf("Library '%s': expected group '%s', got '%s'", tt.key, tt.group, lib.Group)
		}
		if lib.Name != tt.name {
			t.Errorf("Library '%s': expected name '%s', got '%s'", tt.key, tt.name, lib.Name)
		}
		if lib.Version != tt.version {
			t.Errorf("Library '%s': expected version '%s', got '%s'", tt.key, tt.version, lib.Version)
		}
	}
}

func TestVersionCatalog_DependencyResolution(t *testing.T) {
	// Create directory structure with version catalog
	projectDir, err := os.MkdirTemp("", "gradle-catalog")
	if err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	defer os.RemoveAll(projectDir)

	gradleDir := filepath.Join(projectDir, "gradle")
	os.Mkdir(gradleDir, 0755)

	// Create settings.gradle to mark project root
	os.WriteFile(filepath.Join(projectDir, "settings.gradle"), []byte(""), 0644)

	// Create version catalog
	catalogContent := `[versions]
spring = "5.3.0"

[libraries]
spring-core = { module = "org.springframework:spring-core", version.ref = "spring" }
guava = "com.google.guava:guava:30.1-jre"
`
	os.WriteFile(filepath.Join(gradleDir, "libs.versions.toml"), []byte(catalogContent), 0644)

	// Create build.gradle with catalog references
	buildContent := `dependencies {
    implementation libs.spring.core
    implementation(libs.guava)
    implementation 'org.direct:dependency:1.0.0'
}`
	buildFile := filepath.Join(projectDir, "build.gradle")
	os.WriteFile(buildFile, []byte(buildContent), 0644)

	parser := &GradleParser{}
	pkgs, err := parser.Parse(buildFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPkgs := map[string]string{
		"org.direct:dependency":        "1.0.0",
		"org.springframework:spring-core": "5.3.0",
		"com.google.guava:guava":          "30.1-jre",
	}

	if len(pkgs) != len(expectedPkgs) {
		t.Fatalf("Expected %d packages, got %d: %+v", len(expectedPkgs), len(pkgs), pkgs)
	}

	for _, pkg := range pkgs {
		expectedVersion, ok := expectedPkgs[pkg.PackageName]
		if !ok {
			t.Errorf("Unexpected package: %s", pkg.PackageName)
			continue
		}
		if pkg.Version != expectedVersion {
			t.Errorf("Package %s: expected version %s, got %s", pkg.PackageName, expectedVersion, pkg.Version)
		}
	}
}

func TestIsProjectReference(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"implementation project(':core')", true},
		{"implementation(project(':core'))", true},
		{`implementation project(":core")`, true},
		{"api project(':shared')", true},
		{"implementation 'org.example:lib:1.0'", false},
		{`implementation("org.example:lib:1.0")`, false},
	}

	for _, tt := range tests {
		result := isProjectReference(tt.input)
		if result != tt.expected {
			t.Errorf("isProjectReference(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsFileReference(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"implementation files('libs/local.jar')", true},
		{"implementation fileTree(dir: 'libs', include: ['*.jar'])", true},
		{"implementation(files('libs/local.jar'))", true},
		{"implementation 'org.example:lib:1.0'", false},
	}

	for _, tt := range tests {
		result := isFileReference(tt.input)
		if result != tt.expected {
			t.Errorf("isFileReference(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizePlatformDependency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"implementation platform('org.springframework.boot:spring-boot-dependencies:2.5.0')",
			"implementation 'org.springframework.boot:spring-boot-dependencies:2.5.0'",
		},
		{
			"implementation enforcedPlatform('com.google.cloud:libraries-bom:26.1.0')",
			"implementation 'com.google.cloud:libraries-bom:26.1.0'",
		},
		{
			`implementation(platform("org.junit:junit-bom:5.9.0"))`,
			`implementation("org.junit:junit-bom:5.9.0")`,
		},
		{
			"implementation 'org.example:lib:1.0'",
			"implementation 'org.example:lib:1.0'",
		},
	}

	for _, tt := range tests {
		result := normalizePlatformDependency(tt.input)
		if result != tt.expected {
			t.Errorf("normalizePlatformDependency(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestVersionCatalogParser_ParseFile(t *testing.T) {
	// Test parsing libs.versions.toml directly
	parser := &VersionCatalogParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "gradle", "libs.versions.toml"))
	if err != nil {
		t.Fatalf("Failed to parse libs.versions.toml: %v", err)
	}

	if len(pkgs) == 0 {
		t.Errorf("Expected packages from version catalog, got none")
	}

	// Verify expected packages are present
	expectedPackages := map[string]string{
		"org.springframework:spring-core":        "5.3.20",
		"org.springframework.boot:spring-boot-starter-web": "2.7.0",
		"com.google.guava:guava":                 "31.1-jre",
		"org.apache.logging.log4j:log4j-core":   "2.17.1",
	}

	found := make(map[string]bool)
	for _, pkg := range pkgs {
		if expectedVersion, ok := expectedPackages[pkg.PackageName]; ok {
			found[pkg.PackageName] = true
			if pkg.Version != expectedVersion {
				t.Errorf("Package %s: expected version %s, got %s", pkg.PackageName, expectedVersion, pkg.Version)
			}
			if pkg.PackageManager != "gradle" {
				t.Errorf("Expected package manager 'gradle', got '%s'", pkg.PackageManager)
			}
		}
	}

	for pkgName := range expectedPackages {
		if !found[pkgName] {
			t.Errorf("Expected package not found: %s", pkgName)
		}
	}
}

// TestGradleParser_LocationIndices asserts that the Gradle parser populates
// StartIndex and EndIndex on each Location, not just Line.
func TestGradleParser_LocationIndices(t *testing.T) {
	parser := &GradleParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "build.gradle"))
	if err != nil {
		t.Fatalf("Failed to parse build.gradle: %v", err)
	}

	// build.gradle line 40 (1-based):
	//         implementation 'org.apache.logging.log4j:log4j-core:2.14.0' // Log4Shell
	// 8 spaces + "implementation 'org.apache.logging.log4j:log4j-core:2.14.0'" (= 8 + 59 = 67)
	cases := map[string]struct {
		line, startIdx, endIdx int
	}{
		"org.apache.logging.log4j:log4j-core":             {39, 8, 67},
		"commons-collections:commons-collections":        {40, 8, 70},
		"org.springframework:spring-web":                 {45, 8, 69},
	}

	for _, pkg := range pkgs {
		want, ok := cases[pkg.PackageName]
		if !ok {
			continue
		}
		if len(pkg.Locations) == 0 {
			t.Errorf("%s: no Locations", pkg.PackageName)
			continue
		}
		got := pkg.Locations[0]
		if got.Line != want.line || got.StartIndex != want.startIdx || got.EndIndex != want.endIdx {
			t.Errorf("%s: got Location{Line=%d, Start=%d, End=%d}, want {Line=%d, Start=%d, End=%d}",
				pkg.PackageName, got.Line, got.StartIndex, got.EndIndex, want.line, want.startIdx, want.endIdx)
		}
	}
}

// TestComputeGradleLocations_MultiLine asserts that a dependency spanning multiple
// source lines produces one Location per non-empty contributing line (Maven-style).
func TestComputeGradleLocations_MultiLine(t *testing.T) {
	raws := []rawLineInfo{
		{LineNum: 5, Content: "    implementation("},
		{LineNum: 6, Content: "        \"org.springframework:spring-core:5.3.0\""},
		{LineNum: 7, Content: "    )"},
	}
	locs := computeGradleLocations(raws)
	if len(locs) != 3 {
		t.Fatalf("expected 3 Locations, got %d", len(locs))
	}
	want := []models.Location{
		{Line: 5, StartIndex: 4, EndIndex: 19}, // "    implementation(" length 19
		{Line: 6, StartIndex: 8, EndIndex: 47}, // 8 spaces + "\"org.springframework:spring-core:5.3.0\"" (39) = 47
		{Line: 7, StartIndex: 4, EndIndex: 5},  // "    )" length 5
	}
	for i, w := range want {
		if locs[i] != w {
			t.Errorf("loc[%d]: got %+v, want %+v", i, locs[i], w)
		}
	}
}

// TestStripInlineComment verifies trailing // comments are removed but // inside
// strings is preserved.
func TestStripInlineComment(t *testing.T) {
	cases := []struct{ in, out string }{
		{"implementation 'foo:bar:1.0' // comment", "implementation 'foo:bar:1.0' "},
		{`implementation "https://example.com"`, `implementation "https://example.com"`},
		{"no comment here", "no comment here"},
		{"// whole line is a comment", ""},
	}
	for _, c := range cases {
		if got := stripInlineComment(c.in); got != c.out {
			t.Errorf("stripInlineComment(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestGradleParser_AndroidFlavorConfigurations(t *testing.T) {
	content := `
dependencies {
    // Standard
    implementation "androidx.core:core-ktx:1.13.1"
    api "com.squareup.retrofit2:retrofit:2.11.0"
    compileOnly "org.projectlombok:lombok:1.18.38"
    runtimeOnly "com.squareup.okhttp3:logging-interceptor:4.12.0"

    // Build-type specific
    debugImplementation "com.squareup.leakcanary:leakcanary-android:2.14"
    releaseImplementation "com.google.firebase:firebase-crashlytics:19.0.0"
    debugApi "com.google.code.gson:gson:2.13.1"
    releaseApi "com.google.guava:guava:33.2.1-android"

    // Flavor specific
    freeImplementation "com.google.android.gms:play-services-ads:24.4.0"
    paidImplementation "com.android.billingclient:billing:8.0.0"
    devImplementation "com.squareup.okhttp3:mockwebserver:4.12.0"
    prodImplementation "com.google.firebase:firebase-analytics:22.0.0"

    // Flavor + BuildType
    freeDebugImplementation "com.example:free-debug-sdk:1.0.0"
    paidReleaseImplementation "com.example:paid-release-sdk:1.0.0"

    // Multi-dimension flavors
    freeDevImplementation "com.example:free-dev-sdk:1.0.0"
    paidProdImplementation "com.example:paid-prod-sdk:1.0.0"

    // Test scoped
    testImplementation "junit:junit:4.13.2"
    testApi "org.mockito:mockito-core:5.12.0"
    testCompileOnly "org.projectlombok:lombok:1.18.38"
    testRuntimeOnly "org.junit.platform:junit-platform-launcher:1.12.2"
    androidTestImplementation "androidx.test.espresso:espresso-core:3.6.1"
    androidTestApi "androidx.test:runner:1.6.1"
    androidTestCompileOnly "org.projectlombok:lombok:1.18.38"
    androidTestRuntimeOnly "androidx.test:core:1.6.1"

    // Annotation processing
    annotationProcessor "com.google.dagger:dagger-compiler:2.57"
    kapt "com.google.dagger:hilt-compiler:2.57"
    kaptTest "com.google.dagger:hilt-compiler:2.57"
    kaptAndroidTest "com.google.dagger:hilt-compiler:2.57"
}`

	expected := map[string]string{
		"androidx.core:core-ktx":                          "1.13.1",
		"com.squareup.retrofit2:retrofit":                 "2.11.0",
		"org.projectlombok:lombok":                        "1.18.38",
		"com.squareup.okhttp3:logging-interceptor":        "4.12.0",
		"com.squareup.leakcanary:leakcanary-android":      "2.14",
		"com.google.firebase:firebase-crashlytics":        "19.0.0",
		"com.google.code.gson:gson":                       "2.13.1",
		"com.google.guava:guava":                          "33.2.1-android",
		"com.google.android.gms:play-services-ads":        "24.4.0",
		"com.android.billingclient:billing":               "8.0.0",
		"com.squareup.okhttp3:mockwebserver":              "4.12.0",
		"com.google.firebase:firebase-analytics":          "22.0.0",
		"com.example:free-debug-sdk":                      "1.0.0",
		"com.example:paid-release-sdk":                    "1.0.0",
		"com.example:free-dev-sdk":                        "1.0.0",
		"com.example:paid-prod-sdk":                       "1.0.0",
		"junit:junit":                                     "4.13.2",
		"org.mockito:mockito-core":                        "5.12.0",
		"org.junit.platform:junit-platform-launcher":      "1.12.2",
		"androidx.test.espresso:espresso-core":            "3.6.1",
		"androidx.test:runner":                            "1.6.1",
		"androidx.test:core":                              "1.6.1",
		"com.google.dagger:dagger-compiler":               "2.57",
		"com.google.dagger:hilt-compiler":                 "2.57",
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.gradle")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &GradleParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make(map[string]string)
	for _, p := range pkgs {
		got[p.PackageName] = p.Version
	}

	for name, wantVer := range expected {
		if gotVer, ok := got[name]; !ok {
			t.Errorf("missing package %q", name)
		} else if gotVer != wantVer {
			t.Errorf("package %q: version = %q, want %q", name, gotVer, wantVer)
		}
	}
}
