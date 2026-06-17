package composer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type composerJSON struct {
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

type lockPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type composerLock struct {
	Packages    []lockPackage `json:"packages"`
	PackagesDev []lockPackage `json:"packages-dev"`
}

type ComposerJsonParser struct{}

func isSkippedPackage(name string) bool {
	return name == "php" ||
		strings.HasPrefix(name, "ext-") ||
		strings.HasPrefix(name, "lib-")
}

func findLockVersion(name string, lock composerLock) string {
	for _, pkg := range lock.Packages {
		if pkg.Name == name {
			return pkg.Version
		}
	}
	for _, pkg := range lock.PackagesDev {
		if pkg.Name == name {
			return pkg.Version
		}
	}
	return ""
}

func checkComposerRangeSpecifier(version string) bool {
	return strings.HasPrefix(version, "^") ||
		strings.HasPrefix(version, "~") ||
		strings.Contains(version, "*") ||
		strings.Contains(version, ">") ||
		strings.Contains(version, "<") ||
		strings.Contains(version, "!=") ||
		strings.Contains(version, " - ") ||
		strings.Contains(version, "||") ||
		strings.Contains(version, "latest")
}

func getResolvedVersion(name, specVersion string, lock composerLock) string {
	if !checkComposerRangeSpecifier(specVersion) {
		return specVersion
	}

	if lockVer := findLockVersion(name, lock); lockVer != "" {
		return lockVer
	}

	return "latest"
}

func findPositions(fileContent string, key string, sectionName string) (lineStart, startIndex, endIndex int) {
	lines := strings.Split(fileContent, "\n")

	sectionPattern := fmt.Sprintf("\"%s\"", sectionName)
	inSection := false
	braceDepth := 0
	baseBraceDepth := 0

	keyPattern := fmt.Sprintf("\"%s\"", key)
	for i, line := range lines {
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")
		braceDepth += openBraces - closeBraces

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

			endPos := startPos
			valueStart := strings.Index(line[startPos:], ":")
			if valueStart < 0 {
				continue
			}
			valueStart += startPos + 1

			for valueStart < len(line) && (line[valueStart] == ' ' || line[valueStart] == '\t') {
				valueStart++
			}

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

			lineStart = i
			startIndex = startPos
			endIndex = endPos
			return
		}
	}
	return 0, 0, 0
}

func (p *ComposerJsonParser) Parse(manifestFile string) ([]models.Package, error) {
	fileContent, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest composerJSON
	if err := json.Unmarshal(fileContent, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse composer.json: %w", err)
	}

	lockPath := filepath.Join(filepath.Dir(manifestFile), "composer.lock")
	var lock composerLock
	lockContent, err := os.ReadFile(lockPath)
	if err == nil {
		if err := json.Unmarshal(lockContent, &lock); err != nil {
			fmt.Printf("Warning: could not parse composer.lock: %v\n", err)
		}
	}

	var results []models.Package

	processDeps := func(depMap map[string]string, sectionName string) {
		for name, version := range depMap {
			if isSkippedPackage(name) {
				continue
			}
			resolvedVersion := getResolvedVersion(name, version, lock)
			lineStart, startIndex, endIndex := findPositions(string(fileContent), name, sectionName)

			results = append(results, models.Package{
				PackageManager: "packagist",
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

	processDeps(manifest.Require, "require")
	processDeps(manifest.RequireDev, "require-dev")

	sort.Slice(results, func(i, j int) bool {
		return results[i].Locations[0].Line < results[j].Locations[0].Line
	})

	return results, nil
}
