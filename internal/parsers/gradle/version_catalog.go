package gradle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// VersionCatalogParser implements parsing of Gradle version catalogs (libs.versions.toml)
type VersionCatalogParser struct{}

// Parse implements the Parser interface for version catalog files
func (p *VersionCatalogParser) Parse(manifestFile string) ([]models.Package, error) {
	catalog := parseVersionCatalog(manifestFile)
	if catalog == nil {
		return nil, fmt.Errorf("failed to parse version catalog: %w", fmt.Errorf("invalid TOML format"))
	}

	var packages []models.Package

	// Convert catalog libraries to packages
	for _, lib := range catalog.Libraries {
		if lib.Group != "" && lib.Name != "" {
			version := lib.Version
			if version == "" {
				version = "latest"
			}
			packages = append(packages, models.Package{
				PackageManager: "gradle",
				PackageName:    lib.Group + ":" + lib.Name,
				Version:        version,
				FilePath:       manifestFile,
				Locations: []models.Location{{
					Line:       lib.Line,
					StartIndex: lib.StartIndex,
					EndIndex:   lib.EndIndex,
				}},
			})
		}
	}

	return packages, nil
}

// VersionCatalog represents a parsed Gradle version catalog (libs.versions.toml)
type VersionCatalog struct {
	Versions  map[string]string
	Libraries map[string]CatalogLibrary
}

// CatalogLibrary represents a library entry in the version catalog
type CatalogLibrary struct {
	Group      string
	Name       string
	Version    string
	Line       int // 0-based line number in the TOML file
	StartIndex int // offset of first non-whitespace character on the line
	EndIndex   int // offset just past the last non-whitespace character on the line
}

// findVersionCatalog locates gradle/libs.versions.toml relative to the project root
func findVersionCatalog(manifestFile string) string {
	projectRoot := findProjectRoot(filepath.Dir(manifestFile))
	catalogPath := filepath.Join(projectRoot, "gradle", "libs.versions.toml")
	if _, err := os.Stat(catalogPath); err == nil {
		return catalogPath
	}
	return ""
}

// parseVersionCatalog reads and parses a libs.versions.toml file
func parseVersionCatalog(path string) *VersionCatalog {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	catalog := &VersionCatalog{
		Versions:  make(map[string]string),
		Libraries: make(map[string]CatalogLibrary),
	}

	lines := strings.Split(string(content), "\n")
	// Strip trailing \r so byte offsets are consistent on CRLF files
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	currentSection := ""

	sectionPattern := regexp.MustCompile(`^\s*\[(\w+)\]\s*$`)
	simpleKV := regexp.MustCompile(`^\s*([^\s=]+)\s*=\s*"([^"]+)"\s*$`)

	for lineIdx, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if match := sectionPattern.FindStringSubmatch(trimmed); len(match) > 1 {
			currentSection = match[1]
			continue
		}

		switch currentSection {
		case "versions":
			if match := simpleKV.FindStringSubmatch(trimmed); len(match) > 2 {
				catalog.Versions[match[1]] = match[2]
			}
		case "libraries":
			parseCatalogLibraryEntry(trimmed, raw, lineIdx, catalog)
		}
	}

	// Resolve version.ref references
	for key, lib := range catalog.Libraries {
		if strings.HasPrefix(lib.Version, "ref:") {
			refName := strings.TrimPrefix(lib.Version, "ref:")
			if resolved, ok := catalog.Versions[refName]; ok {
				lib.Version = resolved
				catalog.Libraries[key] = lib
			}
		}
	}

	return catalog
}

// parseCatalogLibraryEntry parses a single library line from the version catalog.
// trimmed is the whitespace-stripped line content used for regex matching;
// raw is the original line used to compute byte offsets for Location indices.
func parseCatalogLibraryEntry(trimmed, raw string, lineIdx int, catalog *VersionCatalog) {
	startIdx, endIdx := lineExtent(raw)

	// Pattern: key = "group:name:version"
	simplePattern := regexp.MustCompile(`^\s*([^\s=]+)\s*=\s*"([^"]+)"\s*$`)
	if match := simplePattern.FindStringSubmatch(trimmed); len(match) > 2 {
		parts := strings.Split(match[2], ":")
		if len(parts) >= 2 {
			lib := CatalogLibrary{
				Group:      parts[0],
				Name:       parts[1],
				Line:       lineIdx,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}
			if len(parts) >= 3 {
				lib.Version = parts[2]
			}
			catalog.Libraries[match[1]] = lib
			return
		}
	}

	// Pattern: key = { module = "group:name", version.ref = "xxx" }
	// Pattern: key = { module = "group:name", version = "xxx" }
	// Pattern: key = { group = "g", name = "n", version.ref = "xxx" }
	// Pattern: key = { group = "g", name = "n", version = "xxx" }
	kvPattern := regexp.MustCompile(`^\s*([^\s=]+)\s*=\s*\{(.+)\}\s*$`)
	if match := kvPattern.FindStringSubmatch(trimmed); len(match) > 2 {
		key := match[1]
		body := match[2]

		lib := CatalogLibrary{}

		// Extract module = "group:name"
		modulePattern := regexp.MustCompile(`module\s*=\s*"([^"]+)"`)
		if m := modulePattern.FindStringSubmatch(body); len(m) > 1 {
			parts := strings.Split(m[1], ":")
			if len(parts) >= 2 {
				lib.Group = parts[0]
				lib.Name = parts[1]
			}
		}

		// Extract group/name separately
		groupPattern := regexp.MustCompile(`group\s*=\s*"([^"]+)"`)
		namePattern := regexp.MustCompile(`name\s*=\s*"([^"]+)"`)
		if m := groupPattern.FindStringSubmatch(body); len(m) > 1 {
			lib.Group = m[1]
		}
		if m := namePattern.FindStringSubmatch(body); len(m) > 1 {
			lib.Name = m[1]
		}

		// Extract version.ref or version
		versionRefPattern := regexp.MustCompile(`version\.ref\s*=\s*"([^"]+)"`)
		versionPattern := regexp.MustCompile(`(?:^|[^.])version\s*=\s*"([^"]+)"`)
		if m := versionRefPattern.FindStringSubmatch(body); len(m) > 1 {
			lib.Version = "ref:" + m[1]
		} else if m := versionPattern.FindStringSubmatch(body); len(m) > 1 {
			lib.Version = m[1]
		}

		if lib.Group != "" && lib.Name != "" {
			lib.Line = lineIdx
			lib.StartIndex = startIdx
			lib.EndIndex = endIdx
			catalog.Libraries[key] = lib
		}
	}
}

// lineExtent returns the offset of the first non-whitespace char and the offset
// just past the last non-whitespace char on the line.
func lineExtent(line string) (int, int) {
	startIdx := len(line) - len(strings.TrimLeft(line, " \t"))
	endIdx := len(strings.TrimRight(line, " \t"))
	return startIdx, endIdx
}

// catalogKeyToDependency resolves a version catalog accessor (e.g., "spring.core")
// to a library entry. In Gradle, dots in the accessor map to dashes in catalog keys.
func catalogKeyToDependency(ref string, catalog *VersionCatalog) *CatalogLibrary {
	if catalog == nil {
		return nil
	}

	// In Gradle, dots in accessor map to dashes in catalog keys
	// e.g., libs.spring.core -> spring-core
	catalogKey := strings.ReplaceAll(ref, ".", "-")

	if lib, ok := catalog.Libraries[catalogKey]; ok {
		return &lib
	}

	return nil
}

// parseVersionCatalogDependencies extracts dependencies from version catalog references in content
func parseVersionCatalogDependencies(content string, catalog *VersionCatalog) []models.Package {
	if catalog == nil {
		return nil
	}

	var packages []models.Package

	// Match patterns like:
	// implementation(libs.spring.core)
	// implementation libs.spring.core
	configPattern := `(?i)\b(` + configKeywords + `)\s*(?:\(\s*)?libs\.([a-zA-Z0-9.]+)\s*\)?`
	pattern := regexp.MustCompile(configPattern)

	lines := strings.Split(content, "\n")
	// Strip trailing \r so byte offsets are consistent on CRLF files
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}

		matches := pattern.FindAllStringSubmatch(trimmed, -1)
		for _, match := range matches {
			if len(match) > 2 {
				ref := match[2]
				lib := catalogKeyToDependency(ref, catalog)
				if lib != nil && lib.Group != "" && lib.Name != "" {
					startIdx, endIdx := lineExtent(stripInlineComment(raw))
					version := lib.Version
					if version == "" {
						version = "latest"
					}
					packages = append(packages, models.Package{
						PackageManager: "gradle",
						PackageName:    lib.Group + ":" + lib.Name,
						Version:        version,
						Locations: []models.Location{{
							Line:       i,
							StartIndex: startIdx,
							EndIndex:   endIdx,
						}},
					})
				}
			}
		}
	}

	return packages
}
