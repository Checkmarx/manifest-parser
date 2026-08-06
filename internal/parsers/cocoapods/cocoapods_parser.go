// Package cocoapods parses CocoaPods manifest and spec files.
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

func errReadFile(err error) error {
	return fmt.Errorf("failed to read manifest file: %w", err)
}

func splitLinesCRLF(s string) []string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}

func stripInlineComment(line string) string {
	inSingle, inDouble := false, false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && (inSingle || inDouble) && i+1 < len(line):
			i++
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

func lineExtent(raw, codeOnly string) (int, int) {
	start := len(raw) - len(strings.TrimLeft(raw, " \t"))
	end := len(strings.TrimRight(codeOnly, " \t"))
	return start, end
}

func resolveVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "latest"
	}
	for _, op := range []string{"~>", ">=", "<=", "!=", ">", "<", "="} {
		if strings.HasPrefix(raw, op) {
			return "latest"
		}
	}
	return raw
}
