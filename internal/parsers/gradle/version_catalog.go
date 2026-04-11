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
	lineNum := 1
	for _, lib := range catalog.Libraries {
		if lib.Group != "" && lib.Name != "" {
			packages = append(packages, models.Package{
				PackageManager: "gradle",
				PackageName:    lib.Group + ":" + lib.Name,
				Version:        lib.Version,
				FilePath:       manifestFile,
				Locations:      []models.Location{{Line: lineNum}},
			})
			lineNum++
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
	Group   string
	Name    string
	Version string
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
	currentSection := ""

	sectionPattern := regexp.MustCompile(`^\s*\[(\w+)\]\s*$`)
	simpleKV := regexp.MustCompile(`^\s*([^\s=]+)\s*=\s*"([^"]+)"\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if match := sectionPattern.FindStringSubmatch(line); len(match) > 1 {
			currentSection = match[1]
			continue
		}

		switch currentSection {
		case "versions":
			if match := simpleKV.FindStringSubmatch(line); len(match) > 2 {
				catalog.Versions[match[1]] = match[2]
			}
		case "libraries":
			parseCatalogLibraryEntry(line, catalog)
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

// parseCatalogLibraryEntry parses a single library line from the version catalog
func parseCatalogLibraryEntry(line string, catalog *VersionCatalog) {
	// Pattern: key = "group:name:version"
	simplePattern := regexp.MustCompile(`^\s*([^\s=]+)\s*=\s*"([^"]+)"\s*$`)
	if match := simplePattern.FindStringSubmatch(line); len(match) > 2 {
		parts := strings.Split(match[2], ":")
		if len(parts) >= 2 {
			lib := CatalogLibrary{
				Group: parts[0],
				Name:  parts[1],
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
	if match := kvPattern.FindStringSubmatch(line); len(match) > 2 {
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
			catalog.Libraries[key] = lib
		}
	}
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
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 2 {
				ref := match[2]
				lib := catalogKeyToDependency(ref, catalog)
				if lib != nil && lib.Group != "" && lib.Name != "" {
					packages = append(packages, models.Package{
						PackageManager: "gradle",
						PackageName:    lib.Group + ":" + lib.Name,
						Version:        lib.Version,
						Locations:      []models.Location{{Line: i + 1}},
					})
				}
			}
		}
	}

	return packages
}
