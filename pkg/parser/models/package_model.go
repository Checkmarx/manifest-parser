package models

type Location struct {
	Line       int
	StartIndex int
	EndIndex   int
}

type Package struct {
	PackageManager string
	PackageName    string
	Version        string
	FilePath       string
	Locations      []Location
}
