package cocoapods

import (
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// Podfile.lock format (CocoaPods uses pseudo-YAML — we parse line by line):
//
//   PODS:
//     - Alamofire (5.2.0)
//     - Realm (10.20.0):
//       - Realm/Headers (= 10.20.0)
//     - Realm/Headers (10.20.0)
//     - SwiftyJSON (4.3.0)
//
//   DEPENDENCIES:
//     - Alamofire (= 5.2.0)
//     - Realm
//     - SwiftyJSON (~> 4.0)
//
//   SPEC REPOS:
//     ...
//
// We only mine the PODS: section. CocoaPods indents pods at exactly 2 spaces
// under PODS:; nested transitive declarations are at 4+ spaces. Matching the
// 2-space indent precisely lets us:
//   - keep legitimate top-level subspec deps like "Firebase/Core (7.0.0)"
//   - skip transitive lines like "    - FirebaseAnalytics (~> 7.0.0)"
//     (which would otherwise capture the ranged version before the resolved one)

var podsLockEntryPattern = regexp.MustCompile(
	`^  -\s+([^\s(]+)\s*\(([^)]+)\)\s*:?\s*$`,
)

func parsePodfileLock(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	var packages []models.Package
	inPods := false
	seen := map[string]bool{}

	for i, raw := range lines {
		trimmed := strings.TrimSpace(raw)

		// Section detection: a top-level key (no leading whitespace) ending with ':'.
		if len(raw) > 0 && raw[0] != ' ' && raw[0] != '\t' && strings.HasSuffix(trimmed, ":") {
			inPods = trimmed == "PODS:"
			continue
		}
		if !inPods {
			continue
		}

		match := podsLockEntryPattern.FindStringSubmatch(raw)
		if match == nil {
			continue
		}
		name, version := match[1], strings.TrimSpace(match[2])
		if seen[name] {
			continue // PODS section can list the same pod twice (defensive)
		}
		seen[name] = true

		startIdx, endIdx := lineExtent(raw, strings.TrimRight(raw, " \t"))
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
