package maven

import (
	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestMavenPomParser_Parse(t *testing.T) {
	parser := &MavenPomParser{}
	manifestFile := "../../../test/resources/pom.xml"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []models.Package{
		{
			PackageName: "org.eclipse.jetty.ee10:jetty-ee10-bom",
			Version:     "12.0.10",
			LineStart:   159,
			LineEnd:     165,
			FilePath:    manifestFile,
		},
		{
			PackageName: "org.ow2.asm:asm",
			Version:     "9.7",
			LineStart:   166,
			LineEnd:     170,
			FilePath:    manifestFile,
		},
		{
			PackageName: "org.apache.commons:commons-exec",
			Version:     "",
			LineStart:   174,
			LineEnd:     177,
			FilePath:    manifestFile,
		},
		{
			PackageName: "org.asciidoctor:asciidoctorj",
			Version:     "",
			LineStart:   210,
			LineEnd:     213,
			FilePath:    manifestFile,
		},
	}

	testdata.ValidatePackages(t, packages, expectedPackages)
}
