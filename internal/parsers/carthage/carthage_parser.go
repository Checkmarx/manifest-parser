// Package carthage parses Carthage Cartfile / Cartfile.private / Cartfile.resolved.
//
// All three files share the same grammar — only the meaning differs:
//   - Cartfile          : production dependency specifiers (may be ranged)
//   - Cartfile.private  : private / test dependency specifiers
//   - Cartfile.resolved : lock file with concrete resolved versions
//
// Three origin keywords are supported per the official Carthage README:
//   github "owner/repo" <version-spec>
//   git    "url"        <version-spec>
//   binary "url"        <version-spec>
//
// The version spec is optional. When present it can be one of:
//   "1.2.3"           — exact tag (resolves concretely)
//   == 1.2.3          — exact equal
//   ~> 1.2.3          — semver-compatible range  -> "latest"
//   >= 1.2.3          — minimum                  -> "latest"
//   "branch-name"     — branch                   -> "latest"
//   "abc1234"         — commit SHA               -> "latest"
package carthage

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

const packageManagerName = "carthage"

// CarthageParser handles all three Cartfile variants — they share the same grammar.
type CarthageParser struct{}

// originLine matches one Carthage dependency declaration:
//   <origin-keyword> "<source>" [<version-spec>]
//
// Group 1: origin keyword (github | git | binary)
// Group 2: source string inside quotes (owner/repo for github, URL otherwise)
// Group 3: optional version specifier — everything after the source, comments already stripped
var originLine = regexp.MustCompile(
	`^\s*(github|git|binary)\s+"([^"]+)"\s*(.*)$`,
)

// Parse implements the Parser interface.
func (p *CarthageParser) Parse(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}

	var packages []models.Package
	for i, raw := range lines {
		code := stripInlineComment(raw)
		if strings.TrimSpace(code) == "" {
			continue
		}

		match := originLine.FindStringSubmatch(code)
		if match == nil {
			continue
		}
		origin, source, spec := match[1], match[2], strings.TrimSpace(match[3])

		name := packageNameFromSource(origin, source)
		if name == "" {
			continue
		}

		startIdx := len(raw) - len(strings.TrimLeft(raw, " \t"))
		endIdx := len(strings.TrimRight(code, " \t"))
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        resolveVersion(spec),
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       i,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}},
		})
	}

	return packages, nil
}

// packageNameFromSource turns the quoted source string into a package name.
//
//   github "Alamofire/Alamofire"            -> "Alamofire/Alamofire"
//   git    "https://gitserver.com/foo.git"  -> "foo"
//   binary "https://example.com/foo.json"   -> "foo"
//
// For github origin we keep the "owner/repo" form so downstream consumers can
// build a canonical identity. For git / binary origins we use the last path
// segment minus any trailing .git / .json suffix.
func packageNameFromSource(origin, source string) string {
	if origin == "github" {
		return source
	}
	trimmed := strings.TrimRight(source, "/")
	if i := strings.LastIndexAny(trimmed, "/:"); i >= 0 {
		trimmed = trimmed[i+1:]
	}
	trimmed = strings.TrimSuffix(trimmed, ".git")
	trimmed = strings.TrimSuffix(trimmed, ".json")
	return trimmed
}

// resolveVersion returns the concrete version if the spec is a quoted exact
// tag or `== X.Y.Z`, otherwise "latest" (per parser contract).
func resolveVersion(spec string) string {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "latest"
	}

	// "1.2.3" or "branch-name" — quoted form
	if strings.HasPrefix(spec, `"`) && strings.HasSuffix(spec, `"`) && len(spec) >= 2 {
		inner := spec[1 : len(spec)-1]
		if looksLikeSemver(inner) {
			return inner
		}
		return "latest"
	}

	// == 1.2.3 — explicit-equality operator (concrete)
	if strings.HasPrefix(spec, "==") {
		v := strings.TrimSpace(strings.TrimPrefix(spec, "=="))
		if looksLikeSemver(v) {
			return v
		}
		return "latest"
	}

	// Anything else (~> X, >= X, branch refs, etc.) is non-concrete.
	return "latest"
}

// looksLikeSemver is a very loose check: digits, dots, and optional pre-release
// suffix. Good enough to distinguish "1.2.3" from "main" / "abc123def".
//
// Commit SHAs are all-hex (no dots) so they fail this check and resolve to "latest".
// Branch names that happen to look like versions are rare and would resolve to
// the version literal — acceptable behaviour.
var semverLike = regexp.MustCompile(`^v?\d+(\.\d+)*([.-][A-Za-z0-9.+-]+)?$`)

func looksLikeSemver(s string) bool {
	return semverLike.MatchString(s)
}

// stripInlineComment removes a trailing `# ...` from a Cartfile line, respecting
// double-quoted strings so `#` inside quotes is preserved.
func stripInlineComment(line string) string {
	inDouble := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\\' && inDouble && i+1 < len(line):
			i++ // skip escaped char
		case ch == '"':
			inDouble = !inDouble
		case !inDouble && ch == '#':
			return line[:i]
		}
	}
	return line
}
