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

	lines := strings.Split(string(content), "\n")

	// Extract variables
	variables := extractVariables(manifestFile, string(content))

	var packages []models.Package

	// Parse main dependencies
	mainDeps := parseDependencies(string(content), lines, variables, false)
	for i := range mainDeps {
		mainDeps[i].FilePath = manifestFile
	}
	packages = append(packages, mainDeps...)

	// Note: Buildscript dependencies are also parsed as main for simplicity

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

// parseDependencies parses dependencies from the content
func parseDependencies(content string, lines []string, variables map[string]string, isBuildscript bool) []models.Package {
	var packages []models.Package

	// Patterns for different dependency declarations
	patterns := []*regexp.Regexp{
		// String notation: implementation 'group:name:version'
		regexp.MustCompile(`(?i)(implementation|api|compile|runtime|testImplementation|testCompile|androidTestImplementation|classpath)\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)(implementation|api|compile|runtime|testImplementation|testCompile|androidTestImplementation|classpath)\s*\(\s*['"]([^'"]+)['"]\s*\)`),
		// Map notation: implementation group: 'g', name: 'n', version: 'v'
		regexp.MustCompile(`(?i)(implementation|api|compile|runtime|testImplementation|testCompile|androidTestImplementation|classpath)\s*group\s*:\s*['"]([^'"]+)['"]\s*,\s*name\s*:\s*['"]([^'"]+)['"]\s*,\s*version\s*:\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)(implementation|api|compile|runtime|testImplementation|testCompile|androidTestImplementation|classpath)\s*\(\s*group\s*:\s*['"]([^'"]+)['"]\s*,\s*name\s*:\s*['"]([^'"]+)['"]\s*,\s*version\s*:\s*['"]([^'"]+)['"]\s*\)`),
	}

	depsLines := strings.Split(content, "\n")
	for _, line := range depsLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 0 {
				var group, name, version string
				if len(matches) == 3 {
					// String notation
					depStr := resolveVariables(matches[2], variables)
					parts := strings.Split(depStr, ":")
					if len(parts) >= 2 {
						group = parts[0]
						name = parts[1]
						if len(parts) > 2 {
							version = parts[2]
						}
					}
				} else if len(matches) == 5 {
					// Map notation
					group = resolveVariables(matches[2], variables)
					name = resolveVariables(matches[3], variables)
					version = resolveVariables(matches[4], variables)
				}

				if group != "" && name != "" {
					// Handle version ranges and classifiers
					cleanVersion := cleanVersion(version)

					// Find line number
					lineNum := findLineNumber(content, line)

					packages = append(packages, models.Package{
						PackageManager: "gradle",
						PackageName:    group + ":" + name,
						Version:        cleanVersion,
						FilePath:       "", // Will be set later
						Locations: []models.Location{
							{Line: lineNum},
						},
					})
				}
			}
		}
	}

	return packages
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
