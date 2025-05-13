package parser

import (
	"github.com/Checkmarx/manifest-parser/internal/parsers/dotnet"
	"github.com/Checkmarx/manifest-parser/internal/parsers/golang"
	"github.com/Checkmarx/manifest-parser/internal/parsers/maven"
	"github.com/Checkmarx/manifest-parser/internal/parsers/npm"
	"github.com/Checkmarx/manifest-parser/internal/parsers/pypi"
)

func ParsersFactory(manifest string) Parser {
	manifestType := selectManifestFile(manifest)

	switch manifestType {
	case MavenPom:
		return &maven.MavenPomParser{}
	case DotnetCsproj:
		return &dotnet.DotnetCsprojParser{}
	case DotnetDirectoryPackagesProps:
		return &dotnet.DotnetDirectoryPackagesPropsParser{}
	case PypiRequirements:
		return &pypi.PypiParser{}
	case NpmPackageJson:
		return &npm.NpmPackageJsonParser{}
	case DotnetPackagesConfig:
		return &dotnet.DotnetPackagesConfigParser{}
	case GoMod:
		return &golang.GoModParser{}
	default:
		return nil
	}
}
