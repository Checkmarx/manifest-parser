package parser

// options are the settings a parser is built with, see ParsersFactoryWithOptions.
type options struct {
	versionSpecs bool
}

// Option changes how ParsersFactoryWithOptions builds a parser.
type Option func(*options)

// WithVersionSpecs makes a parser report, in Package.VersionSpec, the version
// range a dependency is declared with when no exact version can be determined.
// Without it VersionSpec is always empty. It applies to the npm, PyPI, Maven,
// Gradle and NuGet parsers; the other parsers ignore it.
func WithVersionSpecs() Option {
	return func(o *options) {
		o.versionSpecs = true
	}
}
