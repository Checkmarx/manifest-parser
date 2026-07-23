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

func TestManifestFileSelector_ExpectSwiftPackage(t *testing.T) {
	manifest := "Package.swift"
	got := selectManifestFile(manifest)
	want := SwiftPackage
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSwiftPackageResolved(t *testing.T) {
	manifest := "Package.resolved"
	got := selectManifestFile(manifest)
	want := SwiftPackageResolved
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectSwiftPackageToolchainVariant(t *testing.T) {
	// Apple-documented multi-toolchain manifest. Real projects ship these even though
	// Checkmarx's core SCA list only mentions Package.swift.
	for _, manifest := range []string{"Package@swift-5.5.swift", "Package@swift-6.swift"} {
		got := selectManifestFile(manifest)
		want := SwiftPackage
		if got != want {
			t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
		}
	}
}

func TestManifestFileSelector_ExpectComposerJson(t *testing.T) {
	manifest := "composer.json"
	got := selectManifestFile(manifest)
	want := ComposerJson
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectGemfile(t *testing.T) {
	manifest := "Gemfile"
	got := selectManifestFile(manifest)
	want := RubyGemsGemfile
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_ExpectBowerJson(t *testing.T) {
	manifest := "bower.json"
	got := selectManifestFile(manifest)
	want := BowerJson
	if got != want {
		t.Errorf("selectManifestFile(%q) = %v; want %v", manifest, got, want)
	}
}

func TestManifestFileSelector_YarnLockNotStandalone(t *testing.T) {
	if got := selectManifestFile("yarn.lock"); got != -1 {
		t.Errorf("yarn.lock should not be a standalone manifest; got %v, want -1", got)
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
