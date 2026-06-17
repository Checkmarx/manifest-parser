package parser

import (
	"github.com/Checkmarx/manifest-parser/internal/parsers/bower"
	"github.com/Checkmarx/manifest-parser/internal/parsers/carthage"
	"github.com/Checkmarx/manifest-parser/internal/parsers/cocoapods"
	"github.com/Checkmarx/manifest-parser/internal/parsers/composer"
	"github.com/Checkmarx/manifest-parser/internal/parsers/dotnet"
	"github.com/Checkmarx/manifest-parser/internal/parsers/golang"
	"github.com/Checkmarx/manifest-parser/internal/parsers/gradle"
	"github.com/Checkmarx/manifest-parser/internal/parsers/maven"
	"github.com/Checkmarx/manifest-parser/internal/parsers/npm"
	"github.com/Checkmarx/manifest-parser/internal/parsers/pypi"
	"github.com/Checkmarx/manifest-parser/internal/parsers/rubygems"
	"github.com/Checkmarx/manifest-parser/internal/parsers/sbt"
	"github.com/Checkmarx/manifest-parser/internal/parsers/swiftpm"
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
	case SwiftPackage, SwiftPackageResolved:
		return &swiftpm.SwiftPmParser{}
	case CocoaPodsPodfile, CocoaPodsPodfileLock, CocoaPodsPodspec, CocoaPodsPodspecJSON:
		return &cocoapods.CocoaPodsParser{}
	case CarthageCartfile, CarthageCartfilePrivate, CarthageCartfileResolved:
		return &carthage.CarthageParser{}
	case ComposerJson:
		return &composer.ComposerJsonParser{}
	case RubyGemsGemfile:
		return &rubygems.GemfileParser{}
	case BowerJson:
		return &bower.BowerJsonParser{}
	default:
		return nil
	}
}
