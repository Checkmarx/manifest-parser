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
// (for central package management in dotnet)
type DotnetDirectoryPackagesPropsParser struct{}

// PackageVersion represents a <PackageVersion> element
// (reusing the same struct as csproj for simplicity)
type PackageVersion struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}

// findPackageVersionPosition finds the position of a package name in the file content
func findPackageVersionPosition(content string, packageName string) (startIndex, endIndex int) {
	pattern := fmt.Sprintf(`<PackageVersion\s+Include=\"%s\"`, regexp.QuoteMeta(packageName))
	startIndex = strings.Index(content, pattern)
	if startIndex == -1 {
		return 0, 0
	}
	packageStart := startIndex + len(`<PackageVersion Include="`)
	packageEnd := packageStart + len(packageName)
	return packageStart, packageEnd // Return 0-indexed positions
}

// parseVersion handles version resolution for Directory.Packages.props
func parseVersionProps(version string) string {
	if version == "" {
		return "latest"
	}
	if strings.ContainsAny(version, "[]()*^~><") {
		return "latest"
	}
	return version
}

// Parse implements the Parser interface for Directory.Packages.props files
func (p *DotnetDirectoryPackagesPropsParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Handle empty file
	if len(content) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var packages []models.Package

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
			if elem.Name.Local == "PackageVersion" {
				var pkgVer PackageVersion
				if err := decoder.DecodeElement(&pkgVer, &elem); err != nil {
					return nil, fmt.Errorf("failed to decode PackageVersion: %w", err)
				}
				line, _ := decoder.InputPos()
				startIndex, endIndex := findPackageVersionPosition(string(content), pkgVer.Include)
				packages = append(packages, models.Package{
					PackageManager: "dotnet",
					PackageName:    pkgVer.Include,
					Version:        parseVersionProps(pkgVer.Version),
					Filepath:       manifestFile,
					LineStart:      line,
					LineEnd:        line,
					StartIndex:     startIndex + 1, // Convert to 1-indexed
					EndIndex:       endIndex + 1,   // Convert to 1-indexed
				})
			}
		}
	}

	return packages, nil
}
