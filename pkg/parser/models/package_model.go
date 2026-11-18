package models

type Location struct {
	Line       int
	StartIndex int
	EndIndex   int
}

type Package struct {
	PackageManager string
	PackageName    string
	// Version is the exact version, or "latest" when none can be determined.
	Version  string
	FilePath string
	// VersionSpec is the version as declared (a range, a wildcard or a tag), set
	// only when Version is "latest" and the parser was built with
	// parser.WithVersionSpecs. Empty otherwise.
	VersionSpec string `json:",omitempty"`
	Locations   []Location
}
