package parser

import (
	"path/filepath"
	"strings"
)

type Manifest int

const (
	PypiRequirements Manifest = iota
	NpmPackageJson
	DotnetCsproj
	DotnetDirectoryPackagesProps
	DotnetPackagesConfig
	MavenPom
	GoMod
	GradleBuild
	GradleVersionCatalog
	SbtBuild
	SwiftPackage
	CocoaPodsPodfile
	CocoaPodsPodspec
	CocoaPodsPodspecJSON
	CarthageCartfile
	CarthageCartfilePrivate
	ComposerJson
	RubyGemsGemfile
	BowerJson
	SetuptoolsSetupCfg
	SetuptoolsSetupPy
	PoetryPyproject
	DartPubspec
)

// selectManifestFile a method to select a manifest file type by its name
func selectManifestFile(manifest string) Manifest {

	manifestFileName := filepath.Base(manifest)

	manifestFileExtension := filepath.Ext(manifestFileName)
	if manifestFileExtension == ".csproj" {
		return DotnetCsproj
	}

	if manifestFileExtension == ".txt" {
		// check if file name starts with "requirement", "packages", or "constraint"
		if strings.HasPrefix(manifestFileName, "requirement") ||
			strings.HasPrefix(manifestFileName, "packages") ||
			strings.HasPrefix(manifestFileName, "constraint") {
			return PypiRequirements
		}
	}

	if manifestFileExtension == ".sbt" {
		return SbtBuild
	}

	if manifestFileName == "pom.xml" {
		return MavenPom
	}

	if manifestFileName == "package.json" {
		return NpmPackageJson
	}

	if manifestFileName == "Directory.Packages.props" {
		return DotnetDirectoryPackagesProps
	}

	if manifestFileName == "packages.config" {
		return DotnetPackagesConfig
	}

	if manifestFileName == "go.mod" {
		return GoMod
	}

	if manifestFileName == "build.gradle" || manifestFileName == "build.gradle.kts" {
		return GradleBuild
	}

	if manifestFileName == "libs.versions.toml" {
		return GradleVersionCatalog
	}

	// SwiftPM:
	//   Package.swift            - Swift DSL manifest (Checkmarx SCA)
	//   Package.resolved         - JSON lock file (helper for version resolution)
	//   Package@swift-X.Y.swift  - Swift-version-tooled manifest. Real Apple feature;
	//                              libraries supporting multiple Swift toolchains ship these.
	//                              Not on Checkmarx's core list but common in the wild.
	//   Package@swift-X.Y.resolved - lock file for multi-toolchain variant (helper)
	if manifestFileName == "Package.swift" ||
		(strings.HasPrefix(manifestFileName, "Package@swift-") && strings.HasSuffix(manifestFileName, ".swift")) {
		return SwiftPackage
	}

	// CocoaPods:
	//   Podfile           - Ruby DSL app manifest (Checkmarx SCA)
	//   Podfile.lock      - YAML lock file (helper for version resolution)
	//   *.podspec         - Ruby DSL pod author spec      (pragmatic)
	//   *.podspec.json    - JSON pod author spec          (pragmatic)
	if manifestFileName == "Podfile" {
		return CocoaPodsPodfile
	}
	if strings.HasSuffix(manifestFileName, ".podspec.json") {
		return CocoaPodsPodspecJSON
	}
	if strings.HasSuffix(manifestFileName, ".podspec") {
		return CocoaPodsPodspec
	}

	// Carthage:
	//   Cartfile          - production dependencies     (Checkmarx SCA, required)
	//   Cartfile.private  - private/test dependencies   (Checkmarx SCA)
	//   Cartfile.resolved - lock file with resolved versions (helper for version resolution)
	if manifestFileName == "Cartfile" {
		return CarthageCartfile
	}
	if manifestFileName == "Cartfile.private" {
		return CarthageCartfilePrivate
	}

	if manifestFileName == "composer.json" {
		return ComposerJson
	}

	if manifestFileName == "Gemfile" {
		return RubyGemsGemfile
	}

	if manifestFileName == "bower.json" {
		return BowerJson
	}

	// yarn.lock is consumed as a sibling helper by the npm parser, not as a standalone manifest.

	if manifestFileName == "setup.cfg" {
		return SetuptoolsSetupCfg
	}

	if manifestFileName == "setup.py" {
		return SetuptoolsSetupPy
	}

	if manifestFileName == "pyproject.toml" {
		return PoetryPyproject
	}

	// Dart / Flutter (pub):
	//   pubspec.yaml - declared dependencies (main manifest)
	//   pubspec.lock - YAML lock file with resolved versions (helper for version resolution)
	if manifestFileName == "pubspec.yaml" {
		return DartPubspec
	}

	return -1
}
