package directory_packages_props

import (
	"encoding/xml"
	"github.com/Checkmarx/manifest-parser/internal"
	"github.com/Checkmarx/manifest-parser/internal/parsers/csproj"
	"io"
	"os"
	"strings"
)

type DotnetDirectoryPackagesPropsParser struct{}

func (p *DotnetDirectoryPackagesPropsParser) Parse(manifestFile string) ([]internal.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var packages []internal.Package
	var currentElement *csproj.PackageReference

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
			if elem.Name.Local == "PackageVersion" {
				currentElement = &csproj.PackageReference{}
				err := decoder.DecodeElement(currentElement, &elem)
				if err != nil {
					return nil, err
				}
				line, _ := decoder.InputPos()
				packages = append(packages, internal.Package{
					PackageName: currentElement.Include,
					Version:     currentElement.Version,
					LineStart:   line,
					LineEnd:     line,
					Filepath:    manifestFile,
				})
			}
		}
	}

	return packages, nil
}
