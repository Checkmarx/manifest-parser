package npm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Checkmarx/manifest-parser/internal/versionspec"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// Full package.json structure to capture all dependency types
type packageJSON struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

// Modern package-lock.json structure (supporting both v1 and v2 formats).
type lockFile struct {
	LockfileVersion int `json:"lockfileVersion"`

	// v1 structure
	Dependencies map[string]struct {
		Version string `json:"version"`
	} `json:"dependencies"`

	// v2/v3 structure
	Packages map[string]struct {
		Version      string            `json:"version"`
		Resolved     string            `json:"resolved,omitempty"`
		Integrity    string            `json:"integrity,omitempty"`
		Dependencies map[string]string `json:"dependencies,omitempty"`
	} `json:"packages"`

	// Populated from yarn.lock when package-lock.json is absent.
	YarnVersions map[string]string `json:"-"`
}

// NpmParser extracts packages with position information from package.json
type NpmPackageJsonParser struct {
	// ReportVersionSpecs sets Package.VersionSpec to the declared range when no
	// exact version is known. See parser.WithVersionSpecs.
	ReportVersionSpecs bool
}

// Extract line and character positions for a key in JSON within a specific section
func findPositions(fileContent string, key string, sectionName string) (lineStart, startIndex, endIndex int) {
	lines := strings.Split(fileContent, "\n")

	// First, find the section
	sectionPattern := fmt.Sprintf("\"%s\"", sectionName)
	inSection := false
	braceDepth := 0
	baseBraceDepth := 0

	keyPattern := fmt.Sprintf("\"%s\"", key)
	for i, line := range lines {
		// Count braces to track nesting
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		braceDepth += openBraces - closeBraces

		// Look for the section start
		if !inSection && strings.Contains(line, sectionPattern) {
			inSection = true
			baseBraceDepth = braceDepth
			continue
		}

		// If we're in the section, check if we've exited (brace depth decreased back to or below base)
		if inSection && braceDepth <= baseBraceDepth-1 {
			inSection = false
			continue
		}

		// Only search for the key when we're inside the correct section
		if inSection && strings.Contains(line, keyPattern) {
			// Find the start of the key (after indentation)
			startPos := strings.Index(line, keyPattern)
			if startPos < 0 {
				continue
			}

			// Find the end of the value (including quotes and comma)
			endPos := startPos
			valueStart := strings.Index(line[startPos:], ":")
			if valueStart < 0 {
				continue
			}
			valueStart += startPos + 1

			// Skip whitespace after colon
			for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
				valueStart++
			}

			// Find the end of the value
			if valueStart < len(line) {
				if line[valueStart] == '"' {
					// String value
					endPos = strings.Index(line[valueStart+1:], "\"")
					if endPos >= 0 {
						endPos += valueStart + 2 // +2 for both quotes
					}
				} else {
					// Non-string value (number, boolean, etc.)
					endPos = strings.IndexAny(line[valueStart:], ",}\n")
					if endPos >= 0 {
						endPos += valueStart
					}
				}
			}

			// If we found a comma, include it
			if endPos >= 0 && endPos < len(line) && line[endPos] == ',' {
				endPos++
			}

			// If we couldn't find the end position, fallback to end of line
			if endPos < 0 {
				endPos = len(line)
			}

			lineStart = i
			startIndex = startPos
			endIndex = endPos
			return
		}
	}
	return 0, 0, 0
}

func (p *NpmPackageJsonParser) Parse(manifestFile string) ([]models.Package, error) {
	// Read the entire file for position tracking
	fileContent, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse package.json
	var pkg packageJSON
	if err := json.Unmarshal(fileContent, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Resolve ranged versions from package-lock.json, or yarn.lock if absent.
	dir := filepath.Dir(manifestFile)
	var lock lockFile
	lockContent, err := os.ReadFile(filepath.Join(dir, "package-lock.json"))
	if err == nil {
		if err := json.Unmarshal(lockContent, &lock); err != nil {
			fmt.Printf("Warning: could not parse package-lock.json: %v\n", err)
		}
	} else if yarnVersions, yarnErr := loadYarnLockVersions(filepath.Join(dir, "yarn.lock")); yarnErr == nil {
		lock.YarnVersions = yarnVersions
	}

	var results []models.Package

	// Process all dependency types
	processDeps := func(depMap map[string]string, depType string) {
		for name, version := range depMap {
			_, declaredSpec := versionspec.Classify("npm", version)
			resolvedVersion, versionSpec := versionspec.Result(exactVersion(name, version, lock), declaredSpec, p.ReportVersionSpecs)
			lineStart, startIndex, endIndex := findPositions(string(fileContent), name, depType)

			results = append(results, models.Package{
				PackageManager: "npm",
				PackageName:    name,
				Version:        resolvedVersion,
				VersionSpec:    versionSpec,
				FilePath:       manifestFile,
				Locations: []models.Location{{
					Line:       lineStart,
					StartIndex: startIndex,
					EndIndex:   endIndex,
				}},
			})
		}
	}

	processDeps(pkg.Dependencies, "dependencies")
	processDeps(pkg.DevDependencies, "devDependencies")
	processDeps(pkg.PeerDependencies, "peerDependencies")
	processDeps(pkg.OptionalDependencies, "optionalDependencies")

	// Sort packages by line number
	sort.Slice(results, func(i, j int) bool {
		return results[i].Locations[0].Line < results[j].Locations[0].Line
	})

	return results, nil
}

// getResolvedVersion returns the exact version of a declared npm version, or
// "latest" when there is none (see exactVersion).
func getResolvedVersion(name, specVersion string, lock lockFile) string {
	if version := exactVersion(name, specVersion, lock); version != "" {
		return version
	}

	return versionspec.Latest
}

// exactVersion returns the exact version of a declared npm version: the
// declared one when it is exact, else the one package-lock.json or yarn.lock
// resolves it to. It returns "" when there is none, and never looks a source
// (file:, git+, workspace:, npm: aliases) up in the lock files.
func exactVersion(name, specVersion string, lock lockFile) string {
	exact, spec := versionspec.Classify("npm", specVersion)
	if exact != "" {
		return exact
	}

	if spec == "" {
		return ""
	}

	// Try v1 format first
	if deps := lock.Dependencies; deps != nil {
		if entry, ok := deps[name]; ok && entry.Version != "" && isLockVersionGreater(specVersion, entry.Version) {
			return entry.Version
		}
	}

	// Try v2/v3 format with various path patterns
	if pkgs := lock.Packages; pkgs != nil {
		// Common paths in package-lock.json
		pathVariations := []string{
			"node_modules/" + name,
			"node_modules/" + name + "@" + specVersion,
			"node_modules/" + name + "@" + strings.TrimPrefix(specVersion, "^"),
			"node_modules/" + name + "@" + strings.TrimPrefix(specVersion, "~"),
		}

		for _, path := range pathVariations {
			if entry, ok := pkgs[path]; ok && entry.Version != "" && isLockVersionGreater(specVersion, entry.Version) {
				return entry.Version
			}
		}
	}

	if yv := lock.YarnVersions; yv != nil {
		if version, ok := yv[name]; ok && version != "" && isLockVersionGreater(specVersion, version) {
			return version
		}
	}

	return ""
}
func isLockVersionGreater(specVersion, lockVersion string) bool {
	specVersion = stripRangeSpecifier(specVersion)
	lockVersion = stripRangeSpecifier(lockVersion)

	specParts := strings.Split(specVersion, ".")
	lockParts := strings.Split(lockVersion, ".")

	maxLen := len(specParts)
	if len(lockParts) > maxLen {
		maxLen = len(lockParts)
	}
	for i := 0; i < maxLen; i++ {
		var specPart, lockPart int
		if i < len(specParts) {
			specPart, _ = strconv.Atoi(specParts[i])
		}
		if i < len(lockParts) {
			lockPart, _ = strconv.Atoi(lockParts[i])
		}
		if lockPart == specPart {
			continue
		}
		if lockPart > specPart {
			return true
		} else if lockPart < specPart {
			return false
		}
	}
	// In case if the lock version is same as the spec version, we consider it,
	// as range specifiers indicate version greaterThan or equalTo
	return true
}

func stripRangeSpecifier(version string) string {
	if checkRangeSpecifiersPresent(version) {
		return version[1:]
	}
	return version
}
func checkRangeSpecifiersPresent(version string) bool {
	if strings.HasPrefix(version, "^") ||
		strings.HasPrefix(version, "~") ||
		strings.Contains(version, "*") ||
		strings.Contains(version, ">") ||
		strings.Contains(version, "<") {
		return true
	}
	return false
}
