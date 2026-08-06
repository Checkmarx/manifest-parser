// Package dart parses Dart / Flutter pub manifest and lock files.
package dart

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// packageManagerName is the ecosystem identifier for Dart packages. Dart's
// package manager is pub and its registry is pub.dev.
const packageManagerName = "pub"

// DartParser dispatches to the right sub-parser based on the file name.
type DartParser struct{}

// Parse implements the Parser interface.
func (p *DartParser) Parse(manifestFile string) ([]models.Package, error) {
	name := filepath.Base(manifestFile)
	switch name {
	case "pubspec.yaml":
		return parsePubspec(manifestFile)
	case "pubspec.lock":
		return parsePubspecLock(manifestFile)
	}
	return nil, fmt.Errorf("unsupported Dart file: %s", name)
}

func errReadFile(err error) error {
	return fmt.Errorf("failed to read manifest file: %w", err)
}

// splitLinesCRLF splits content into lines, stripping trailing \r so len(line)
// is correct on CRLF (Windows) files.
func splitLinesCRLF(s string) []string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}

// lineExtent returns the 0-based byte offsets of the first non-whitespace char
// and of the end of the trimmed line content.
func lineExtent(raw string) (int, int) {
	start := len(raw) - len(strings.TrimLeft(raw, " \t"))
	end := len(strings.TrimRight(raw, " \t"))
	return start, end
}

// stripYAMLComment removes a trailing YAML comment (` #...`). Per YAML, a comment
// marker must be preceded by whitespace or start the line, so a `#` embedded in a
// value (e.g. a URL fragment) without a preceding space is left intact. Comment
// markers inside quotes are ignored.
func stripYAMLComment(line string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '\'' && !inDouble:
			inSingle = !inSingle
		case c == '"' && !inSingle:
			inDouble = !inDouble
		case c == '#' && !inSingle && !inDouble:
			if i == 0 || line[i-1] == ' ' || line[i-1] == '\t' {
				return line[:i]
			}
		}
	}
	return line
}

// resolveVersion normalises a pubspec version constraint to a concrete version,
// or the literal "latest" for anything ranged, wildcarded, empty, or "any".
func resolveVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "any" {
		return "latest"
	}
	// Ranged / open constraints: ^ ~ > < *, or a multi-range like ">=1.0.0 <2.0.0".
	if strings.ContainsAny(raw, "^~><*") || strings.Contains(raw, " ") {
		return "latest"
	}
	return raw
}

// resolveVersionWithLock resolves an inline constraint, falling back to the
// sibling pubspec.lock (then "latest") when the constraint is not concrete.
func resolveVersionWithLock(name, constraint string, lock map[string]string) string {
	if v := resolveVersion(constraint); v != "latest" {
		return v
	}
	if lv, ok := lock[name]; ok && lv != "" {
		return lv
	}
	return "latest"
}
