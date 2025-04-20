package parser

import (
	"github.com/Checkmarx/manifest-parser/internal/parsers/csproj"
	"github.com/Checkmarx/manifest-parser/internal/parsers/directory_packages_props"
	"github.com/Checkmarx/manifest-parser/internal/parsers/package_json"
	"github.com/Checkmarx/manifest-parser/internal/parsers/packages_config"
	"github.com/Checkmarx/manifest-parser/internal/parsers/pom_xml"
	"github.com/Checkmarx/manifest-parser/internal/parsers/pypi"
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
