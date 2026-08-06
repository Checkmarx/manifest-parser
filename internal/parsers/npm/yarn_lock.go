package npm

import (
	"os"
	"regexp"
	"strings"
)

// Matches Yarn 1 (`  version "x.y.z"`) and Yarn 2 (`  version: x.y.z`).
var versionLinePattern = regexp.MustCompile(`^\s+version\s*:?\s*"?([^"\s]+)"?`)

// loadYarnLockVersions returns a name → resolved-version map from a yarn.lock,
// used as a fallback for ranged versions in package.json when package-lock.json is absent.
func loadYarnLockVersions(yarnLockPath string) (map[string]string, error) {
	content, err := os.ReadFile(yarnLockPath)
	if err != nil {
		return nil, err
	}

	lines := splitLinesCRLF(string(content))
	versions := make(map[string]string)
	var pendingName string

	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !startsWithWhitespace(raw) && strings.HasSuffix(trimmed, ":") {
			pendingName = ""
			if trimmed == "__metadata:" {
				continue
			}
			specs := strings.Split(strings.TrimSuffix(trimmed, ":"), ",")
			name := packageNameFromYarnSpec(specs[0])
			if name == "" {
				continue
			}
			pendingName = name
			continue
		}

		if pendingName == "" {
			continue
		}

		match := versionLinePattern.FindStringSubmatch(raw)
		if match == nil {
			continue
		}
		versions[pendingName] = match[1]
		pendingName = ""
	}

	return versions, nil
}

// packageNameFromYarnSpec strips quotes and the version selector from a yarn.lock
// entry header. Handles scoped packages (@scope/name) and Yarn 2 protocol prefixes (npm:, workspace:).
func packageNameFromYarnSpec(spec string) string {
	spec = strings.TrimSpace(strings.Trim(strings.TrimSpace(spec), `"`))
	if spec == "" {
		return ""
	}
	if strings.HasPrefix(spec, "@") {
		if idx := strings.Index(spec[1:], "@"); idx >= 0 {
			return spec[:idx+1]
		}
		return spec
	}
	if idx := strings.Index(spec, "@"); idx >= 0 {
		return spec[:idx]
	}
	return spec
}

func startsWithWhitespace(s string) bool {
	return len(s) > 0 && (s[0] == ' ' || s[0] == '\t')
}

func splitLinesCRLF(s string) []string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}
