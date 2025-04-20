package parsers

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

	return -1
}
