package versionspec

import "strings"

// Latest is what a package reports as its version when no exact version is
// known. It is a sentinel, not a version.
const Latest = "latest"

// Spec returns the VersionSpec of a package: declared, the version as the
// dependency file writes it, when reportSpecs is set and version is Latest.
// It is empty otherwise, so a spec is never reported next to a version, and
// never unless the caller asked for it.
//
// declared is reported as written: it is not checked to be a range. A package
// that names a source other than a registry reports whatever the file says,
// and the caller sorts it out. The one value it never reports is Latest: a
// file that declares "latest" names no range, so it is the same as declaring
// nothing.
func Spec(declared, version string, reportSpecs bool) string {
	if !reportSpecs || version != Latest {
		return ""
	}

	declared = strings.TrimSpace(declared)
	if declared == Latest {
		return ""
	}

	return declared
}
