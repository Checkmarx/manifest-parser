package parser

import (
	"ManifestParser/internal/parsers/csproj"
	"ManifestParser/internal/parsers/directory_packages_props"
	"ManifestParser/internal/parsers/package_json"
	"ManifestParser/internal/parsers/packages_config"
	"ManifestParser/internal/parsers/pom_xml"
	"ManifestParser/internal/parsers/pypi"
)

func ParsersFactory(manifest string) Parser {
	manifestType := selectManifestFile(manifest)

	switch manifestType {
	case MavenPom:
		return &pom_xml.MavenPomParser{}
	case DotnetCsproj:
		return &csproj.DotnetCsprojParser{}
	case DotnetDirectoryPackagesProps:
		return &directory_packages_props.DotnetDirectoryPackagesPropsParser{}
	case PypiRequirements:
		return &pypi.PypiParser{}
	case NpmPackageJson:
		return &package_json.NpmPackageJsonParser{}
	case DotnetPackagesConfig:
		return &packages_config.DotnetPackagesConfigParser{}
	default:
		return nil
	}
}
