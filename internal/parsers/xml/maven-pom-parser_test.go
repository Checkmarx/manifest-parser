package xml

import (
	"testing"

	"github.com/Checkmarx/manifest-parser/internal"
)

func TestMavenPomParser_Parse(t *testing.T) {
	parser := &MavenPomParser{}
	manifestFile := "../../../test/resources/pom.xml"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Error("Error parsing manifest file: ", err)
	}

	expectedPackages := []internal.Package{
		{
			PackageName: "org.eclipse.jetty.ee10:jetty-ee10-bom",
			Version:     "12.0.10",
			LineStart:   159,
			LineEnd:     165,
			Filepath:    manifestFile,
		},
		{
			PackageName: "org.ow2.asm:asm",
			Version:     "9.7",
			LineStart:   166,
			LineEnd:     170,
			Filepath:    manifestFile,
		},
		{
			PackageName: "org.apache.commons:commons-exec",
			Version:     "",
			LineStart:   174,
			LineEnd:     177,
			Filepath:    manifestFile,
		},
		{
			PackageName: "org.asciidoctor:asciidoctorj",
			Version:     "",
			LineStart:   210,
			LineEnd:     213,
			Filepath:    manifestFile,
		},
	}

	internal.ValidatePackages(t, packages, expectedPackages)
}
