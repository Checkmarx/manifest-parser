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
	SwiftPackageResolved
	CocoaPodsPodfile
	CocoaPodsPodfileLock
	CocoaPodsPodspec
	CocoaPodsPodspecJSON
	CarthageCartfile
	CarthageCartfilePrivate
	CarthageCartfileResolved
	ComposerJson
	RubyGemsGemfile
	BowerJson
)

// selectManifestFile a method to select a manifest file type by its name
func selectManifestFile(manifest string) Manifest {

	manifestFileName := filepath.Base(manifest)

	manifestFileExtension := filepath.Ext(manifestFileName)
	if manifestFileExtension == ".csproj" {
		return DotnetCsproj
	}

	if manifestFileExtension == ".txt" {
		//check if file name starts with "requirement" or "packages"
		if strings.HasPrefix(manifestFileName, "requirement") ||
			strings.HasPrefix(manifestFileName, "packages") {
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
	//   Package.resolved         - JSON lock file (Checkmarx SCA)
	//   Package@swift-X.Y.swift  - Swift-version-tooled manifest. Real Apple feature;
	//                              libraries supporting multiple Swift toolchains ship these.
	//                              Not on Checkmarx's core list but common in the wild.
	if manifestFileName == "Package.resolved" {
		return SwiftPackageResolved
	}
	if manifestFileName == "Package.swift" ||
		(strings.HasPrefix(manifestFileName, "Package@swift-") && strings.HasSuffix(manifestFileName, ".swift")) {
		return SwiftPackage
	}

	// CocoaPods:
	//   Podfile           - Ruby DSL app manifest (Checkmarx SCA)
	//   Podfile.lock      - YAML lock file        (Checkmarx SCA)
	//   *.podspec         - Ruby DSL pod author spec      (pragmatic)
	//   *.podspec.json    - JSON pod author spec          (pragmatic)
	if manifestFileName == "Podfile" {
		return CocoaPodsPodfile
	}
	if manifestFileName == "Podfile.lock" {
		return CocoaPodsPodfileLock
	}
	if strings.HasSuffix(manifestFileName, ".podspec.json") {
		return CocoaPodsPodspecJSON
	}
	if strings.HasSuffix(manifestFileName, ".podspec") {
		return CocoaPodsPodspec
	}

	// Carthage (all three share the same syntax; routing differs only to
	// distinguish resolved-vs-spec semantics downstream):
	//   Cartfile          - production dependencies     (Checkmarx SCA, required)
	//   Cartfile.private  - private/test dependencies   (Checkmarx SCA)
	//   Cartfile.resolved - lock file with resolved vers (Checkmarx SCA)
	if manifestFileName == "Cartfile" {
		return CarthageCartfile
	}
	if manifestFileName == "Cartfile.private" {
		return CarthageCartfilePrivate
	}
	if manifestFileName == "Cartfile.resolved" {
		return CarthageCartfileResolved
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

	return -1
}
