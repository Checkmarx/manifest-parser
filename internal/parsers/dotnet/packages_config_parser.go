package dotnet

import (
	"encoding/xml"
	"io"
	"os"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type DotnetPackagesConfigParser struct{}

type PackageConfig struct {
	ID      string
	Version string
	Line    int
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

// findPackageTagPosition returns the start column and EndIndex as the length of the line + 1
func findPackageTagPosition(lines []string, lineNum int) (startCol, endCol int) {
	if lineNum > 0 && lineNum <= len(lines) {
		line := lines[lineNum-1]
		idx := strings.Index(line, "<package")
		if idx >= 0 {
			startCol = idx + 1 // 1-based, including leading spaces
			endCol = len(line) + 1
			return startCol, endCol
		}
	}
	return 0, 0
}

func (p *DotnetPackagesConfigParser) Parse(manifest string) ([]models.Package, error) {
	content, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}
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
			return nil, err
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Local == PackageTag {
				var id, version string
				for _, attr := range elem.Attr {
					if attr.Name.Local == "id" {
						id = attr.Value
					}
					if attr.Name.Local == "version" {
						version = attr.Value
					}
				}
				if id != "" && version != "" {
					lineStart, _ := decoder.InputPos()
					pkgs = append(pkgs, PackageConfig{
						ID:      id,
						Version: version,
						Line:    lineStart,
					})
				}
			}
		}
	}

	var packages []models.Package
	for _, pkg := range pkgs {
		startCol, endCol := findPackageTagPosition(lines, pkg.Line)
		packages = append(packages, models.Package{
			PackageName: pkg.ID,
			Version:     parseVersionConfig(pkg.Version),
			LineStart:   pkg.Line,
			LineEnd:     pkg.Line,
			StartIndex:  startCol,
			EndIndex:    endCol,
			Filepath:    manifest,
		})
	}
	return packages, nil
}
