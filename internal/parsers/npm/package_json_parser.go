package npm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// Full package.json structure to capture all dependency types
type packageJSON struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

// Modern package-lock.json structure (supporting both v1 and v2 formats)
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
}

// NpmParser extracts packages with position information from package.json
type NpmPackageJsonParser struct{}

// Extract line and character positions for a key in JSON
func findPositions(fileContent string, key string) (lineStart, startIndex, endIndex int) {
	lines := strings.Split(fileContent, "\n")

	keyPattern := fmt.Sprintf("\"%s\"", key)
	for i, line := range lines {
		if strings.Contains(line, keyPattern) {
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
			if endPos < len(line) && line[endPos] == ',' {
				endPos++
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

	// Try to load package-lock.json
	lockPath := filepath.Join(filepath.Dir(manifestFile), "package-lock.json")
	var lock lockFile
	lockContent, err := os.ReadFile(lockPath)
	if err == nil {
		if err := json.Unmarshal(lockContent, &lock); err != nil {
			// Just log and continue - we'll use specified versions if lock parsing fails
			fmt.Printf("Warning: could not parse package-lock.json: %v\n", err)
		}
	}

	var results []models.Package

	// Process all dependency types
	processDeps := func(depMap map[string]string, depType string) {
		for name, version := range depMap {
			resolvedVersion := getResolvedVersion(name, version, lock)
			lineStart, startIndex, endIndex := findPositions(string(fileContent), name)

			results = append(results, models.Package{
				PackageManager: "npm",
				PackageName:    name,
				Version:        resolvedVersion,
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

// - Returns the exact version directly if specified in package.json
// - Looks up in package-lock.json if version contains range specifiers
// - Falls back to sensible defaults if necessary
func getResolvedVersion(name, specVersion string, lock lockFile) string {
	// Check if version is already exact - if so, return it directly
	if !strings.HasPrefix(specVersion, "^") &&
		!strings.HasPrefix(specVersion, "~") &&
		!strings.Contains(specVersion, "*") &&
		!strings.Contains(specVersion, ">") &&
		!strings.Contains(specVersion, "<") &&
		!strings.Contains(specVersion, "latest") {
		return specVersion
	}

	// Try v1 format first
	if deps := lock.Dependencies; deps != nil {
		if entry, ok := deps[name]; ok && entry.Version != "" {
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
			"", // Root package
		}

		for _, path := range pathVariations {
			if entry, ok := pkgs[path]; ok && entry.Version != "" {
				return entry.Version
			}
		}
	}

	// For version specifiers, return "latest" as fallback
	if strings.HasPrefix(specVersion, "^") ||
		strings.HasPrefix(specVersion, "~") ||
		strings.Contains(specVersion, "*") ||
		strings.Contains(specVersion, ">") ||
		strings.Contains(specVersion, "<") {
		return "latest"
	}

	// Otherwise return the specified version
	return specVersion
}
