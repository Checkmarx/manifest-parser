package gradle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// GradleParser implements parsing of Gradle build files
type GradleParser struct{}

// Parse implements the Parser interface for Gradle build files
func (p *GradleParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	manifestContent := string(content)

	// Extract variables
	variables := extractVariables(manifestFile, manifestContent)

	var packages []models.Package

	// Parse main dependencies
	mainDeps := parseDependencies(manifestContent, variables)
	for i := range mainDeps {
		mainDeps[i].FilePath = manifestFile
	}
	packages = append(packages, mainDeps...)

	return packages, nil
}

// extractVariables extracts variable definitions from the build file and gradle.properties
func extractVariables(manifestFile, content string) map[string]string {
	vars := make(map[string]string)

	// Read gradle.properties if exists
	gradlePropsPath := filepath.Join(filepath.Dir(manifestFile), "gradle.properties")
	if propsContent, err := os.ReadFile(gradlePropsPath); err == nil {
		for _, line := range strings.Split(string(propsContent), "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					vars[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Extract from ext blocks (Groovy)
	extPattern := regexp.MustCompile(`(?s)ext\s*\{([^}]+)\}`)
	if matches := extPattern.FindStringSubmatch(content); len(matches) > 1 {
		extContent := matches[1]
		// Simple key = 'value' or key: 'value'
		varPatterns := []*regexp.Regexp{
			regexp.MustCompile(`(\w+)\s*=\s*['"]([^'"]+)['"]`),
			regexp.MustCompile(`(\w+)\s*:\s*['"]([^'"]+)['"]`),
		}
		for _, pattern := range varPatterns {
			for _, match := range pattern.FindAllStringSubmatch(extContent, -1) {
				if len(match) > 2 {
					vars[match[1]] = match[2]
				}
			}
		}
	}

	// Extract ext.key = 'value' (outside blocks)
	extVarPattern := regexp.MustCompile(`ext\.(\w+)\s*=\s*['"]([^'"]+)['"]`)
	for _, match := range extVarPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 2 {
			vars[match[1]] = match[2]
		}
	}

	// Extract Kotlin DSL val/const
	kotlinVarPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:val|const val)\s+(\w+)\s*=\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?:val|const val)\s+(\w+)\s*=\s*(\d+(?:\.\d+)*[^\s'"]*)`), // for versions without quotes
	}
	for _, pattern := range kotlinVarPatterns {
		for _, match := range pattern.FindAllStringSubmatch(content, -1) {
			if len(match) > 2 {
				vars[match[1]] = match[2]
			}
		}
	}

	return vars
}

type dependencyStatement struct {
	Line int
	Text string
}

// parseDependencies parses dependencies from the content
func parseDependencies(content string, variables map[string]string) []models.Package {
	var packages []models.Package

	statements := extractDependencyStatements(content)
	for _, stmt := range statements {
		for _, pkg := range parseDependencyStatement(stmt.Text, variables) {
			pkg.Locations = []models.Location{{Line: stmt.Line}}
			packages = append(packages, pkg)
		}
	}

	return packages
}

func extractDependencyStatements(content string) []dependencyStatement {
	startPattern := regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\b`)
	var statements []dependencyStatement
	var buffer strings.Builder
	active := false
	startLine := 0

	lines := strings.Split(content, "\n")
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}

		if !active {
			if startPattern.MatchString(line) {
				active = true
				startLine = i + 1
				buffer.Reset()
				buffer.WriteString(line)
				if dependencyStatementComplete(buffer.String()) {
					statements = append(statements, dependencyStatement{Line: startLine, Text: buffer.String()})
					active = false
				}
			}
			continue
		}

		buffer.WriteString(" ")
		buffer.WriteString(line)
		if dependencyStatementComplete(buffer.String()) {
			statements = append(statements, dependencyStatement{Line: startLine, Text: buffer.String()})
			active = false
		}
	}

	return statements
}

func dependencyStatementComplete(statement string) bool {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*['"]([^'"\)]+)['"]`),
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*\(\s*['"]([^'"\)]+)['"]\s*\)`),
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*\(\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]\s*\)`),
		regexp.MustCompile(`(?i)group\s*[:=]\s*['"]([^'"]+)['"].*name\s*[:=]\s*['"]([^'"]+)['"].*version\s*[:=]\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)group\s*[:=]\s*[^,\s]+.*name\s*[:=]\s*[^,\s]+.*version\s*[:=]\s*[^,\s]+`),
	}

	for _, pattern := range patterns {
		if pattern.MatchString(statement) {
			return true
		}
	}

	return false
}

func parseDependencyStatement(statement string, variables map[string]string) []models.Package {
	var packages []models.Package

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*['"]([^'"\)]+)['"]`),
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*\(\s*['"]([^'"\)]+)['"]\s*\)`),
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)\b(implementation|api|compile|compileOnly|runtime|runtimeOnly|testImplementation|testCompile|testRuntimeOnly|androidTestImplementation|annotationProcessor|classpath|kapt)\s*\(\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]\s*\)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(statement)
		if len(matches) > 0 {
			var group, name, version string
			if len(matches) == 3 {
				depStr := resolveVariables(matches[2], variables)
				parts := strings.Split(depStr, ":")
				if len(parts) >= 2 {
					group = parts[0]
					name = parts[1]
					if len(parts) > 2 {
						version = strings.Join(parts[2:], ":")
					}
				}
			} else if len(matches) == 5 {
				group = resolveVariables(matches[2], variables)
				name = resolveVariables(matches[3], variables)
				version = resolveVariables(matches[4], variables)
			}

			if group != "" && name != "" {
				packages = append(packages, models.Package{
					PackageManager: "gradle",
					PackageName:    group + ":" + name,
					Version:        cleanVersion(version),
					FilePath:       "",
					Locations:      []models.Location{{}},
				})
			}
		}
	}

	if len(packages) == 0 {
		if pkg := parseDependencyKeyValue(statement, variables); pkg != nil {
			packages = append(packages, *pkg)
		}
	}

	return packages
}

func parseDependencyKeyValue(statement string, variables map[string]string) *models.Package {
	fields := map[string]string{}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(group|name|version)\s*[:=]\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)(group|name|version)\s*[:=]\s*([A-Za-z_][A-Za-z0-9_]*)`),
	}

	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringSubmatch(statement, -1) {
			if len(match) > 2 {
				key := strings.ToLower(match[1])
				value := match[2]
				fields[key] = resolveVariables(value, variables)
			}
		}
	}

	if fields["group"] == "" || fields["name"] == "" {
		return nil
	}

	return &models.Package{
		PackageManager: "gradle",
		PackageName:    fields["group"] + ":" + fields["name"],
		Version:        cleanVersion(fields["version"]),
		FilePath:       "",
		Locations:      []models.Location{{}},
	}
}

// resolveVariables replaces ${var} or $var with values
func resolveVariables(str string, variables map[string]string) string {
	// ${var}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	str = re.ReplaceAllStringFunc(str, func(match string) string {
		varName := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if val, ok := variables[varName]; ok {
			return val
		}
		return match
	})

	// $var
	re = regexp.MustCompile(`\$(\w+)`)
	str = re.ReplaceAllStringFunc(str, func(match string) string {
		varName := strings.TrimPrefix(match, "$")
		if val, ok := variables[varName]; ok {
			return val
		}
		return match
	})

	return str
}

// cleanVersion handles version ranges and classifiers
func cleanVersion(version string) string {
	// Remove brackets for ranges, take the lower bound
	if strings.HasPrefix(version, "[") && strings.HasSuffix(version, "]") {
		version = strings.Trim(version, "[]")
		parts := strings.Split(version, ",")
		if len(parts) > 0 {
			version = strings.TrimSpace(parts[0])
		}
	}
	if strings.HasPrefix(version, "(") && strings.HasSuffix(version, ")") {
		version = strings.Trim(version, "()")
		parts := strings.Split(version, ",")
		if len(parts) > 0 {
			version = strings.TrimSpace(parts[0])
		}
	}
	// For now, keep classifiers as is
	return version
}

// findLineNumber finds the line number of a substring in content
func findLineNumber(content, substr string) int {
	index := strings.Index(content, substr)
	if index == -1 {
		return 0
	}
	return strings.Count(content[:index], "\n") + 1
}
