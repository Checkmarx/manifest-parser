package parsers

type Parser interface {
	Parse(manifestFile string) ([]Package, error)
}
