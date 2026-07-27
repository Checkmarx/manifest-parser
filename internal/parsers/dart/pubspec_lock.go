package dart

import (
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// pubspec.lock is real YAML, but we line-scan it (rather than unmarshalling) so we
// can record the exact source location of each package entry.
//
//   packages:
//     http:
//       dependency: "direct main"
//       description:
//         name: http
//         url: "https://pub.dev"
//       source: hosted
//       version: "0.13.6"
//   sdks:
//     dart: ">=3.0.0 <4.0.0"
//
// Package names sit at the first indent level under the top-level `packages:` key;
// each package's resolved `version:` sits deeper in the same block. pub always
// resolves a concrete version for every package (including git/path sources).

var lockNamePattern = regexp.MustCompile(`^([A-Za-z0-9_.-]+):\s*$`)

// scanLockPackages walks pubspec.lock line by line and invokes fn once per package
// in the `packages:` section with its name, resolved version ("" if none found),
// and the 0-based line index of the package's name line. The package-entry indent
// is detected from the first entry, so any consistent indent width works.
func scanLockPackages(lines []string, fn func(name, version string, lineIdx int)) {
	inPackages := false
	baseIndent := -1
	curName := ""
	curLine := -1

	flush := func(version string) {
		if curName != "" {
			fn(curName, version, curLine)
			curName = ""
		}
	}

	for i, raw := range lines {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		indent := indentWidth(raw)

		if indent == 0 {
			flush("") // defensive: package with no version line
			inPackages = strings.HasPrefix(strings.TrimSpace(raw), "packages:")
			baseIndent = -1
			continue
		}
		if !inPackages {
			continue
		}
		if baseIndent == -1 {
			baseIndent = indent
		}

		if indent == baseIndent {
			flush("") // defensive: previous package had no version line
			if m := lockNamePattern.FindStringSubmatch(strings.TrimLeft(raw, " ")); m != nil {
				curName = m[1]
				curLine = i
			}
			continue
		}

		// Deeper line: the resolved version belongs to the current package.
		if curName != "" {
			t := strings.TrimSpace(raw)
			if strings.HasPrefix(t, "version:") {
				v := strings.Trim(strings.TrimSpace(strings.TrimPrefix(t, "version:")), `"'`)
				flush(v)
			}
		}
	}
	flush("")
}

func parsePubspecLock(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	var packages []models.Package

	scanLockPackages(lines, func(name, version string, lineIdx int) {
		start, end := lineExtent(lines[lineIdx])
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        resolveVersion(version),
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       lineIdx,
				StartIndex: start,
				EndIndex:   end,
			}},
		})
	})

	return packages, nil
}

// loadLockVersions reads pubspec.lock (if present) and returns a package-name to
// resolved-version map, used by the pubspec.yaml parser to resolve ranged and
// source-map (git/path/hosted) dependencies. Returns nil when the lock is absent.
func loadLockVersions(lockPath string) map[string]string {
	content, err := os.ReadFile(lockPath)
	if err != nil {
		return nil
	}

	versions := map[string]string{}
	scanLockPackages(splitLinesCRLF(string(content)), func(name, version string, _ int) {
		versions[name] = version
	})
	return versions
}
