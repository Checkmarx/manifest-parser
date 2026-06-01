package poetry

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// PoetryPyprojectParser parses pyproject.toml Poetry dependency sections.
type PoetryPyprojectParser struct{}

var (
	groupDepSectionRe     = regexp.MustCompile(`^\[tool\.poetry\.group\.[^.]+\.dependencies\]$`)
	inlineTableVersionRe  = regexp.MustCompile(`version\s*=\s*"([^"]*)"`)
	pep621OptDepSectionRe = regexp.MustCompile(`^\[project\.optional-dependencies\]$`)
)

func isPoetryDepsSection(line string) bool {
	return line == "[tool.poetry.dependencies]" ||
		line == "[tool.poetry.dev-dependencies]" ||
		groupDepSectionRe.MatchString(line)
}

func parsePoetryVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "latest"
	}
	if strings.Contains(v, "*") {
		return "latest"
	}
	for _, op := range []string{"^", "~", ">", "<", ",", "!", "=", ";", "~="} {
		if strings.Contains(v, op) {
			return "latest"
		}
	}
	return v
}

func pyprojectLineIndices(raw, pkgName string) (int, int) {
	startIdx := strings.Index(raw, pkgName)
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := len(raw)
	if commentIdx := strings.Index(raw, "#"); commentIdx >= 0 {
		endIdx = commentIdx
	}
	endIdx = strings.LastIndexFunc(raw[:endIdx], func(r rune) bool {
		return r != ' ' && r != '\t'
	}) + 1
	return startIdx, endIdx
}

func parsePyprojectDepLine(line string) (name, version string, ok bool) {
	eqIdx := strings.Index(line, " = ")
	if eqIdx < 0 {
		return "", "", false
	}
	name = strings.TrimSpace(line[:eqIdx])
	if name == "" || name == "python" {
		return "", "", false
	}

	valueStr := strings.TrimSpace(line[eqIdx+3:])

	if strings.HasPrefix(valueStr, "{") {
		if m := inlineTableVersionRe.FindStringSubmatch(valueStr); m != nil {
			version = m[1]
		} else {
			version = "latest"
		}
	} else if len(valueStr) >= 2 && valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"' {
		version = valueStr[1 : len(valueStr)-1]
	} else {
		version = "latest"
	}

	return name, version, true
}

// parseLockFile reads poetry.lock and returns a map of package name to version
func parseLockFile(manifestDir string) map[string]string {
	lockVersions := make(map[string]string)

	lockPath := filepath.Join(manifestDir, "poetry.lock")

	file, err := os.Open(lockPath)
	if err != nil {
		return lockVersions
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentPackageName string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[[package]]") {
			currentPackageName = ""
			continue
		}

		if strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "[[") {
			currentPackageName = ""
			continue
		}

		if strings.HasPrefix(trimmed, "name = ") {
			currentPackageName = strings.TrimSpace(strings.TrimPrefix(trimmed, "name = "))
			currentPackageName = strings.Trim(currentPackageName, "\"")
			continue
		}

		if currentPackageName != "" && strings.HasPrefix(trimmed, "version = ") {
			version := strings.TrimSpace(strings.TrimPrefix(trimmed, "version = "))
			version = strings.Trim(version, "\"")
			lockVersions[currentPackageName] = version
			currentPackageName = ""
			continue
		}
	}

	return lockVersions
}

// resolveVersionWithLock resolves version using poetry.lock if available
func resolveVersionWithLock(pkgName, version string, lockVersions map[string]string) string {
	if !strings.ContainsAny(version, "^~><,!=;*") {
		return version
	}

	if strings.HasPrefix(version, "==") {
		return strings.TrimSpace(version[2:])
	}

	if lockVersion, found := lockVersions[pkgName]; found {
		return lockVersion
	}

	return "latest"
}

func parsePep621Requirement(req string) (name, version string, ok bool) {
	req = strings.TrimSpace(req)
	if strings.HasPrefix(req, "\"") {
		req = strings.TrimPrefix(req, "\"")
	}
	if strings.HasSuffix(req, "\",") {
		req = strings.TrimSuffix(req, "\",")
	} else if strings.HasSuffix(req, "\"") {
		req = strings.TrimSuffix(req, "\"")
	}
	if strings.HasSuffix(req, ",") {
		req = strings.TrimSuffix(req, ",")
	}
	req = strings.TrimSpace(req)
	if req == "" {
		return "", "", false
	}

	for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">", "<", ";"} {
		if idx := strings.Index(req, sep); idx >= 0 {
			name = strings.TrimSpace(req[:idx])
			versionPart := strings.TrimSpace(req[idx+len(sep):])
			if idx2 := strings.Index(versionPart, ";"); idx2 >= 0 {
				versionPart = strings.TrimSpace(versionPart[:idx2])
			}
			version = sep + versionPart
			return name, version, name != ""
		}
	}

	name = strings.TrimSpace(req)
	return name, "latest", name != ""
}

func (p *PoetryPyprojectParser) Parse(manifestFile string) ([]models.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	manifestDir := filepath.Dir(manifestFile)
	lockVersions := parseLockFile(manifestDir)

	var packages []models.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inPoetryDepsSection := false
	inPep621Section := false
	inPep621Array := false
	skipUntilCloseBrace := false

	for scanner.Scan() {
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		if skipUntilCloseBrace {
			if strings.Contains(trimmed, "}") {
				skipUntilCloseBrace = false
			}
			lineNum++
			continue
		}

		if strings.HasPrefix(trimmed, "[") {
			inPoetryDepsSection = isPoetryDepsSection(trimmed)
			inPep621Section = trimmed == "[project]" || pep621OptDepSectionRe.MatchString(trimmed)
			inPep621Array = false
			lineNum++
			continue
		}

		if inPep621Array {
			if strings.TrimSpace(trimmed) == "]" {
				inPep621Array = false
				lineNum++
				continue
			}

			name, version, ok := parsePep621Requirement(trimmed)
			if ok {
				resolvedVersion := resolveVersionWithLock(name, version, lockVersions)
				startIdx, endIdx := pyprojectLineIndices(raw, name)
				packages = append(packages, models.Package{
					PackageManager: "pypi",
					PackageName:    name,
					Version:        resolvedVersion,
					FilePath:       manifestFile,
					Locations: []models.Location{{
						Line:       lineNum,
						StartIndex: startIdx,
						EndIndex:   endIdx,
					}},
				})
			}
			lineNum++
			continue
		}

		if inPoetryDepsSection && (trimmed == "" || strings.HasPrefix(trimmed, "#")) {
			lineNum++
			continue
		}

		if inPoetryDepsSection {
			line := trimmed
			if idx := strings.Index(line, "#"); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}

			name, version, ok := parsePyprojectDepLine(line)
			if !ok {
				lineNum++
				continue
			}

			valueStr := strings.TrimSpace(line[strings.Index(line, " = ")+3:])
			if strings.HasPrefix(valueStr, "{") && !strings.Contains(valueStr, "}") {
				skipUntilCloseBrace = true
			}

			resolvedVersion := resolveVersionWithLock(name, version, lockVersions)

			startIdx, endIdx := pyprojectLineIndices(raw, name)
			packages = append(packages, models.Package{
				PackageManager: "pypi",
				PackageName:    name,
				Version:        resolvedVersion,
				FilePath:       manifestFile,
				Locations: []models.Location{{
					Line:       lineNum,
					StartIndex: startIdx,
					EndIndex:   endIdx,
				}},
			})
			lineNum++
			continue
		}

		if (inPep621Section) && (strings.Contains(trimmed, " = [") || strings.Contains(trimmed, "=[")) && !strings.HasPrefix(trimmed, "[") {
			openIdx := strings.Index(trimmed, "[")
			closeIdx := strings.LastIndex(trimmed, "]")
			if openIdx >= 0 && closeIdx > openIdx {
				arrayContent := trimmed[openIdx+1 : closeIdx]
				parts := strings.Split(arrayContent, ",")
				for _, part := range parts {
					name, version, ok := parsePep621Requirement(part)
					if ok {
						resolvedVersion := resolveVersionWithLock(name, version, lockVersions)
						startIdx, endIdx := pyprojectLineIndices(raw, name)
						packages = append(packages, models.Package{
							PackageManager: "pypi",
							PackageName:    name,
							Version:        resolvedVersion,
							FilePath:       manifestFile,
							Locations: []models.Location{{
								Line:       lineNum,
								StartIndex: startIdx,
								EndIndex:   endIdx,
							}},
						})
					}
				}
			} else if openIdx >= 0 {
				inPep621Array = true
			}
		}

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return packages, nil
}
