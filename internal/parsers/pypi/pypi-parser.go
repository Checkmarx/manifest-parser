package pypi

import (
	"bufio"
	"os"
	"strings"

	"github.com/Checkmarx/manifest-parser/internal"
)

// PypiParser implements parsing of requirements.txt
type PypiParser struct{}

func (p *PypiParser) Parse(manifestFile string) ([]internal.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []internal.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		// skip empty lines
		if line == "" {
			continue
		}
		// Skip comment lines (lines starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Handle inline comments (remove everything after # character)
		if strings.Contains(line, "#") {
			line = strings.SplitN(line, "#", 2)[0]
			line = strings.TrimSpace(line)
		}

		var version string
		switch {
		// exact version
		case strings.Contains(line, "=="):
			parts := strings.SplitN(line, "==", 2)
			version = parts[1]

		// any range or other specifiers
		case strings.ContainsAny(line, "><~,"):
			version = "latest"

		default:
			// not a valid requirement line → skip
			continue
		}

		// extract package name (before any version specifier)
		pkgName := line
		for _, sep := range []string{"==", ">=", "<=", ">", "<", "~=", "!=", ","} {
			if strings.Contains(line, sep) {
				parts := strings.SplitN(line, sep, 2)
				pkgName = parts[0]
				break
			}
		}
		pkgName = strings.TrimSpace(pkgName)

		// compute character positions in the raw line
		idx := strings.Index(raw, pkgName)
		if idx < 0 {
			// fallback to start of trimmed line
			idx = strings.Index(raw, strings.TrimLeft(raw, " \t"))
		}
		// Go uses 0-based offsets; convert to 1-based columns
		startCol := idx + 1
		endCol := startCol + len(pkgName) - 1

		packages = append(packages, internal.Package{
			PackageManager: "pypi",
			PackageName:    pkgName,
			Version:        version,
			Filepath:       manifestFile,
			LineStart:      lineNum,
			LineEnd:        lineNum,
			StartIndex:     startCol,
			EndIndex:       endCol,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return packages, nil
}
