// Package carthage parses Carthage manifest files (Cartfile, Cartfile.private, Cartfile.resolved).
package carthage

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

const packageManagerName = "carthage"

// CarthageParser parses Carthage manifest files.
type CarthageParser struct{}

var originLine = regexp.MustCompile(`^\s*(github|git|binary)\s+"([^"]+)"\s*(.*)$`)

// Parse implements the Parser interface.
// If a sibling Cartfile.resolved exists, exact versions from the lock file override ranges.
func (p *CarthageParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}

	// Load lock file versions if present.
	lockVersions := loadCartfileLockVersions(manifestFile)

	var packages []models.Package
	for i, raw := range lines {
		code := stripInlineComment(raw)
		if strings.TrimSpace(code) == "" {
			continue
		}

		match := originLine.FindStringSubmatch(code)
		if match == nil {
			continue
		}
		origin, source, spec := match[1], match[2], strings.TrimSpace(match[3])

		name := packageNameFromSource(origin, source)
		if name == "" {
			continue
		}

		resolvedVersion := resolveVersion(spec)

		// Override with lock file version only if the manifest version is "latest" (unresolved range).
		// Don't override if the manifest already has an exact version.
		if resolvedVersion == "latest" {
			if lockedVer, ok := lockVersions[name]; ok && lockedVer != "latest" {
				resolvedVersion = lockedVer
			}
		}

		startIdx := len(raw) - len(strings.TrimLeft(raw, " \t"))
		endIdx := len(strings.TrimRight(code, " \t"))
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        resolvedVersion,
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       i,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}},
		})
	}

	return packages, nil
}

func packageNameFromSource(origin, source string) string {
	if origin == "github" {
		return source
	}
	trimmed := strings.TrimRight(source, "/")
	if i := strings.LastIndexAny(trimmed, "/:"); i >= 0 {
		trimmed = trimmed[i+1:]
	}
	trimmed = strings.TrimSuffix(trimmed, ".git")
	trimmed = strings.TrimSuffix(trimmed, ".json")
	return trimmed
}

func resolveVersion(spec string) string {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "latest"
	}

	if strings.HasPrefix(spec, `"`) && strings.HasSuffix(spec, `"`) && len(spec) >= 2 {
		inner := spec[1 : len(spec)-1]
		if looksLikeSemver(inner) {
			return inner
		}
		return "latest"
	}

	if strings.HasPrefix(spec, "==") {
		v := strings.TrimSpace(strings.TrimPrefix(spec, "=="))
		if looksLikeSemver(v) {
			return v
		}
		return "latest"
	}

	return "latest"
}

var semverLike = regexp.MustCompile(`^v?\d+(\.\d+)*([.-][A-Za-z0-9.+-]+)?$`)

func looksLikeSemver(s string) bool {
	return semverLike.MatchString(s)
}

// parseCartfileLock parses a Cartfile.resolved lock file and returns one Package per line.
// The format is identical to Cartfile: github/git/binary declarations with resolved versions.
func parseCartfileLock(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}

	var packages []models.Package
	for i, raw := range lines {
		code := stripInlineComment(raw)
		if strings.TrimSpace(code) == "" {
			continue
		}

		match := originLine.FindStringSubmatch(code)
		if match == nil {
			continue
		}
		origin, source, spec := match[1], match[2], strings.TrimSpace(match[3])

		name := packageNameFromSource(origin, source)
		if name == "" {
			continue
		}

		startIdx := len(raw) - len(strings.TrimLeft(raw, " \t"))
		endIdx := len(strings.TrimRight(code, " \t"))
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        resolveVersion(spec),
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       i,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}},
		})
	}

	return packages, nil
}

// loadCartfileLockVersions attempts to load a sibling Cartfile.resolved and returns
// a map of package name -> exact version for those found in the lock.
// Returns nil if the lock file is not found or fails to parse.
func loadCartfileLockVersions(manifestFile string) map[string]string {
	lockFile := filepath.Join(filepath.Dir(manifestFile), "Cartfile.resolved")
	lockedPkgs, err := parseCartfileLock(lockFile)
	if err != nil {
		return nil // Lock file not found or invalid; continue with manifest versions
	}

	versionMap := make(map[string]string)
	for _, pkg := range lockedPkgs {
		versionMap[pkg.PackageName] = pkg.Version
	}
	return versionMap
}

func stripInlineComment(line string) string {
	inDouble := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && inDouble && i+1 < len(line):
			i++ // skip escaped char
		case ch == '"':
			inDouble = !inDouble
		case !inDouble && ch == '#':
			return line[:i]
		}
	}
	return line
}
