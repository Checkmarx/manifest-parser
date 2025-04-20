package parser

import (
	"ManifestParser/parsers/csproj_parser"
	"ManifestParser/parsers/directory_packages_props_parser"
	"ManifestParser/parsers/package_json_parser"
	"ManifestParser/parsers/packages_config_parser"
	"ManifestParser/parsers/pom_xml_parser"
	"ManifestParser/parsers/pypi_parser"
)

func ParsersFactory(manifest string) Parser {
	manifestType := selectManifestFile(manifest)

	switch manifestType {
	case MavenPom:
		return &pom_xml_parser.MavenPomParser{}
	case DotnetCsproj:
		return &csproj_parser.DotnetCsprojParser{}
	case DotnetDirectoryPackagesProps:
		return &directory_packages_props_parser.DotnetDirectoryPackagesPropsParser{}
	case PypiRequirements:
		return &pypi_parser.PypiParser{}
	case NpmPackageJson:
		return &package_json_parser.NpmPackageJsonParser{}
	case DotnetPackagesConfig:
		return &packages_config_parser.DotnetPackagesConfigParser{}
	default:
		return nil
	}
}
