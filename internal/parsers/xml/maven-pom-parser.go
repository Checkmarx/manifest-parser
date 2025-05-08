package xml

import (
	"encoding/xml"
	"github.com/Checkmarx/manifest-parser/pkg/models"
	"io"
	"os"
	"strings"
)

type MavenPomParser struct{}

type Dependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
}

func (p *MavenPomParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(strings.NewReader(string(content)))
	var packages []models.Package
	var currentElement *Dependency

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
			if elem.Name.Local == "dependency" {
				currentElement = &Dependency{}
				lineStart, _ := decoder.InputPos()

				err := decoder.DecodeElement(currentElement, &elem)
				if err != nil {
					return nil, err
				}
				lineEnd, _ := decoder.InputPos()
				packages = append(packages, models.Package{
					PackageName: currentElement.GroupId + ":" + currentElement.ArtifactId,
					Version:     currentElement.Version,
					LineStart:   lineStart,
					LineEnd:     lineEnd,
					Filepath:    manifestFile,
				})
			}
		}
	}

	return packages, nil
}
