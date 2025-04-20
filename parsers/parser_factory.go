package parsers

func ParserFactory(manifest string) Parser {
	manifestType := selectManifestFile(manifest)

	switch manifestType {
	case MavenPom:
		return &MavenPomParser{}
	case DotnetCsproj:
		return &DotnetCsprojParser{}
	case DotnetDirectoryPackagesProps:
		return &DotnetDirectoryPackagesPropsParser{}
	case PypiRequirements:
		return &PypiParser{}
	case NpmPackageJson:
		return &NpmPackageJsonParser{}
	case DotnetPackagesConfig:
		return &DotnetPackagesConfigParser{}
	default:
		return nil
	}
}
