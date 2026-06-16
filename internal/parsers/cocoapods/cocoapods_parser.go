// Package cocoapods parses CocoaPods manifest, lock, and spec files.
//
// Four file formats are supported:
//   - Podfile        — Ruby DSL app manifest (Checkmarx SCA core list)
//   - Podfile.lock   — YAML lock file        (Checkmarx SCA core list)
//   - *.podspec      — Ruby DSL pod spec     (pragmatic addition for pod authors)
//   - *.podspec.json — JSON pod spec         (pragmatic addition for pod authors)
package cocoapods

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

const packageManagerName = "cocoapods"

// CocoaPodsParser dispatches to the right sub-parser based on the file name.
type CocoaPodsParser struct{}

// Parse implements the Parser interface.
func (p *CocoaPodsParser) Parse(manifestFile string) ([]models.Package, error) {
	name := filepath.Base(manifestFile)
	switch {
	case name == "Podfile":
		return parsePodfile(manifestFile)
	case name == "Podfile.lock":
		return parsePodfileLock(manifestFile)
	case strings.HasSuffix(name, ".podspec.json"):
		return parsePodspecJSON(manifestFile)
	case strings.HasSuffix(name, ".podspec"):
		return parsePodspec(manifestFile)
	}
	return nil, fmt.Errorf("unsupported CocoaPods file: %s", name)
}

// errReadFile wraps file-read errors with a consistent prefix.
func errReadFile(err error) error {
	return fmt.Errorf("failed to read manifest file: %w", err)
}

// splitLinesCRLF splits on \n and strips trailing \r so byte offsets stay correct
// on Windows CRLF files.
func splitLinesCRLF(s string) []string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}

// stripInlineComment removes a trailing `# ...` from a Ruby source line, respecting
// quotes so `#` inside a string literal is preserved.
func stripInlineComment(line string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && (inSingle || inDouble) && i+1 < len(line):
			i++ // skip escaped char
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case !inSingle && !inDouble && ch == '#':
			return line[:i]
		}
	}
	return line
}

// lineExtent returns the offset of the first non-whitespace char and the offset
// just past the last non-whitespace char on the (already comment-stripped) line.
func lineExtent(raw, codeOnly string) (int, int) {
	start := len(raw) - len(strings.TrimLeft(raw, " \t"))
	end := len(strings.TrimRight(codeOnly, " \t"))
	return start, end
}

// resolveVersion takes a raw Podfile/podspec version-specifier and returns the
// concrete version, or "latest" for ranges, git refs, or omitted versions.
func resolveVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "latest"
	}
	// Reject anything that starts with an operator like ~>, >=, <, <=, !=, =
	for _, op := range []string{"~>", ">=", "<=", "!=", ">", "<", "="} {
		if strings.HasPrefix(raw, op) {
			return "latest"
		}
	}
	return raw
}
