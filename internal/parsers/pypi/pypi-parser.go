package pypi

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// PypiParser implements parsing of requirements.txt
type PypiParser struct{}

func extractPackageName(line string, re *regexp.Regexp, lineNum int, manifestFile string) (string, bool) {
	if match := re.FindStringSubmatch(line); match != nil {
		return match[1], true
	}
	log.Printf("Skipping line %d in %s: no valid package name found", lineNum, manifestFile)
	return "", false
}

func extractVersion(line string) string {
	var version string
	switch {
	case strings.Contains(line, "=="):
		parts := strings.SplitN(line, "==", 2)
		if len(parts) == 2 {
			version = strings.TrimSpace(parts[1])
			if strings.Contains(version, "*") {
				version = "latest"
			}
		} else {
			version = "latest"
		}
	default:
		version = "latest"
	}
	return version
}

func computeIndices(raw, pkgName string) (int, int) {
	idx := strings.Index(raw, pkgName)
	if idx < 0 {
		idx = strings.Index(raw, strings.TrimLeft(raw, " \t"))
	}
	startCol := idx + 1
	withoutComment := strings.SplitN(raw, "#", 2)[0]
	trimmedLine := strings.TrimRight(withoutComment, " ")
	endCol := len(trimmedLine)
	return startCol, endCol
}

func (p *PypiParser) Parse(manifestFile string) ([]models.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []models.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0

	re := regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+)(?:\[.*\])?(?:[>=<!~,\s].*)?$`)

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "#") {
			line = strings.SplitN(line, "#", 2)[0]
			line = strings.TrimSpace(line)
		}
		if strings.Contains(line, ";") {
			line = strings.SplitN(line, ";", 2)[0]
			line = strings.TrimSpace(line)
		}

		pkgName, ok := extractPackageName(line, re, lineNum, manifestFile)
		if !ok {
			continue
		}
		version := extractVersion(line)
		startCol, endCol := computeIndices(raw, pkgName)

		packages = append(packages, models.Package{
			PackageManager: "pypi",
			PackageName:    pkgName,
			Version:        version,
			FilePath:       manifestFile,
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
