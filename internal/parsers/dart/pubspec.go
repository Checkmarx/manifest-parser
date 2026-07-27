package dart

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// pubspec.yaml declares dependencies under three sections. Recognised value shapes:
//
//   dependencies:
//     http: ^0.13.5              # caret constraint
//     provider: 6.0.5            # exact version
//     path: '>=1.8.0 <2.0.0'     # ranged -> resolved from lock, else "latest"
//     collection: any            # any -> resolved from lock, else "latest"
//     flutter:                   # SDK dep - skipped (not a pub.dev package)
//       sdk: flutter
//     my_git:                    # git source map, version from lock / "latest"
//       git:
//         url: https://github.com/example/pkg.git
//         ref: main
//     my_hosted:                 # hosted source map with explicit version
//       hosted: https://my-pub.example.com
//       version: ^1.0.0
//   dev_dependencies:
//     test: ^1.21.0
//   dependency_overrides:
//     collection: 1.17.0
//
// We line-scan to record exact source locations. The indent of dependency entries
// is detected per section (the first child line), so any consistent indent width
// works; source-map details nested deeper than that are attributed to their parent.

var (
	sectionHeaderPattern = regexp.MustCompile(`^([A-Za-z_]+):\s*$`)
	// depEntryPattern matches "name:" or "name: value" on an already-dedented line.
	depEntryPattern = regexp.MustCompile(`^([A-Za-z0-9_.-]+):(.*)$`)
)

var dependencySections = map[string]bool{
	"dependencies":         true,
	"dev_dependencies":     true,
	"dependency_overrides": true,
}

// indentWidth returns the number of leading spaces on a line.
func indentWidth(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

func parsePubspec(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	lockVersions := loadLockVersions(filepath.Join(filepath.Dir(manifestFile), "pubspec.lock"))

	var packages []models.Package
	currentSection := ""
	baseIndent := -1 // indent of dependency entries within the current section

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		if strings.TrimSpace(stripYAMLComment(raw)) == "" {
			continue
		}
		indent := indentWidth(raw)

		// Top-level key (no indentation) switches section context.
		if indent == 0 {
			currentSection = ""
			baseIndent = -1
			if m := sectionHeaderPattern.FindStringSubmatch(stripYAMLComment(raw)); m != nil && dependencySections[m[1]] {
				currentSection = m[1]
			}
			continue
		}
		if currentSection == "" {
			continue
		}

		// The first child line establishes the dependency-entry indent for the section.
		if baseIndent == -1 {
			baseIndent = indent
		}
		if indent != baseIndent {
			continue // deeper = nested source map (handled via look-ahead); shallower = stray
		}

		m := depEntryPattern.FindStringSubmatch(strings.TrimLeft(stripYAMLComment(raw), " "))
		if m == nil {
			continue
		}
		name := m[1]
		constraint := strings.TrimSpace(m[2])

		if constraint == "" {
			// Nested source map: classify by peeking at the indented block.
			block := gatherBlock(lines, i, baseIndent)
			if blockHasKey(block, "sdk") {
				continue // SDK-sourced dep (flutter, flutter_test, ...) - not on pub.dev
			}
			constraint = blockInlineVersion(block) // "" for git/path without a version
		}

		start, end := lineExtent(raw)
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        resolveVersionWithLock(name, constraint, lockVersions),
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       i,
				StartIndex: start,
				EndIndex:   end,
			}},
		})
	}

	return packages, nil
}

// gatherBlock returns the lines nested under the dependency at index i (indent
// greater than baseIndent), stopping at the next dependency or section header.
func gatherBlock(lines []string, i, baseIndent int) []string {
	var block []string
	for j := i + 1; j < len(lines); j++ {
		l := lines[j]
		if strings.TrimSpace(l) == "" {
			continue
		}
		if indentWidth(l) <= baseIndent {
			break
		}
		block = append(block, l)
	}
	return block
}

// blockHasKey reports whether any line in the block is a mapping key `<key>:`.
func blockHasKey(block []string, key string) bool {
	for _, l := range block {
		if strings.HasPrefix(strings.TrimSpace(stripYAMLComment(l)), key+":") {
			return true
		}
	}
	return false
}

// blockInlineVersion returns the value of a `version:` key within the block, or "".
func blockInlineVersion(block []string) string {
	for _, l := range block {
		t := strings.TrimSpace(stripYAMLComment(l))
		if strings.HasPrefix(t, "version:") {
			return strings.TrimSpace(strings.TrimPrefix(t, "version:"))
		}
	}
	return ""
}
