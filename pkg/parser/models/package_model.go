package models

type Package struct {
	PackageManager string
	PackageName    string
	Version        string
	Filepath       string
	LineStart      int
	LineEnd        int
	StartIndex     int
	EndIndex       int
}
