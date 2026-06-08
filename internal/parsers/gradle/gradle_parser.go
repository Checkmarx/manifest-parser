package gradle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// configKeywords defines all supported Gradle dependency configuration keywords
var configKeywords = `implementation|api|compile|compileOnly|runtime|runtimeOnly|` +
	`testImplementation|testCompile|testCompileOnly|testRuntimeOnly|` +
	`androidTestImplementation|debugImplementation|releaseImplementation|` +
	`annotationProcessor|classpath|kapt|ksp|compileOnlyApi|` +
	`testFixturesImplementation|testFixturesApi|lintChecks`

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

	// Load version catalog if available
	var catalog *VersionCatalog
	if catalogPath := findVersionCatalog(manifestFile); catalogPath != "" {
		catalog = parseVersionCatalog(catalogPath)
	}

	var packages []models.Package

	// Parse main dependencies
	mainDeps := parseDependencies(manifestContent, variables)
	for i := range mainDeps {
		mainDeps[i].FilePath = manifestFile
	}
	packages = append(packages, mainDeps...)

	// Parse version catalog dependencies (libs.xxx references)
	if catalog != nil {
		catalogDeps := parseVersionCatalogDependencies(manifestContent, catalog)
		for i := range catalogDeps {
			catalogDeps[i].FilePath = manifestFile
		}
		packages = append(packages, catalogDeps...)
	}

	return packages, nil
}

// extractVariables extracts variable definitions from the build file and gradle.properties
func extractVariables(manifestFile, content string) map[string]string {
	vars := make(map[string]string)

	// Read gradle.properties if exists
	gradlePropsPath := filepath.Join(filepath.Dir(manifestFile), "gradle.properties")
	if propsContent, err := os.ReadFile(gradlePropsPath); err == nil {
		parsePropertiesInto(string(propsContent), vars)
	}

	// Walk up to project root for parent gradle.properties
	projectRoot := findProjectRoot(filepath.Dir(manifestFile))
	if projectRoot != filepath.Dir(manifestFile) {
		rootPropsPath := filepath.Join(projectRoot, "gradle.properties")
		if propsContent, err := os.ReadFile(rootPropsPath); err == nil {
			parsePropertiesInto(string(propsContent), vars)
		}
	}

	// Extract from ext blocks (Groovy) — handle all ext blocks, filter commented lines
	extPattern := regexp.MustCompile(`(?s)ext\s*\{([^}]+)\}`)
	for _, matches := range extPattern.FindAllStringSubmatch(content, -1) {
		if len(matches) > 1 {
			// Filter commented lines from ext block content
			var filteredLines []string
			for _, line := range strings.Split(matches[1], "\n") {
				trimmed := strings.TrimSpace(line)
				if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "*") {
					filteredLines = append(filteredLines, line)
				}
			}
			extContent := strings.Join(filteredLines, "\n")
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
	Line     int
	Text     string
	RawLines []rawLineInfo
}

// rawLineInfo records a single source line that contributes to a dependency statement.
// Content is the raw line with \r stripped (no other trimming) so byte offsets stay accurate.
type rawLineInfo struct {
	LineNum int
	Content string
}

// parseDependencies parses dependencies from the content
func parseDependencies(content string, variables map[string]string) []models.Package {
	var packages []models.Package

	statements := extractDependencyStatements(content)
	for _, stmt := range statements {
		locations := computeGradleLocations(stmt.RawLines)
		for _, pkg := range parseDependencyStatement(stmt.Text, variables) {
			pkg.Locations = locations
			packages = append(packages, pkg)
		}
	}

	return packages
}

func extractDependencyStatements(content string) []dependencyStatement {
	startPattern := regexp.MustCompile(`(?i)\b(` + configKeywords + `)\b`)
	var statements []dependencyStatement
	var buffer strings.Builder
	var rawLines []rawLineInfo
	active := false
	startLine := 0

	lines := strings.Split(content, "\n")
	// Strip trailing \r so byte offsets are consistent on CRLF files
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}

	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}

		if !active {
			if startPattern.MatchString(line) {
				// Skip non-Maven dependency references
				if isProjectReference(line) || isFileReference(line) || isVersionCatalogReference(line) {
					continue
				}
				active = true
				startLine = i
				buffer.Reset()
				buffer.WriteString(line)
				rawLines = []rawLineInfo{{LineNum: i, Content: raw}}
				normalized := normalizePlatformDependency(buffer.String())
				if dependencyStatementComplete(normalized) {
					statements = append(statements, dependencyStatement{Line: startLine, Text: normalized, RawLines: rawLines})
					active = false
				}
			}
			continue
		}

		buffer.WriteString(" ")
		buffer.WriteString(line)
		rawLines = append(rawLines, rawLineInfo{LineNum: i, Content: raw})
		normalized := normalizePlatformDependency(buffer.String())
		if dependencyStatementComplete(normalized) {
			statements = append(statements, dependencyStatement{Line: startLine, Text: normalized, RawLines: rawLines})
			active = false
		}
	}

	return statements
}

// computeGradleLocations emits one Location per contributing source line (Maven-style).
// For each line: StartIndex = offset of first non-whitespace character; EndIndex = end
// of code on the line, with any trailing // ... comment and trailing whitespace stripped.
func computeGradleLocations(rawLines []rawLineInfo) []models.Location {
	locations := make([]models.Location, 0, len(rawLines))
	for _, rl := range rawLines {
		code := stripInlineComment(rl.Content)
		code = strings.TrimRight(code, " \t")
		if strings.TrimSpace(code) == "" {
			continue
		}
		startIdx := len(rl.Content) - len(strings.TrimLeft(rl.Content, " \t"))
		locations = append(locations, models.Location{
			Line:       rl.LineNum,
			StartIndex: startIdx,
			EndIndex:   len(code),
		})
	}
	if len(locations) == 0 {
		return nil
	}
	return locations
}

// stripInlineComment removes a trailing `// ...` from a Gradle source line,
// taking quote state into account so // inside a quoted string is preserved.
func stripInlineComment(line string) string {
	inSingle := false
	inDouble := false
	for i := 0; i < len(line)-1; i++ {
		ch := line[i]
		switch {
		case ch == '\\' && (inSingle || inDouble):
			i++ // skip escaped char
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case !inSingle && !inDouble && ch == '/' && line[i+1] == '/':
			return line[:i]
		}
	}
	return line
}

func dependencyStatementComplete(statement string) bool {
	kw := configKeywords
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*\(\s*['"]([^'"]+)['"]\s*\)`),
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*\(\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]\s*\)`),
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

	kw := configKeywords
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*\(\s*['"]([^'"]+)['"]\s*\)`),
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]`),
		regexp.MustCompile(`(?i)\b(` + kw + `)\s*\(\s*group\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*name\s*[:=]\s*['"]([^'"]+)['"]\s*,\s*version\s*[:=]\s*['"]([^'"]+)['"]\s*\)`),
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
	if version == "" {
		return "latest"
	}
	// Check for any range or wildcard patterns
	if strings.ContainsAny(version, "[]()^~*><") || strings.Contains(version, "+") {
		return "latest"
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

// parsePropertiesInto parses key=value properties into the given map (does not overwrite existing keys)
func parsePropertiesInto(content string, vars map[string]string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				if _, exists := vars[key]; !exists {
					vars[key] = strings.TrimSpace(parts[1])
				}
			}
		}
	}
}

// findProjectRoot walks up from dir looking for settings.gradle or settings.gradle.kts
func findProjectRoot(dir string) string {
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, "settings.gradle")); err == nil {
			return current
		}
		if _, err := os.Stat(filepath.Join(current, "settings.gradle.kts")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return dir
}

// isProjectReference checks if a dependency statement is a project reference
func isProjectReference(statement string) bool {
	pattern := regexp.MustCompile(`(?i)\b(?:` + configKeywords + `)\s*(?:\(\s*)?project\s*\(`)
	return pattern.MatchString(statement)
}

// isFileReference checks if a dependency statement is a file reference (files/fileTree)
func isFileReference(statement string) bool {
	pattern := regexp.MustCompile(`(?i)\b(?:` + configKeywords + `)\s*(?:\(\s*)?(?:files|fileTree)\s*\(`)
	return pattern.MatchString(statement)
}

// isVersionCatalogReference checks if a dependency uses version catalog syntax (libs.xxx)
func isVersionCatalogReference(statement string) bool {
	pattern := regexp.MustCompile(`(?i)\b(?:` + configKeywords + `)\s*(?:\(\s*)?libs\.`)
	return pattern.MatchString(statement)
}

// normalizePlatformDependency strips platform() and enforcedPlatform() wrappers
func normalizePlatformDependency(statement string) string {
	pattern := regexp.MustCompile(`\b(?:platform|enforcedPlatform)\s*\(\s*(['"][^'"]+['"])\s*\)`)
	return pattern.ReplaceAllString(statement, "$1")
}
