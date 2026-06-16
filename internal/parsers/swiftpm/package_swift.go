package swiftpm

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// Swift DSL .package(...) forms we need to recognise:
//
//   .package(url: "URL", from: "1.0.0")
//   .package(url: "URL", .upToNextMajor(from: "1.0.0"))
//   .package(url: "URL", .upToNextMinor(from: "1.0.0"))
//   .package(url: "URL", exact: "1.0.0")
//   .package(url: "URL", "1.0.0"..<"2.0.0")          // range -> "latest"
//   .package(url: "URL", branch: "main")             // branch -> "latest"
//   .package(url: "URL", revision: "abc123")         // revision -> "latest"
//   .package(name: "n", url: "URL", from: "1.0.0")   // legacy named form
//   .package(path: "../local")                       // local path -> skipped
//   .package(id: "scope.name", from: "1.0.0")        // registry identifier -> use id as name

var (
	// urlPattern extracts the repository URL from inside a .package(...) call.
	urlPattern = regexp.MustCompile(`url\s*:\s*"([^"]+)"`)
	// idPattern matches the registry-identifier form .package(id: "scope.name", ...).
	idPattern = regexp.MustCompile(`id\s*:\s*"([^"]+)"`)
	// pathPattern detects local-path .package(path: "..."), which we skip.
	pathPattern = regexp.MustCompile(`path\s*:\s*"`)

	versionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`exact\s*:\s*"([^"]+)"`),
		regexp.MustCompile(`from\s*:\s*"([^"]+)"`),
		regexp.MustCompile(`\.upToNextMajor\s*\(\s*from\s*:\s*"([^"]+)"\s*\)`),
		regexp.MustCompile(`\.upToNextMinor\s*\(\s*from\s*:\s*"([^"]+)"\s*\)`),
	}
)

// parsePackageSwift parses a Package.swift manifest using line-oriented scanning.
// Multi-line .package(...) calls are accumulated until parentheses balance, then parsed.
func parsePackageSwift(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	statements := extractPackageStatements(lines)

	var packages []models.Package
	for _, stmt := range statements {
		pkg := parsePackageStatement(stmt)
		if pkg == nil {
			continue
		}
		pkg.FilePath = manifestFile
		pkg.Locations = computeStatementLocations(stmt.rawLines)
		packages = append(packages, *pkg)
	}
	return packages, nil
}

// packageStatement records a .package(...) call gathered from one or more source lines.
type packageStatement struct {
	text     string         // concatenated content of the .package(...) call
	rawLines []rawLineEntry // each contributing source line, in order
}

type rawLineEntry struct {
	line    int
	content string
}

// extractPackageStatements walks the file line by line. When it sees `.package(`
// it accumulates lines until the parentheses balance, producing one packageStatement
// per .package(...) call. Quotes and escape characters are respected.
func extractPackageStatements(lines []string) []packageStatement {
	var statements []packageStatement
	var buf strings.Builder
	var raws []rawLineEntry
	depth := 0
	active := false

	for i, raw := range lines {
		stripped := stripLineComment(raw)
		if !active {
			idx := strings.Index(stripped, ".package(")
			if idx < 0 {
				continue
			}
			active = true
			buf.Reset()
			raws = nil
			buf.WriteString(stripped[idx:])
			raws = append(raws, rawLineEntry{line: i, content: raw})
			depth = parenDelta(stripped[idx:])
			if depth == 0 {
				statements = append(statements, packageStatement{text: buf.String(), rawLines: raws})
				active = false
			}
			continue
		}
		buf.WriteString(" ")
		buf.WriteString(stripped)
		raws = append(raws, rawLineEntry{line: i, content: raw})
		depth += parenDelta(stripped)
		if depth <= 0 {
			statements = append(statements, packageStatement{text: buf.String(), rawLines: raws})
			active = false
		}
	}

	return statements
}

// parsePackageStatement extracts package name and version from one .package(...) statement.
// Returns nil if it's a local-path declaration (which we deliberately skip — no remote dep).
func parsePackageStatement(stmt packageStatement) *models.Package {
	text := stmt.text

	// .package(path: ...) — local path, skip.
	if pathPattern.MatchString(text) && !urlPattern.MatchString(text) && !idPattern.MatchString(text) {
		return nil
	}

	var name string
	if m := urlPattern.FindStringSubmatch(text); len(m) > 1 {
		name = packageNameFromURL(m[1])
	} else if m := idPattern.FindStringSubmatch(text); len(m) > 1 {
		name = m[1]
	}
	if name == "" {
		return nil
	}

	return &models.Package{
		PackageManager: packageManagerName,
		PackageName:    name,
		Version:        extractVersion(text),
	}
}

// packageNameFromURL derives the package name from a SwiftPM repo URL: last path
// component with the .git suffix removed.
//   "https://github.com/apple/swift-nio.git" -> "swift-nio"
//   "git@github.com:apple/swift-log.git"      -> "swift-log"
func packageNameFromURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	if i := strings.LastIndexAny(url, "/:"); i >= 0 {
		url = url[i+1:]
	}
	url = strings.TrimSuffix(url, ".git")
	return path.Base(url)
}

// extractVersion returns the version string for a .package(...) statement, or
// the literal "latest" if the version is ranged, branch-pinned, revision-pinned,
// or otherwise not a concrete semantic version.
func extractVersion(text string) string {
	for _, pat := range versionPatterns {
		if m := pat.FindStringSubmatch(text); len(m) > 1 {
			return m[1]
		}
	}
	return "latest"
}

// computeStatementLocations emits one Location per contributing source line,
// Maven-style: StartIndex = first non-whitespace char, EndIndex = end of code.
func computeStatementLocations(raws []rawLineEntry) []models.Location {
	out := make([]models.Location, 0, len(raws))
	for _, rl := range raws {
		code := strings.TrimRight(stripLineComment(rl.content), " \t")
		if strings.TrimSpace(code) == "" {
			continue
		}
		startIdx := len(rl.content) - len(strings.TrimLeft(rl.content, " \t"))
		out = append(out, models.Location{
			Line:       rl.line,
			StartIndex: startIdx,
			EndIndex:   len(code),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// stripLineComment removes a trailing `// ...` from a Swift source line, respecting
// quotes so `//` inside a string literal is preserved.
func stripLineComment(line string) string {
	inDouble := false
	for i := 0; i < len(line)-1; i++ {
		ch := line[i]
		switch {
		case ch == '\\' && inDouble:
			i++ // skip escaped char
		case ch == '"':
			inDouble = !inDouble
		case !inDouble && ch == '/' && line[i+1] == '/':
			return line[:i]
		}
	}
	return line
}

// parenDelta counts ( minus ) in a line, ignoring parens inside double-quoted strings.
func parenDelta(line string) int {
	delta := 0
	inDouble := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && inDouble && i+1 < len(line):
			i++ // skip escaped char
		case ch == '"':
			inDouble = !inDouble
		case !inDouble && ch == '(':
			delta++
		case !inDouble && ch == ')':
			delta--
		}
	}
	return delta
}
