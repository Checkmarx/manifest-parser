package parser

import (
	"testing"
)

func TestManifestFileSelector_ExpectPom(t *testing.T) {
	manifest := "pom.xml"
	got := selectManifestFile(manifest)
	want := MavenPom
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectCsproj(t *testing.T) {
	manifest := "example.csproj"
	got := selectManifestFile(manifest)
	want := DotnetCsproj
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiRequirements(t *testing.T) {
	manifest := "requirement-dev.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectNpmPackageJson(t *testing.T) {
	manifest := "package.json"
	got := selectManifestFile(manifest)
	want := NpmPackageJson
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectDotnetDirectoryPackagesProps(t *testing.T) {
	manifest := "Directory.Packages.props"
	got := selectManifestFile(manifest)
	want := DotnetDirectoryPackagesProps
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectDotnetPackagesConfig(t *testing.T) {
	manifest := "packages.config"
	got := selectManifestFile(manifest)
	want := DotnetPackagesConfig
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectGoMod(t *testing.T) {
	manifest := "go.mod"
	got := selectManifestFile(manifest)
	want := GoMod
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSbtBuild(t *testing.T) {
	manifest := "build.sbt"
	got := selectManifestFile(manifest)
	want := SbtBuild
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSbtPlugins(t *testing.T) {
	manifest := "plugins.sbt"
	got := selectManifestFile(manifest)
	want := SbtBuild
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSbtCustom(t *testing.T) {
	manifest := "dependencies.sbt"
	got := selectManifestFile(manifest)
	want := SbtBuild
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiRequirementsTxt(t *testing.T) {
	manifest := "requirements.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiRequirementsDev(t *testing.T) {
	manifest := "requirements-dev.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiRequirementSingular(t *testing.T) {
	manifest := "requirement.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiRequirementSingularDev(t *testing.T) {
	manifest := "requirement-dev.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiRequirementsWithPath(t *testing.T) {
	manifest := "/some/path/to/requirements-prod.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiConstraints(t *testing.T) {
	manifest := "constraints.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiConstraintsDev(t *testing.T) {
	manifest := "constraints-dev.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPypiConstraintsWithPath(t *testing.T) {
	manifest := "/some/path/to/constraints-prod.txt"
	got := selectManifestFile(manifest)
	want := PypiRequirements
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSetuptoolsSetupCfg(t *testing.T) {
	manifest := "setup.cfg"
	got := selectManifestFile(manifest)
	want := SetuptoolsSetupCfg
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSetuptoolsSetupPy(t *testing.T) {
	manifest := "setup.py"
	got := selectManifestFile(manifest)
	want := SetuptoolsSetupPy
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectPoetryPyproject(t *testing.T) {
	manifest := "pyproject.toml"
	got := selectManifestFile(manifest)
	want := PoetryPyproject
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}
