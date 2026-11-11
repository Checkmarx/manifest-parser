package versionspec

import (
	"regexp"
	"strings"
)

// pep508Requirement splits a PEP 508 requirement into its name and version
// specifier: `name[extras] (spec) ; markers`.
var pep508Requirement = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)\s*(?:\[[^\]]*\])?\s*(.*)$`)

// ParsePEP508 returns the name and version specifier of a PEP 508 requirement
// (spec is empty when it declares none). Environment markers and extras are
// dropped. ok is false for anything that isn't a plain named requirement, such
// as a direct URL reference ("name @ https://...").
func ParsePEP508(requirement string) (name, spec string, ok bool) {
	requirement = strings.TrimSpace(requirement)
	if i := strings.Index(requirement, ";"); i >= 0 {
		requirement = requirement[:i]
	}

	match := pep508Requirement.FindStringSubmatch(strings.TrimSpace(requirement))
	if match == nil {
		return "", "", false
	}

	spec = strings.TrimSpace(match[2])
	if strings.HasPrefix(spec, "@") {
		return "", "", false
	}

	spec = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(spec, "("), ")"))

	return match[1], spec, true
}

// PEP508Version returns the exact version a requirement pins with "==" or
// "===", or "" when it pins none, and the version specifier the requirement
// declares ("" when it has none). A range, a wildcard, or a pin that isn't a
// single version, such as "==1.0,<2", gives no exact version.
func PEP508Version(requirement string) (exact, declared string) {
	_, spec, ok := ParsePEP508(requirement)
	if !ok || spec == "" {
		return "", ""
	}

	var pinned string
	switch {
	case strings.HasPrefix(spec, "==="):
		pinned = strings.TrimSpace(spec[3:])
	case strings.HasPrefix(spec, "=="):
		pinned = strings.TrimSpace(spec[2:])
	default:
		return "", spec
	}

	exact, _ = Classify("pypi", pinned)

	return exact, spec
}
