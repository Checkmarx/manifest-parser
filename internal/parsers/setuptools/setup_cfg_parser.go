package setuptools

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type SetupCfgParser struct{}

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

func extractPackageName(line string, re *regexp.Regexp) (string, bool) {
	if match := re.FindStringSubmatch(line); match != nil {
		return match[1], true
	}
	return "", false
}

func computeIndices(raw, pkgName string) (int, int) {
	startIdx := strings.Index(raw, pkgName)
	if startIdx < 0 {
		startIdx = strings.IndexFunc(raw, func(r rune) bool {
			return r != ' ' && r != '\t'
		})
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

func (p *SetupCfgParser) Parse(manifestFile string) ([]models.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		log.Printf("Error: Failed to open %s: %v", manifestFile, err)
		return nil, err
	}
	defer file.Close()

	log.Printf("Debug: Parsing setup.cfg file: %s", manifestFile)

	var packages []models.Package
	scanner := bufio.NewScanner(file)
	lineNum := 0

	var currentSection string
	var currentKey string
	re := regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+)(?:\[.*\])?(?:[>=<!~,\s].*)?$`)

	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			lineNum++
			continue
		}

		// Handle section headers [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			currentKey = ""
			log.Printf("Debug: Found section [%s] at line %d", currentSection, lineNum)
			lineNum++
			continue
		}

		// Handle indented dependency lines (continuation of multi-line value)
		if (strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t")) && currentKey != "" {
			depLine := strings.TrimSpace(raw)

			if depLine == "" || strings.HasPrefix(depLine, "#") {
				lineNum++
				continue
			}

			if strings.Contains(depLine, "#") {
				depLine = strings.SplitN(depLine, "#", 2)[0]
				depLine = strings.TrimSpace(depLine)
			}

			if depLine != "" {
				pkgName, ok := extractPackageName(depLine, re)
				if ok {
					version := extractVersion(depLine)
					startCol, endCol := computeIndices(raw, pkgName)
					log.Printf("Debug: Found dependency %s@%s at line %d in section [%s]", pkgName, version, lineNum, currentSection)
					packages = append(packages, models.Package{
						PackageManager: "pypi",
						PackageName:    pkgName,
						Version:        version,
						FilePath:       manifestFile,
						Locations: []models.Location{{
							Line:       lineNum,
							StartIndex: startCol,
							EndIndex:   endCol,
						}},
					})
				} else {
					log.Printf("Warning: Could not parse package name from line %d: %s", lineNum, depLine)
				}
			}
			lineNum++
			continue
		}

		// Handle key = value lines (only for non-indented lines)
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if key == "install_requires" || key == "setup_requires" || key == "tests_require" {
					currentKey = key
					if value != "" && !strings.HasPrefix(value, "#") {
						depLine := strings.TrimSpace(value)
						if depLine != "" {
							pkgName, ok := extractPackageName(depLine, re)
							if ok {
								version := extractVersion(depLine)
								startCol, endCol := computeIndices(raw, pkgName)
								packages = append(packages, models.Package{
									PackageManager: "pypi",
									PackageName:    pkgName,
									Version:        version,
									FilePath:       manifestFile,
									Locations: []models.Location{{
										Line:       lineNum,
										StartIndex: startCol,
										EndIndex:   endCol,
									}},
								})
							}
						}
					}
				} else if currentSection == "options.extras_require" && key != "" && value == "" {
					currentKey = key
				}
			}
		}

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error: Scanner error while reading %s: %v", manifestFile, err)
		return nil, err
	}

	log.Printf("Debug: Successfully parsed %s, found %d dependencies", manifestFile, len(packages))
	return packages, nil
}
