// Package versionspec classifies the versions a dependency file declares. It
// decides whether a declared value is an exact version, a range, a wildcard, a
// tag, or none of those (a source other than the registry, an unresolved
// property, a missing version), and builds the version and version spec a
// Package reports from that.
package versionspec

import (
	"regexp"
	"strings"
)

// Latest is what a package reports as its version when no exact version is
// known. It is a sentinel, not a version.
const Latest = "latest"

// npmExact is a full semantic version, the only exact form of an npm version:
// anything else ("1.x", "^4.0.0", "*", "next") is a range or a tag.
var npmExact = regexp.MustCompile(`^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)

// Classify sorts a declared version in ecosystem into an exact version, a
// version spec (a range, a wildcard or a tag), or neither.
//
// Neither is returned for a value that isn't a registry version at all: an
// empty value, or one with "/", ":" or "$" (an npm "file:../x", "git+https://...",
// "workspace:*", an unresolved "${property}"). No registry version, in any
// ecosystem, contains those.
//
// ecosystem is one of "npm", "pypi", "maven", "gradle" or "nuget". Any other
// ecosystem is treated as having only exact versions.
func Classify(ecosystem, value string) (exact, spec string) {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "/:$") {
		return "", ""
	}

	if isExact(ecosystem, value) {
		return value, ""
	}

	return "", value
}

// Result returns the Version and VersionSpec a Package reports. Version is the
// exact version, or Latest when there isn't one. VersionSpec is spec, and only
// when report is set: it is the caller's opt-in, and it is never set alongside
// an exact version.
func Result(exact, spec string, report bool) (version, versionSpec string) {
	if exact != "" {
		return exact, ""
	}

	if report {
		return Latest, spec
	}

	return Latest, ""
}

// isExact reports whether a value that already passed Classify's registry-version
// check names one fixed version in ecosystem.
func isExact(ecosystem, value string) bool {
	switch ecosystem {
	case "npm":
		return npmExact.MatchString(value)
	case "maven", "gradle":
		lower := strings.ToLower(value)

		return !strings.ContainsAny(value, "[](),^~*><+") &&
			lower != "latest" && lower != "release" && !strings.HasPrefix(lower, "latest.")
	case "pypi":
		return !strings.ContainsAny(value, "<>=!~*,;")
	case "nuget":
		return !strings.ContainsAny(value, "[](),*^~><")
	default:
		return true
	}
}
