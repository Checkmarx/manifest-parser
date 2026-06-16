package cocoapods

import (
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// Podfile forms we recognise (after stripping inline `# ...` comments):
//
//   pod 'Alamofire', '5.2.0'
//   pod 'Alamofire', '~> 5.2'                                 -> latest
//   pod 'Alamofire'                                           -> latest
//   pod 'Realm', :git => 'https://...', :branch => 'main'     -> latest
//   pod 'Realm', git: 'https://...', branch: 'main'           -> latest (Ruby 1.9+ hash syntax)
//   pod 'Firebase/Core', '7.0'                                -> Firebase/Core
//   pod "Quotes", "1.0"                                       -> double-quote variant
//
// `target 'Name' do ... end` blocks just nest pods; the parser doesn't track the target
// context (each `pod` line is emitted independently with its raw source location).

var (
	// podLine: pod 'Name', 'version' OR pod 'Name' (version optional, may have trailing :git=>...)
	// Group 1: pod name (single or double quoted)
	// Group 2: version string (optional; only present if a second quoted string follows the name)
	podLinePattern = regexp.MustCompile(
		`^\s*pod\s+(?:'([^']+)'|"([^"]+)")` +
			`(?:\s*,\s*(?:'([^']*)'|"([^"]*)"))?`,
	)
	// gitRefPattern detects :git=>, :branch=>, :tag=>, :commit=>, :podspec=>, :path=>
	// or modern hash syntax (git:, branch:, etc). Matching any of these means
	// the version is not a concrete semver; resolve to "latest".
	gitRefPattern = regexp.MustCompile(
		`(?::(git|branch|tag|commit|podspec|path)\s*=>|(?:^|\s)(git|branch|tag|commit|podspec|path)\s*:)`,
	)
)

// parsePodfile parses a Podfile and returns one Package per `pod` line.
func parsePodfile(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	var packages []models.Package

	for i, raw := range lines {
		code := stripInlineComment(raw)
		trimmed := strings.TrimSpace(code)
		if trimmed == "" || !strings.HasPrefix(trimmed, "pod ") && !strings.HasPrefix(trimmed, "pod\t") {
			continue
		}

		match := podLinePattern.FindStringSubmatch(code)
		if match == nil {
			continue
		}

		name := firstNonEmpty(match[1], match[2])
		if name == "" {
			continue
		}

		version := firstNonEmpty(match[3], match[4])
		// :git => / :branch => / :tag => etc anywhere on the line forces "latest"
		if gitRefPattern.MatchString(code) {
			version = ""
		}

		startIdx, endIdx := lineExtent(raw, strings.TrimRight(code, " \t"))
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        resolveVersion(version),
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
