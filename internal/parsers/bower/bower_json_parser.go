package bower

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type bowerJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type BowerJsonParser struct{}

// findPositions returns the 0-based line number and byte offsets (start, end)
// of a key's entry within a named JSON section.  It uses brace-depth tracking
// to stay section-aware, matching the npm parser approach for identical JSON
// structure.
func findPositions(fileContent, key, sectionName string) (lineStart, startIndex, endIndex int) {
	lines := strings.Split(fileContent, "\n")

	sectionPattern := fmt.Sprintf("\"%s\"", sectionName)
	keyPattern := fmt.Sprintf("\"%s\"", key)

	inSection := false
	braceDepth := 0
	baseBraceDepth := 0

	for i, line := range lines {
		braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

		if !inSection && strings.Contains(line, sectionPattern) {
			inSection = true
			baseBraceDepth = braceDepth
			continue
		}

		if inSection && braceDepth <= baseBraceDepth-1 {
			inSection = false
			continue
		}

		if inSection && strings.Contains(line, keyPattern) {
			startPos := strings.Index(line, keyPattern)
			if startPos < 0 {
				continue
			}

			valueStart := strings.Index(line[startPos:], ":")
			if valueStart < 0 {
				continue
			}
			valueStart += startPos + 1

			for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
				valueStart++
			}

			endPos := valueStart
			if valueStart < len(line) {
				if line[valueStart] == '"' {
					endPos = strings.Index(line[valueStart+1:], "\"")
					if endPos >= 0 {
						endPos += valueStart + 2
					}
				} else {
					endPos = strings.IndexAny(line[valueStart:], ",}\n")
					if endPos >= 0 {
						endPos += valueStart
					}
				}
			}

			if endPos >= 0 && endPos < len(line) && line[endPos] == ',' {
				endPos++
			}

			if endPos < 0 {
				endPos = len(line)
			}

			return i, startPos, endPos
		}
	}
	return 0, 0, 0
}

// resolveVersion returns the version string to store in the Package record.
// Bower has no lock file, so any range or non-semver specifier becomes "latest".
// Empty version strings also resolve to "latest" per the module contract.
func resolveVersion(version string) string {
	if version == "" || isRangeOrNonSemver(version) {
		return "latest"
	}
	return version
}

func isRangeOrNonSemver(version string) bool {
	rangeOrNonSemverPrefixes := []string{
		"^", "~", ">", "<",
		"git+", "git://", "https://", "http://",
		"./", "../",
	}
	for _, prefix := range rangeOrNonSemverPrefixes {
		if strings.HasPrefix(version, prefix) {
			return true
		}
	}
	// wildcard: * anywhere (e.g. "*", "1.*", "1.2.*")
	if strings.Contains(version, "*") {
		return true
	}
	// x-range wildcard as a version component (e.g. "1.x", "1.x.x", "1.2.x")
	// Does NOT match pre-release tags that happen to contain "x" (e.g. "1.0.0-fix").
	if version == "x" || strings.Contains(version, ".x") {
		return true
	}
	// GitHub shorthand ("user/repo" or "user/repo#tag") and URL paths
	if strings.Contains(version, "/") {
		return true
	}
	return false
}

func (p *BowerJsonParser) Parse(manifestFile string) ([]models.Package, error) {
	fileContent, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var bower bowerJSON
	if err := json.Unmarshal(fileContent, &bower); err != nil {
		return nil, fmt.Errorf("failed to parse bower.json: %w", err)
	}

	content := string(fileContent)
	var results []models.Package

	processDeps := func(depMap map[string]string, depType string) {
		for name, version := range depMap {
			lineStart, startIdx, endIdx := findPositions(content, name, depType)
			results = append(results, models.Package{
				PackageManager: "npm",
				PackageName:    name,
				Version:        resolveVersion(version),
				FilePath:       manifestFile,
				Locations: []models.Location{{
					Line:       lineStart,
					StartIndex: startIdx,
					EndIndex:   endIdx,
				}},
			})
		}
	}

	processDeps(bower.Dependencies, "dependencies")
	processDeps(bower.DevDependencies, "devDependencies")

	sort.Slice(results, func(i, j int) bool {
		return results[i].Locations[0].Line < results[j].Locations[0].Line
	})

	return results, nil
}
