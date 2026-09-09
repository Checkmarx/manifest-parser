package poetry

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// PoetryPyprojectParser parses pyproject.toml: Poetry's own tables, plus the
// PEP 621/735 sections uv (and other PEP 621 tools) use — uv has no manifest
// format of its own, so it has no dedicated parser.
type PoetryPyprojectParser struct{}

var (
	groupDepSectionRe         = regexp.MustCompile(`^\[tool\.poetry\.group\.[^.]+\.dependencies\]$`)
	inlineTableVersionRe      = regexp.MustCompile(`version\s*=\s*"([^"]*)"`)
	pep621OptDepSectionRe     = regexp.MustCompile(`^\[project\.optional-dependencies\]$`)
	dependencyGroupsSectionRe = regexp.MustCompile(`^\[dependency-groups\]$`)
	pypiNameSepRe             = regexp.MustCompile(`[-_.]+`)
)

func isPoetryDepsSection(line string) bool {
	return line == "[tool.poetry.dependencies]" ||
		line == "[tool.poetry.dev-dependencies]" ||
		groupDepSectionRe.MatchString(line)
}

// isArrayDependencySection reports whether line opens a PEP 621/735 table whose
// entries are arrays of PEP 508 requirement strings.
func isArrayDependencySection(line string) bool {
	return line == "[project]" ||
		pep621OptDepSectionRe.MatchString(line) ||
		dependencyGroupsSectionRe.MatchString(line)
}

// normalizePyPIName applies PEP 503 normalization (lowercase; -/_/. equivalent)
// so lock lookups aren't broken by naming-convention differences.
func normalizePyPIName(name string) string {
	return pypiNameSepRe.ReplaceAllString(strings.ToLower(name), "-")
}

// stripExtras strips a PEP 508 extras suffix ("requests[security]" -> "requests")
// so PackageName matches the real package and the lock file's entry for it.
func stripExtras(name string) string {
	if idx := strings.Index(name, "["); idx >= 0 {
		return strings.TrimSpace(name[:idx])
	}
	return name
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

// parseLockPackageVersions scans [[package]] blocks (poetry.lock and uv.lock
// share this shape) into a PEP-503-normalized name -> version map.
func parseLockPackageVersions(lockPath string) map[string]string {
	lockVersions := make(map[string]string)

	file, err := os.Open(lockPath)
	if err != nil {
		return lockVersions
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentPackageName string

	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(trimmed, "[[package]]") {
			currentPackageName = ""
			continue
		}

		if strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "[[") {
			currentPackageName = ""
			continue
		}

		if strings.HasPrefix(trimmed, "name = ") {
			currentPackageName = normalizePyPIName(strings.Trim(strings.TrimPrefix(trimmed, "name = "), "\""))
			continue
		}

		if currentPackageName != "" && strings.HasPrefix(trimmed, "version = ") {
			version := strings.Trim(strings.TrimPrefix(trimmed, "version = "), "\"")
			lockVersions[currentPackageName] = version
			currentPackageName = ""
			continue
		}
	}

	return lockVersions
}

// loadLockVersions merges poetry.lock and uv.lock; if both exist for a package,
// poetry.lock wins and uv.lock fills in the rest. Neither is ever scanned as a
// standalone manifest — both are resolvers only, like every other lock file here.
func loadLockVersions(manifestDir string) map[string]string {
	lockVersions := parseLockPackageVersions(filepath.Join(manifestDir, "poetry.lock"))
	for name, version := range parseLockPackageVersions(filepath.Join(manifestDir, "uv.lock")) {
		if _, exists := lockVersions[name]; !exists {
			lockVersions[name] = version
		}
	}
	return lockVersions
}

// resolveVersionWithLock resolves version using poetry.lock/uv.lock if available.
// "latest" is the callers' sentinel for "no specifier given", not a real version,
// so it must still go through lock resolution rather than being returned as-is.
func resolveVersionWithLock(pkgName, version string, lockVersions map[string]string) string {
	if version != "latest" && !strings.ContainsAny(version, "^~><,!=;*") {
		return version
	}

	if strings.HasPrefix(version, "===") {
		return strings.TrimSpace(version[3:])
	}
	if strings.HasPrefix(version, "==") {
		// "==1.*" is a wildcard match, not an exact pin, so it still needs the lock.
		if exact := strings.TrimSpace(version[2:]); !strings.Contains(exact, "*") {
			return exact
		}
	}

	if lockVersion, found := lockVersions[normalizePyPIName(pkgName)]; found {
		return lockVersion
	}

	return "latest"
}

func parsePep621Requirement(req string) (name, version string, ok bool) {
	req = strings.TrimSpace(req)
	if strings.HasPrefix(req, "{") {
		// e.g. PEP 735 {include-group = "test"} — not a requirement string; the
		// referenced group is parsed on its own, so skip here to avoid double-counting.
		return "", "", false
	}
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

	// Markers (e.g. "; python_version >= '3.8'") can contain their own comparison
	// operators, so they must be split off before searching for the package's
	// own version specifier — otherwise a marker-only requirement like
	// "tomli; python_version < '3.11'" would have its marker's "<" mistaken for
	// the package's separator, corrupting the name into "tomli; python_version".
	reqNoMarker := req
	if idx := strings.Index(req, ";"); idx >= 0 {
		reqNoMarker = strings.TrimSpace(req[:idx])
	}

	for _, sep := range []string{"===", "==", ">=", "<=", "~=", "!=", ">", "<"} {
		if idx := strings.Index(reqNoMarker, sep); idx >= 0 {
			name = stripExtras(strings.TrimSpace(reqNoMarker[:idx]))
			version = sep + strings.TrimSpace(reqNoMarker[idx+len(sep):])
			return name, version, name != ""
		}
	}

	name = stripExtras(strings.TrimSpace(reqNoMarker))
	return name, "latest", name != ""
}

func (p *PoetryPyprojectParser) Parse(manifestFile string) ([]models.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	manifestDir := filepath.Dir(manifestFile)
	lockVersions := loadLockVersions(manifestDir)

	var packages []models.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inPoetryDepsSection := false
	inArrayDependencySection := false
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
			inArrayDependencySection = isArrayDependencySection(trimmed)
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

			arrayItem := trimmed
			if idx := strings.Index(arrayItem, "#"); idx >= 0 {
				arrayItem = strings.TrimSpace(arrayItem[:idx])
			}

			name, version, ok := parsePep621Requirement(arrayItem)
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

		if inArrayDependencySection && (strings.Contains(trimmed, " = [") || strings.Contains(trimmed, "=[")) && !strings.HasPrefix(trimmed, "[") {
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
