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
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}

// parseVersion handles version resolution for Directory.Packages.props
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

// findPackageVersionPosition finds the position of a package version element in the file content
// Returns line number, start column, and end column (all 1-based) for the package name in the element
func findPackageVersionPosition(content string, packageName string) (lineNum, startCol, endCol int) {
	escapedName := regexp.QuoteMeta(packageName)
	pattern := fmt.Sprintf(`<PackageVersion\s+Include="%s"`, escapedName)
	re := regexp.MustCompile(pattern)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		loc := re.FindStringIndex(line)
		if loc != nil {
			// endCol = length of the line (till the last character)
			return i + 1, loc[0] + 1, len(line) + 1
		}
	}
	return 0, 0, 0 // Not found
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

	// Create XML decoder
	strContent := string(content)

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
			if elem.Name.Local == "PackageVersion" {
				var pkgVer PackageVersion
				if err := decoder.DecodeElement(&pkgVer, &elem); err != nil {
					return nil, fmt.Errorf("failed to decode PackageVersion: %w", err)
				}

				// Skip empty package names
				if pkgVer.Include == "" {
					continue
				}

				// Get line number from decoder
				line, _ := decoder.InputPos()

				// Find package version position in file
				_, startCol, endCol := findPackageVersionPosition(strContent, pkgVer.Include)

				// Create package entry
				packages = append(packages, models.Package{
					PackageManager: "dotnet",
					PackageName:    pkgVer.Include,
					Version:        parseVersionProps(pkgVer.Version),
					Filepath:       manifestFile,
					LineStart:      line,
					LineEnd:        line,
					StartIndex:     startCol,
					EndIndex:       endCol,
				})
			}
		}
	}

	return packages, nil
}
