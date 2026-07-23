package setuptools

import (
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type SetupPyParser struct{}

type depWithPosition struct {
	name       string
	version    string
	lineNum    int
	startIndex int
	endIndex   int
}

func extractVersionPy(line string) string {
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

func extractPackageNamePy(line string, re *regexp.Regexp) (string, bool) {
	if match := re.FindStringSubmatch(line); match != nil {
		return match[1], true
	}
	return "", false
}

// findPositionInFile finds the exact line number and column position of text in the file
func findPositionInFile(fullText string, depString string, searchStartPos int) (lineNum, startIndex, endIndex int) {
	searchPos := strings.Index(fullText[searchStartPos:], depString)
	if searchPos == -1 {
		return 0, 0, 0
	}

	actualPos := searchStartPos + searchPos

	lineNum = 0
	colPos := 0
	for i := 0; i < actualPos && i < len(fullText); i++ {
		if fullText[i] == '\n' {
			lineNum++
			colPos = 0
		} else {
			colPos++
		}
	}

	startIndex = colPos
	endIndex = colPos + len(depString)

	return lineNum, startIndex, endIndex
}

// extractDepsFromListContent extracts dependencies from list/dict content and returns positions
func extractDepsFromListContent(content string, fullText string, searchStartPos int) []depWithPosition {
	var deps []depWithPosition

	singleQuoteRe := regexp.MustCompile(`'([^']*)'`)
	doubleQuoteRe := regexp.MustCompile(`"([^"]*)`)

	singleMatches := singleQuoteRe.FindAllStringSubmatchIndex(content, -1)
	doubleMatches := doubleQuoteRe.FindAllStringSubmatchIndex(content, -1)

	type match struct {
		dep            string
		startInContent int
		endInContent   int
	}
	var allMatches []match

	for _, m := range singleMatches {
		if len(m) >= 4 {
			dep := content[m[2]:m[3]]
			startPos := m[0]

			endPos := m[1]
			if endPos < len(content) {
				afterQuote := endPos
				for afterQuote < len(content) && (content[afterQuote] == ' ' || content[afterQuote] == '\t') {
					afterQuote++
				}
				if afterQuote < len(content) && content[afterQuote] == ':' {
					continue
				}
			}

			allMatches = append(allMatches, match{
				dep:            dep,
				startInContent: startPos,
				endInContent:   endPos,
			})
		}
	}

	for _, m := range doubleMatches {
		if len(m) >= 4 {
			dep := content[m[2]:m[3]]
			startPos := m[0]

			endPos := m[1]
			if endPos < len(content) {
				afterQuote := endPos
				for afterQuote < len(content) && (content[afterQuote] == ' ' || content[afterQuote] == '\t') {
					afterQuote++
				}
				if afterQuote < len(content) && content[afterQuote] == ':' {
					continue
				}
			}

			allMatches = append(allMatches, match{
				dep:            dep,
				startInContent: startPos,
				endInContent:   endPos,
			})
		}
	}

	pkgNameRe := regexp.MustCompile(`^([a-zA-Z0-9_\-\.]+)(?:\[.*\])?(?:[>=<!~,\s].*)?$`)

	for _, m := range allMatches {
		depLine := strings.TrimSpace(m.dep)
		if depLine == "" || strings.HasPrefix(depLine, "#") {
			continue
		}

		pkgName, ok := extractPackageNamePy(depLine, pkgNameRe)
		if !ok {
			continue
		}

		version := extractVersionPy(depLine)

		lineNum, startIdx, endIdx := findPositionInFile(fullText, m.dep, searchStartPos)

		deps = append(deps, depWithPosition{
			name:       pkgName,
			version:    version,
			lineNum:    lineNum,
			startIndex: startIdx,
			endIndex:   endIdx,
		})
	}

	return deps
}

// extractListContent extracts the content between brackets/braces, handling nesting
func extractListContent(text string, startPos int) string {
	if startPos >= len(text) {
		return ""
	}

	openChar := text[startPos]
	var closeChar byte
	switch openChar {
	case '[':
		closeChar = ']'
	case '(':
		closeChar = ')'
	case '{':
		closeChar = '}'
	default:
		return ""
	}

	depth := 0
	inString := false
	stringChar := byte(0)
	escaped := false

	for i := startPos; i < len(text); i++ {
		ch := text[i]

		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if (ch == '"' || ch == '\'') && !inString {
			inString = true
			stringChar = ch
			continue
		}

		if inString && ch == stringChar {
			inString = false
			continue
		}

		if inString {
			continue
		}

		if ch == openChar {
			depth++
		} else if ch == closeChar {
			depth--
			if depth == 0 {
				return text[startPos+1 : i]
			}
		}
	}

	return ""
}

// extractDependencies extracts dependencies from setup() call text for a specific key
func extractDependencies(setupText string, key string, fullText string, searchStartPos int) []depWithPosition {
	var deps []depWithPosition

	keyPattern := key + "="
	keyIndex := strings.Index(setupText, keyPattern)
	if keyIndex == -1 {
		return deps
	}

	startPos := keyIndex + len(keyPattern)
	for startPos < len(setupText) && (setupText[startPos] == ' ' || setupText[startPos] == '\t') {
		startPos++
	}

	if startPos >= len(setupText) {
		return deps
	}

	content := extractListContent(setupText, startPos)
	if content == "" {
		return deps
	}

	deps = extractDepsFromListContent(content, fullText, searchStartPos)
	return deps
}

func (p *SetupPyParser) Parse(manifestFile string) ([]models.Package, error) {
	data, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, err
	}

	text := string(data)
	var packages []models.Package

	setupStart := strings.Index(text, "setup(")
	if setupStart == -1 {
		setupStart = 0
	} else {
		setupStart += len("setup")
	}

	setupContent := extractListContent(text, setupStart)
	if setupContent == "" && setupStart > 0 {
		setupContent = text[setupStart:]
	}

	for _, key := range []string{"install_requires", "setup_requires", "tests_require"} {
		keyPosInText := strings.Index(text, key+"=")
		deps := extractDependencies(setupContent, key, text, keyPosInText)
		for _, dep := range deps {
			packages = append(packages, models.Package{
				PackageManager: "pypi",
				PackageName:    dep.name,
				Version:        dep.version,
				FilePath:       manifestFile,
				Locations: []models.Location{{
					Line:       dep.lineNum,
					StartIndex: dep.startIndex,
					EndIndex:   dep.endIndex,
				}},
			})
		}
	}

	extrasStart := strings.Index(setupContent, "extras_require")
	if extrasStart != -1 {
		eqIndex := strings.Index(setupContent[extrasStart:], "=")
		if eqIndex != -1 {
			dictStartPos := extrasStart + eqIndex + 1
			for dictStartPos < len(setupContent) && (setupContent[dictStartPos] == ' ' || setupContent[dictStartPos] == '\t' || setupContent[dictStartPos] == '\n') {
				dictStartPos++
			}
			if dictStartPos < len(setupContent) {
				dictContent := extractListContent(setupContent, dictStartPos)
				if dictContent != "" {
					extrasStartInText := strings.Index(text, "extras_require")
					deps := extractDepsFromListContent(dictContent, text, extrasStartInText)
					for _, dep := range deps {
						packages = append(packages, models.Package{
							PackageManager: "pypi",
							PackageName:    dep.name,
							Version:        dep.version,
							FilePath:       manifestFile,
							Locations: []models.Location{{
								Line:       dep.lineNum,
								StartIndex: dep.startIndex,
								EndIndex:   dep.endIndex,
							}},
						})
					}
				}
			}
		}
	}
	return packages, nil
}
