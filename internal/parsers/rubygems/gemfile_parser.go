package rubygems

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// GemfileParser extracts packages from a Gemfile
type GemfileParser struct{}

// lockFileEntry represents a parsed gem entry from Gemfile.lock
type lockFileEntry struct {
	name    string
	version string
}

// Parse reads a Gemfile and extracts gem declarations with optional Gemfile.lock resolution
func (p *GemfileParser) Parse(manifestFile string) ([]models.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}
	defer file.Close()


	var packages []models.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex to match: gem 'name' or gem "name" with optional version
	// gem 'rails', '~> 6.0'  or  gem 'pg', '>= 0.18', '< 2.0'
	gemRegex := regexp.MustCompile(`^\s*gem\s+['"]([^'"]+)['"]\s*(?:,\s*['"]([^'"]+)['"])?`)

	// Try to load Gemfile.lock for version resolution
	lockMap := p.loadGemfileLock(manifestFile)

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		// Skip empty lines and pure comments
		if line == "" || strings.HasPrefix(line, "#") {
			lineNum++
			continue
		}

		// Remove inline comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		// Match gem declaration
		matches := gemRegex.FindStringSubmatch(line)
		if matches == nil {
			lineNum++
			continue
		}

		pkgName := matches[1]
		versionSpec := ""
		if len(matches) > 2 && matches[2] != "" {
			versionSpec = matches[2]
		}

		// Resolve version: check lock file first if version is a range specifier
		version := p.resolveVersion(pkgName, versionSpec, lockMap)

		// Calculate positions
		startIdx := strings.Index(raw, pkgName)
		endIdx := startIdx + len(pkgName)

		// If there's a version spec, extend endIdx to cover the version string
		if versionSpec != "" {
			versionStart := strings.Index(raw, versionSpec)
			if versionStart >= 0 {
				endIdx = versionStart + len(versionSpec)
			}
		}

		packages = append(packages, models.Package{
			PackageManager: "rubygems",
			PackageName:    pkgName,
			Version:        version,
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       lineNum,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}},
		})
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading manifest file: %w", err)
	}

	return packages, nil
}

// resolveVersion determines the final version for a package
// Returns exact version if specified, looks up in lock file for range specifiers,
// falls back to "latest" if lock file doesn't have the package
func (p *GemfileParser) resolveVersion(pkgName, versionSpec string, lockMap map[string]string) string {
	// If no version specified, look in lock file or default to "latest"
	if versionSpec == "" {
		if lockVersion, ok := lockMap[pkgName]; ok {
			return lockVersion
		}
		return "latest"
	}

	// If version is exact (no range specifiers), return it directly
	if !p.hasRangeSpecifiers(versionSpec) {
		return versionSpec
	}

	// Version has range specifiers, try to get resolved version from lock file
	if lockVersion, ok := lockMap[pkgName]; ok {
		return lockVersion
	}

	// Lock file doesn't have it, return "latest" for range specifier
	return "latest"
}

// hasRangeSpecifiers checks if a version string contains range operators
func (p *GemfileParser) hasRangeSpecifiers(version string) bool {
	rangeOperators := []string{"~>", ">=", "<=", ">", "<", "!=", "~", "^", "*"}
	for _, op := range rangeOperators {
		if strings.Contains(version, op) {
			return true
		}
	}
	return false
}

// loadGemfileLock loads the Gemfile.lock file and returns a map of package name -> version
// loadGemfileLock loads the Gemfile.lock file and returns a map of package name -> version
// Follows npm pattern: silently continues if lock file not present, logs warning if parsing fails
func (p *GemfileParser) loadGemfileLock(manifestFile string) map[string]string {
	lockPath := filepath.Join(filepath.Dir(manifestFile), "Gemfile.lock")
	lockMap := make(map[string]string)

	file, err := os.Open(lockPath)
	if err != nil {
		// Lock file not present, return empty map (same as npm with missing package-lock.json)
		return lockMap
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inGemSection := false
	inSpecsSection := false

	// Regex to match spec lines: "    rails (6.0.4.7)" (4 spaces + name + version in parens)
	specRegex := regexp.MustCompile(`^    ([a-zA-Z0-9_\-\.]+)\s+\(([^)]+)\)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for GEM section start
		if strings.TrimSpace(line) == "GEM" {
			inGemSection = true
			continue
		}

		// Check for specs section start
		if inGemSection && strings.TrimSpace(line) == "specs:" {
			inSpecsSection = true
			continue
		}

		// Exit specs section when we hit a non-indented line after being in specs
		if inSpecsSection && len(line) > 0 && line[0] != ' ' {
			inSpecsSection = false
			inGemSection = false
			break
		}

		// Parse spec lines (4-space indented)
		if inSpecsSection {
			matches := specRegex.FindStringSubmatch(line)
			if matches != nil {
				name := matches[1]
				version := matches[2]
				lockMap[name] = version
			}
		}
	}

	return lockMap
}
