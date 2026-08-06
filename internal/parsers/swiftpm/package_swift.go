package swiftpm

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

var (
	urlPattern  = regexp.MustCompile(`url\s*:\s*"([^"]+)"`)
	idPattern   = regexp.MustCompile(`id\s*:\s*"([^"]+)"`)
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
// If a sibling Package.resolved exists, exact versions from the lock file override ranges.
func parsePackageSwift(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	statements := extractPackageStatements(lines)

	// Load lock file versions if present.
	lockVersions := loadLockFileVersions(manifestFile)

	var packages []models.Package
	for _, stmt := range statements {
		pkg := parsePackageStatement(stmt)
		if pkg == nil {
			continue
		}
		pkg.FilePath = manifestFile
		pkg.Locations = computeStatementLocations(stmt.rawLines)

		// Override version with lock file if available (case-insensitive lookup).
		if lockedVer, ok := lockVersions[strings.ToLower(pkg.PackageName)]; ok {
			pkg.Version = lockedVer
		}

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

func parsePackageStatement(stmt packageStatement) *models.Package {
	text := stmt.text

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

func packageNameFromURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	if i := strings.LastIndexAny(url, "/:"); i >= 0 {
		url = url[i+1:]
	}
	url = strings.TrimSuffix(url, ".git")
	return path.Base(url)
}

func extractVersion(text string) string {
	for _, pat := range versionPatterns {
		if m := pat.FindStringSubmatch(text); len(m) > 1 {
			return m[1]
		}
	}
	return "latest"
}

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

func stripLineComment(line string) string {
	inDouble := false
	for i := 0; i < len(line)-1; i++ {
		ch := line[i]
		switch {
		case ch == '\\' && inDouble:
			i++
		case ch == '"':
			inDouble = !inDouble
		case !inDouble && ch == '/' && line[i+1] == '/':
			return line[:i]
		}
	}
	return line
}

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

// loadLockFileVersions attempts to load a sibling Package.resolved and returns
// a map of package name -> exact version for those found in the lock.
// Package names are stored in lowercase for case-insensitive lookup.
// Returns nil if the lock file is not found or fails to parse.
func loadLockFileVersions(manifestFile string) map[string]string {
	lockFile := filepath.Join(filepath.Dir(manifestFile), "Package.resolved")
	lockedPkgs, err := parseResolved(lockFile)
	if err != nil {
		return nil // Lock file not found or invalid; continue with manifest versions
	}

	versionMap := make(map[string]string)
	for _, pkg := range lockedPkgs {
		versionMap[strings.ToLower(pkg.PackageName)] = pkg.Version
	}
	return versionMap
}
