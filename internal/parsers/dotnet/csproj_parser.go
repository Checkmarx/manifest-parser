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

// PackageReferenceTag is the XML tag for package references in .csproj files
const PackageReferenceTag = "PackageReference"

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

// computeLocations calculates all locations for a PackageReference element
func computeLocations(lines []string, startLine int) []models.Location {
	var locations []models.Location
	currentLine := lines[startLine]

	// Find the position of the PackageReference tag start in the line
	startIdx := strings.Index(currentLine, "<PackageReference")
	if startIdx < 0 {
		return []models.Location{{
			Line:       startLine,
			StartIndex: 1,
			EndIndex:   len(currentLine),
		}}
	}

	// Check if it's a single-line format
	if strings.Contains(currentLine, "/>") {
		// Single-line format
		endIdx := strings.LastIndex(currentLine, "/>") + 2 // Include the "/>" itself
		return []models.Location{{
			Line:       startLine,
			StartIndex: startIdx,
			EndIndex:   endIdx,
		}}
	}

	// Multi-line format
	// Add the first line
	locations = append(locations, models.Location{
		Line:       startLine,
		StartIndex: startIdx,
		EndIndex:   len(currentLine),
	})

	// Add all lines until the closing tag
	for i := startLine + 1; i < len(lines) && i < startLine+10; i++ { // Limit search to 10 lines
		line := lines[i]
		if strings.Contains(line, "</PackageReference>") {
			startIdxInLineEnd := strings.Index(line, "</PackageReference>")

			endIdx := strings.Index(line, "</PackageReference>") + len("</PackageReference>")
			locations = append(locations, models.Location{
				Line:       i,
				StartIndex: startIdxInLineEnd, // True start index including indentation
				EndIndex:   endIdx,
			})
			break
		}

		startIdxInLine := strings.Index(line, "<Version>")
		// Add intermediate lines
		locations = append(locations, models.Location{
			Line:       i,
			StartIndex: startIdxInLine,
			EndIndex:   len(line),
		})
	}

	return locations
}

// Parse implements the Parser interface for .csproj files
func (p *DotnetCsprojParser) Parse(manifestFile string) ([]models.Package, error) {
	// Read the file content
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Split content into lines for index computation
	strContent := string(content)
	lines := strings.Split(strContent, "\n")

	// Create XML decoder
	decoder := xml.NewDecoder(strings.NewReader(strContent))
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
			if elem.Name.Local == PackageReferenceTag {
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
						lineNum = i
						break
					}
				}

				// Skip if line not found
				if lineNum == 0 {
					continue
				}

				// Compute locations for both single-line and multi-line formats
				locations := computeLocations(lines, lineNum)

				// Determine the version
				version := pkgRef.VersionAttr
				if version == "" {
					version = pkgRef.VersionNested
				}

				// Create package entry
				packages = append(packages, models.Package{
					PackageManager: "nuget",
					PackageName:    pkgRef.Include,
					Version:        parseVersion(version),
					FilePath:       manifestFile,
					Locations:      locations,
				})
			}
		}
	}

	return packages, nil
}
