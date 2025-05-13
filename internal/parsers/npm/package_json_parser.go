package npm

import (
	"encoding/json"
	"fmt"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
	"os"
	"path/filepath"
	"strings"
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
func findPositions(fileContent string, key string) (lineStart, lineEnd, startIndex, endIndex int) {
	lines := strings.Split(fileContent, "\n")

	keyPattern := fmt.Sprintf("\"%s\"", key)
	for i, line := range lines {
		if strings.Contains(line, keyPattern) {
			// Get character positions within the line
			startPos := strings.Index(line, keyPattern)
			if startPos != -1 {
				lineStart = i + 1 // 1-indexed for line numbers
				lineEnd = i + 1
				startIndex = startPos
				endIndex = startPos + len(keyPattern)
				return
			}
		}
	}
	return 0, 0, 0, 0
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

	// Helper function to resolve the exact version of a package:
	// - Returns the exact version directly if specified in package.json
	// - Looks up in package-lock.json if version contains range specifiers
	// - Falls back to sensible defaults if necessary
	getResolvedVersion := func(name, specVersion string) string {

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

	var results []models.Package

	// Process all dependency types
	processDeps := func(depMap map[string]string, depType string) {
		for name, version := range depMap {
			resolvedVersion := getResolvedVersion(name, version)
			lineStart, lineEnd, startIndex, endIndex := findPositions(string(fileContent), name)

			results = append(results, models.Package{
				PackageManager: "npm",
				PackageName:    name,
				Version:        resolvedVersion,
				Filepath:       manifestFile,
				LineStart:      lineStart,
				LineEnd:        lineEnd,
				StartIndex:     startIndex,
				EndIndex:       endIndex,
			})
		}
	}

	processDeps(pkg.Dependencies, "dependencies")
	processDeps(pkg.DevDependencies, "devDependencies")
	processDeps(pkg.PeerDependencies, "peerDependencies")
	processDeps(pkg.OptionalDependencies, "optionalDependencies")

	return results, nil
}
