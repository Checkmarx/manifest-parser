package dotnet

import (
	"encoding/xml"
	"io"
	"os"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type DotnetCsprojParser struct{}

type PackageReference struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}

func (p *DotnetCsprojParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var packages []models.Package
	var currentElement *PackageReference

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
			if elem.Name.Local == "PackageReference" {
				currentElement = &PackageReference{}
				err := decoder.DecodeElement(currentElement, &elem)
				if err != nil {
					return nil, err
				}
				line, _ := decoder.InputPos()
				packages = append(packages, models.Package{
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
