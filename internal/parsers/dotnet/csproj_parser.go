package dotnet

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// DotnetCsprojParser implements parsing of .NET project files (.csproj)
type DotnetCsprojParser struct{}

// PackageReference represents a package reference in the .csproj file
type PackageReference struct {
	Include       string `xml:"Include,attr"`
	VersionAttr   string `xml:"Version,attr"`
	VersionNested string `xml:"Version"`
}

// parseVersion handles version resolution
// - Returns exact version if specified
// - Returns "latest" for version ranges or special version specifiers
func parseVersion(version string) string {
	// Handle empty version
	if version == "" {
		return "latest"
	}

	// If the version contains any kind of brackets, return "latest"
	if strings.ContainsAny(version, "[]()") {
		return "latest"
	}

	// Handle special version specifiers
	if strings.ContainsAny(version, "*^~><") {
		return "latest"
	}

	// Return exact version
	return version
}

// computeIndices calculates start and end indices for PackageReference elements
// Returns startIndex and endIndex for the element in the line
func computeIndices(lines []string, lineNum int) (startIndex, endIndex int, lineStart, lineEnd int) {
	currentLine := lines[lineNum-1] // lineNum is 1-based so we subtract 1

	// Find the position of the PackageReference tag start in the line
	startIdx := strings.Index(currentLine, "<PackageReference")
	if startIdx < 0 {
		return 1, len(currentLine), lineNum, lineNum
	}

	// Check if it's a single-line format
	if strings.Contains(currentLine, "/>") {
		// Single-line format
		endIdx := strings.LastIndex(currentLine, "/>") + 2 // Include the "/>" itself
		return startIdx + 1, endIdx + 1, lineNum, lineNum
	}

	// Multi-line format
	// TODO: Multi-line PackageReference support will be handled in the future.
	// Currently, if the tag spans multiple lines, we only return the first line.
	// The following code is commented out for now:
	/*
		lineEnd = lineNum
		for i := lineNum; i < len(lines) && i < lineNum+10; i++ { // Limit search to 10 lines
			if strings.Contains(lines[i-1], "</PackageReference>") {
				lineEnd = i
				endLine := lines[i-1]
				endIdx := strings.Index(endLine, "</PackageReference>") + len("</PackageReference>")
				return startIdx + 1, endIdx + 1, lineNum, lineEnd
			}
		}
	*/
	// No closing tag found, return the end of the current line
	return startIdx + 1, len(currentLine) + 1, lineNum, lineNum
}

// Parse implements the Parser interface for .csproj files
func (p *DotnetCsprojParser) Parse(manifestFile string) ([]models.Package, error) {
	// Read the file content
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Split content into lines for index computation
	lines := strings.Split(string(content), "\n")

	// Create XML decoder
	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var packages []models.Package

	// Parse XML content
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to parse XML: %w", err)
		}

		// Process each element
		switch elem := token.(type) {
		case xml.StartElement:
			if elem.Name.Local == "PackageReference" {
				var pkgRef PackageReference
				if err := decoder.DecodeElement(&pkgRef, &elem); err != nil {
					return nil, fmt.Errorf("failed to decode PackageReference: %w", err)
				}

				// Skip empty package names
				if pkgRef.Include == "" {
					continue
				}

				// Find line number
				lineNum := 0
				packagePattern := fmt.Sprintf(`PackageReference.*Include="%s"`, pkgRef.Include)
				re := regexp.MustCompile(packagePattern)

				for i, line := range lines {
					if re.MatchString(line) {
						lineNum = i + 1 // 1-indexed line numbers
						break
					}
				}

				// Skip if line not found
				if lineNum == 0 {
					continue
				}

				// Compute indices for both single-line and multi-line formats
				startCol, endCol, lineStart, lineEnd := computeIndices(lines, lineNum)

				// Determine the version
				version := pkgRef.VersionAttr
				if version == "" {
					version = pkgRef.VersionNested
				}

				// Create package entry
				packages = append(packages, models.Package{
					PackageManager: "dotnet",
					PackageName:    pkgRef.Include,
					Version:        parseVersion(version),
					Filepath:       manifestFile,
					LineStart:      lineStart,
					LineEnd:        lineEnd,
					StartIndex:     startCol,
					EndIndex:       endCol,
				})
			}
		}
	}

	return packages, nil
}
