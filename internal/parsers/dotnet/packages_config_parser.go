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

type DotnetPackagesConfigParser struct{}

type PackageConfig struct {
	ID            string `xml:"id,attr"`
	VersionAttr   string `xml:"version,attr"`
	VersionNested string `xml:"version"`
	Line          int
}

const PackageTag = "package"

// parseVersionConfig handles version resolution for packages.config
func parseVersionConfig(version string) string {
	if version == "" {
		return "latest"
	}
	if strings.ContainsAny(version, "[]()") {
		return "latest"
	}
	if strings.ContainsAny(version, "*^~><") {
		return "latest"
	}
	return version
}

// computePackageLocations calculates all locations for a package element
func computePackageLocations(lines []string, startLine int) []models.Location {
	if len(lines) == 0 || startLine < 0 || startLine >= len(lines) {
		return nil
	}

	var locations []models.Location
	currentLine := lines[startLine]
	startIdx := strings.Index(currentLine, "<package")
	if startIdx < 0 {
		return nil
	}

	// Single-line format
	if strings.Contains(currentLine, "/>") {
		closeIdx := strings.Index(currentLine[startIdx:], "/>")
		if closeIdx < 0 {
			return nil
		}
		endIdx := startIdx + closeIdx + 2 // index of "/>" relative to startIdx + len("/>")
		return []models.Location{{
			Line:       startLine,
			StartIndex: startIdx,
			EndIndex:   endIdx,
		}}
	}

	// Multi-line format
	openIdx := strings.Index(currentLine[startIdx:], ">")
	if openIdx < 0 {
		return nil
	}
	endIdx := startIdx + openIdx + 1 // index of ">" relative to startIdx + len(">")
	locations = append(locations, models.Location{
		Line:       startLine,
		StartIndex: startIdx,
		EndIndex:   endIdx,
	})

	for i := startLine + 1; i < len(lines) && i < startLine+10; i++ {
		line := lines[i]
		if strings.Contains(line, "</package>") {
			closeIdx := strings.Index(line, "</package>")
			if closeIdx < 0 {
				continue
			}
			endIdx := closeIdx + len("</package>")
			locations = append(locations, models.Location{
				Line:       i,
				StartIndex: closeIdx,
				EndIndex:   endIdx,
			})
			break
		}
		if strings.Contains(line, "<version>") {
			verStart := strings.Index(line, "<version>")
			verEnd := strings.Index(line[verStart:], "</version>")
			if verStart >= 0 && verEnd >= 0 {
				endIdx := verStart + verEnd + len("</version>")
				locations = append(locations, models.Location{
					Line:       i,
					StartIndex: verStart,
					EndIndex:   endIdx,
				})
			}
		}
	}
	return locations
}

func (p *DotnetPackagesConfigParser) Parse(manifestFile string) ([]models.Package, error) {
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

	decoder := xml.NewDecoder(strings.NewReader(strContent))
	var pkgs []PackageConfig

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to parse XML: %w", err)
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == PackageTag {
				var pkg PackageConfig
				if err := decoder.DecodeElement(&pkg, &elem); err != nil {
					return nil, fmt.Errorf("failed to parse XML: %w", err)
				}

				// Skip empty package IDs
				if pkg.ID == "" {
					continue
				}

				// Find line number
				lineNum := 0
				packagePattern := fmt.Sprintf(`package.*id="%s"`, pkg.ID)
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

				pkg.Line = lineNum
				pkgs = append(pkgs, pkg)
			}
		}
	}

	var packages []models.Package
	for _, pkg := range pkgs {
		// Compute locations for both single-line and multi-line formats
		locations := computePackageLocations(lines, pkg.Line)

		// Determine the version
		version := pkg.VersionAttr
		if version == "" {
			version = pkg.VersionNested
		}

		packages = append(packages, models.Package{
			PackageManager: "nuget",
			PackageName:    pkg.ID,
			Version:        parseVersionConfig(version),
			FilePath:       manifestFile,
			Locations:      locations,
		})
	}
	return packages, nil
}
