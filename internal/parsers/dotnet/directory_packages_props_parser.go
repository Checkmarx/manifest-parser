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

// DotnetDirectoryPackagesPropsParser implements parsing of Directory.Packages.props files
// These files are used for central package management in .NET projects
type DotnetDirectoryPackagesPropsParser struct{}

// PackageVersion represents a <PackageVersion> element in Directory.Packages.props
type PackageVersion struct {
	Include       string `xml:"Include,attr"`
	VersionAttr   string `xml:"Version,attr"`
	VersionNested string `xml:"Version"`
}

const PackageVersionTag = "PackageVersion"

// parseVersionProps handles version resolution for Directory.Packages.props
// Returns:
// - Exact version if specified
// - "latest" for version ranges or special version specifiers
func parseVersionProps(version string) string {
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

// computePackageVersionLocations calculates all locations for a PackageVersion element
func computePackageVersionLocations(lines []string, startLine int) []models.Location {
	// Handle empty lines or invalid start line
	if len(lines) == 0 || startLine < 0 || startLine >= len(lines) {
		return nil
	}

	var locations []models.Location
	currentLine := lines[startLine]

	// Find the position of the PackageVersion tag start in the line
	startIdx := strings.Index(currentLine, "<PackageVersion")
	if startIdx < 0 {
		return nil
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
	// Add the first line - only include the opening tag
	endIdx := strings.Index(currentLine, ">") + 1 // Include the ">" itself
	locations = append(locations, models.Location{
		Line:       startLine,
		StartIndex: startIdx,
		EndIndex:   endIdx,
	})

	// Add all lines until the closing tag
	for i := startLine + 1; i < len(lines) && i < startLine+10; i++ { // Limit search to 10 lines
		line := lines[i]
		if strings.Contains(line, "</PackageVersion>") {
			startIdxInLineEnd := strings.Index(line, "</PackageVersion>")
			endIdx := strings.Index(line, "</PackageVersion>") + len("</PackageVersion>")
			locations = append(locations, models.Location{
				Line:       i,
				StartIndex: startIdxInLineEnd,
				EndIndex:   endIdx,
			})
			break
		}
		startIdxInLine := strings.Index(line, "<Version>")
		if startIdxInLine >= 0 {
			// Add intermediate lines
			locations = append(locations, models.Location{
				Line:       i,
				StartIndex: startIdxInLine,
				EndIndex:   len(line),
			})
		}
	}

	return locations
}

// Parse implements the Parser interface for Directory.Packages.props files
func (p *DotnetDirectoryPackagesPropsParser) Parse(manifestFile string) ([]models.Package, error) {
	// Read the file content
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Handle empty file
	if len(content) == 0 {
		return nil, fmt.Errorf("empty file")
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
			if elem.Name.Local == PackageVersionTag {
				var pkgVer PackageVersion
				if err := decoder.DecodeElement(&pkgVer, &elem); err != nil {
					return nil, fmt.Errorf("failed to decode PackageVersion: %w", err)
				}

				// Skip empty package names
				if pkgVer.Include == "" {
					continue
				}

				// Find line number
				lineNum := 0
				packagePattern := fmt.Sprintf(`PackageVersion.*Include="%s"`, pkgVer.Include)
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
				locations := computePackageVersionLocations(lines, lineNum)

				// Determine the version
				version := pkgVer.VersionAttr
				if version == "" {
					version = pkgVer.VersionNested
				}

				// Create package entry
				packages = append(packages, models.Package{
					PackageManager: "nuget",
					PackageName:    pkgVer.Include,
					Version:        parseVersionProps(version),
					FilePath:       manifestFile,
					Locations:      locations,
				})
			}
		}
	}

	return packages, nil
}
