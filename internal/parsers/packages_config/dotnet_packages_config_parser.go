package packages_config

import (
	"ManifestParser/internal/parsers"
	"encoding/xml"
	"io"
	"os"
	"strings"
)

type DotnetPackagesConfigParser struct{}

func (p *DotnetPackagesConfigParser) Parse(manifest string) ([]parsers.Package, error) {
	content, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var packages []parsers.Package

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
			if elem.Name.Local == "package" {
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
					lineEnd := lineStart
					packages = append(packages, parsers.Package{
						PackageName: id,
						Version:     version,
						LineStart:   lineStart,
						LineEnd:     lineEnd,
						Filepath:    manifest,
					})
				}
			}
		}
	}
	return packages, nil
}
