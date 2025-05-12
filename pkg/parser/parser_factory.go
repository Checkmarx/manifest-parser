package parser

import (
	"github.com/Checkmarx/manifest-parser/internal/parsers/config"
	"github.com/Checkmarx/manifest-parser/internal/parsers/csproj"
	"github.com/Checkmarx/manifest-parser/internal/parsers/json"
	"github.com/Checkmarx/manifest-parser/internal/parsers/mod"
	"github.com/Checkmarx/manifest-parser/internal/parsers/props"
	"github.com/Checkmarx/manifest-parser/internal/parsers/pypi"
	"github.com/Checkmarx/manifest-parser/internal/parsers/xml"
)

func ParsersFactory(manifest string) Parser {
	manifestType := selectManifestFile(manifest)

	switch manifestType {
	case MavenPom:
		return &xml.PackagesConfigParser{}
	case DotnetCsproj:
		return &csproj.DotnetCsprojParser{}
	case DotnetDirectoryPackagesProps:
		return &props.DotnetDirectoryPackagesPropsParser{}
	case PypiRequirements:
		return &pypi.PypiParser{}
	case NpmPackageJson:
		return &json.NpmPackageJsonParser{}
	case DotnetPackagesConfig:
		return &config.DotnetPackagesConfigParser{}
	case GoMod:
		return &mod.GoModParser{}
	default:
		return nil
	}
}
