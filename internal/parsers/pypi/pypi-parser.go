package pypi

import (
	"bufio"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
	"log"
	"os"
	"regexp"
	"strings"
)

// PypiParser implements parsing of requirements.txt
type PypiParser struct{}

func (p *PypiParser) Parse(manifestFile string) ([]models.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []models.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex to match package name, excluding extras
	re := regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+)(?:\[.*\])?(?:[>=<!~,\s].*)?$`)

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

		// Handle inline comments (remove everything after ; character)
		if strings.Contains(line, ";") {
			line = strings.SplitN(line, ";", 2)[0]
			line = strings.TrimSpace(line)
		}

		var version string
		switch {
		// exact version
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
			// any range or other specifiers
			version = "latest"
		}

		// Extract package name using regex
		pkgName := line
		if match := re.FindStringSubmatch(line); match != nil {
			pkgName = match[1] // First capture group is the package name
		} else {
			log.Printf("Skipping line %d in %s: no valid package name found", lineNum, manifestFile)
			continue
		}

		// compute character positions in the raw line
		idx := strings.Index(raw, pkgName)
		if idx < 0 {
			// fallback to start of trimmed line
			idx = strings.Index(raw, strings.TrimLeft(raw, " \t"))
		}
		// Go uses 0-based offsets; convert to 1-based columns
		startCol := idx + 1

		// Calculation of EndIndex
		withoutComment := strings.SplitN(raw, "#", 2)[0]
		// Trim only trailing spaces (right side)
		trimmedLine := strings.TrimRight(withoutComment, " ")
		endCol := len(trimmedLine)

		packages = append(packages, models.Package{
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
