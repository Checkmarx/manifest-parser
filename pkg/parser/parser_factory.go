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
		return &csproj.CsprojParser{}
	case DotnetDirectoryPackagesProps:
		return &props.PropsParser{}
	case PypiRequirements:
		return &pypi.PypiParser{}
	case NpmPackageJson:
		return &json.NpmPackageJsonParser{}
	case DotnetPackagesConfig:
		return &config.ConfigParser{}
	case GoMod:
		return &mod.ModParser{}
	default:
		return nil
	}
}
