package models

type Package struct {
	PackageManager string
	PackageName    string
	Version        string
	FilePath       string
	LineStart      int
	LineEnd        int
	StartIndex     int
	EndIndex       int
}
