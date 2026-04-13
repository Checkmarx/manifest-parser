package sbt

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// SbtParser implements parsing of SBT .sbt files (build.sbt, plugins.sbt, etc.)
type SbtParser struct{}

var (
	// varRegex matches Scala variable declarations:
	//   val name = "value"
	//   lazy val name = "value"
	//   def name = "value"
	varRegex = regexp.MustCompile(`^\s*(?:lazy\s+)?(?:val|def)\s+(\w+)\s*=\s*"([^"]+)"`)

	// depRegex matches SBT dependency declarations:
	//   "groupId" % "artifactId" % "version"
	//   "groupId" %% "artifactId" % "version"
	//   "groupId" %%% "artifactId" % "version"
	//   "groupId" % "artifactId" % variableName
	// With optional trailing scope: % "test" or % Test
	depRegex = regexp.MustCompile(`"([^"]+)"\s+(%{1,3})\s+"([^"]+)"\s+%\s+(?:"([^"]+)"|(\w+))(?:\s+%\s+(?:"[^"]*"|\w+))?`)
)

// extractVariables scans lines for val declarations and returns a variable map
func extractVariables(lines []string) map[string]string {
	vars := make(map[string]string)
	inBlockComment := false

	for _, rawLine := range lines {
		line := stripComments(rawLine, &inBlockComment)
		if inBlockComment {
			continue
		}
		if match := varRegex.FindStringSubmatch(line); match != nil {
			vars[match[1]] = match[2]
		}
	}

	return vars
}

// resolveVersion resolves a version string using the variable map
func resolveVersion(version string, vars map[string]string) string {
	if version == "" {
		return "latest"
	}
	// If it looks like a literal version (starts with digit or contains dots/hyphens typical of versions), return as-is
	if len(version) > 0 && (version[0] >= '0' && version[0] <= '9') {
		return version
	}
	// Try to resolve as a variable
	if resolved, exists := vars[version]; exists {
		return resolved
	}
	return "latest"
}

// stripComments removes comments from a line and tracks block comment state
func stripComments(line string, inBlockComment *bool) string {
	if *inBlockComment {
		if idx := strings.Index(line, "*/"); idx >= 0 {
			*inBlockComment = false
			line = line[idx+2:]
		} else {
			return ""
		}
	}

	// Handle inline block comments: /* ... */ on the same line
	for {
		startIdx := strings.Index(line, "/*")
		if startIdx < 0 {
			break
		}
		endIdx := strings.Index(line[startIdx+2:], "*/")
		if endIdx >= 0 {
			// Block comment opens and closes on same line
			line = line[:startIdx] + line[startIdx+2+endIdx+2:]
		} else {
			// Block comment opens but doesn't close — entering block comment
			*inBlockComment = true
			line = line[:startIdx]
			break
		}
	}

	// Handle single-line comments
	if idx := strings.Index(line, "//"); idx >= 0 {
		line = line[:idx]
	}

	return line
}

// modifierKeywords are SBT dependency modifiers that should be excluded from the location span.
// The EndIndex should cover only the core "g" % "a" % "v" declaration.
var modifierKeywords = []string{
	"exclude(",
	"excludeAll(",
	"intransitive()",
	"withSources()",
	"withJavadoc()",
	"classifier ",
	"classifier(",
	"cross ",
	"cross(",
}

// computeLocationIndices calculates start and end indices for a dependency in a raw line.
// StartIndex = position of the first quote of the groupId.
// EndIndex = end of the core dependency declaration, excluding modifiers, comments, and trailing punctuation.
func computeLocationIndices(rawLine string, groupId string) (int, int) {
	// StartIndex: position of the first quote of the groupId
	searchStr := `"` + groupId + `"`
	startIdx := strings.Index(rawLine, searchStr)
	if startIdx < 0 {
		startIdx = 0
	}

	// Start with the full line
	endIdx := len(rawLine)

	// If there's a trailing comment, stop before it
	if commentIdx := strings.Index(rawLine, "//"); commentIdx >= 0 && commentIdx < endIdx {
		endIdx = commentIdx
	}

	// Trim known dependency modifiers first (before punctuation removal,
	// so keywords like "intransitive()" are still intact when searched)
	endIdx = trimModifiers(rawLine, startIdx, endIdx)

	// Trim trailing whitespace, commas, and closing parentheses
	endIdx = trimTrailingPunctuation(rawLine, endIdx)

	return startIdx, endIdx
}

// trimTrailingPunctuation removes trailing whitespace, commas, and closing parens from the end boundary
func trimTrailingPunctuation(line string, endIdx int) int {
	for endIdx > 0 {
		ch := line[endIdx-1]
		if ch == ' ' || ch == '\t' || ch == ',' || ch == ')' {
			endIdx--
		} else {
			break
		}
	}
	return endIdx
}

// trimModifiers scans the region [startIdx, endIdx) for modifier keywords and truncates endIdx
// to exclude them. Works backwards so nested modifiers are stripped in order.
func trimModifiers(line string, startIdx int, endIdx int) int {
	region := line[startIdx:endIdx]
	for _, kw := range modifierKeywords {
		if idx := strings.Index(region, kw); idx >= 0 {
			// Truncate at the modifier keyword
			candidate := startIdx + idx
			// Only trim if the modifier comes after the core dependency (at least "g" % "a" % "v")
			if candidate > startIdx && candidate < endIdx {
				endIdx = candidate
			}
		}
	}
	return endIdx
}

// Parse implements the Parser interface for SBT build.sbt files
func (p *SbtParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")

	// Pass 1: Extract variable definitions
	vars := extractVariables(lines)

	// Pass 2: Extract dependencies
	var packages []models.Package
	seen := make(map[string]bool)
	inBlockComment := false

	for lineNum, rawLine := range lines {
		line := stripComments(rawLine, &inBlockComment)
		if inBlockComment {
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to extract dependency from this line
		match := depRegex.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		groupId := match[1]
		// match[2] is the operator (%, %%, %%%) — captured but not used
		artifactId := match[3]
		quotedVersion := match[4] // version from quoted string
		bareVersion := match[5]   // version from variable name

		var version string
		if quotedVersion != "" {
			version = quotedVersion
		} else if bareVersion != "" {
			version = resolveVersion(bareVersion, vars)
		} else {
			version = "latest"
		}

		// Build package key for duplicate detection
		pkgKey := groupId + ":" + artifactId
		if seen[pkgKey] {
			continue
		}
		seen[pkgKey] = true

		// Calculate location
		startIdx, endIdx := computeLocationIndices(rawLine, groupId)

		packages = append(packages, models.Package{
			PackageManager: "sbt",
			PackageName:    pkgKey,
			Version:        version,
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       lineNum,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}},
		})
	}

	return packages, nil
}
