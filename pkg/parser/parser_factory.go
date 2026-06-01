package parser

import (
	"github.com/Checkmarx/manifest-parser/internal/parsers/dotnet"
	"github.com/Checkmarx/manifest-parser/internal/parsers/golang"
	"github.com/Checkmarx/manifest-parser/internal/parsers/gradle"
	"github.com/Checkmarx/manifest-parser/internal/parsers/maven"
	"github.com/Checkmarx/manifest-parser/internal/parsers/npm"
	"github.com/Checkmarx/manifest-parser/internal/parsers/poetry"
	"github.com/Checkmarx/manifest-parser/internal/parsers/pypi"
	"github.com/Checkmarx/manifest-parser/internal/parsers/sbt"
	"github.com/Checkmarx/manifest-parser/internal/parsers/setuptools"
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
	case GradleBuild:
		return &gradle.GradleParser{}
	case GradleVersionCatalog:
		return &gradle.VersionCatalogParser{}
	case SbtBuild:
		return &sbt.SbtParser{}
	case SetuptoolsSetupCfg:
		return &setuptools.SetupCfgParser{}
	case SetuptoolsSetupPy:
		return &setuptools.SetupPyParser{}
	case PoetryPyproject:
		return &poetry.PoetryPyprojectParser{}
	default:
		return nil
	}
}
