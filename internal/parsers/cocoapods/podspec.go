package cocoapods

import (
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// .podspec is Ruby DSL like:
//
//   Pod::Spec.new do |s|
//     s.name             = 'MyPod'
//     s.version          = '1.0.0'
//     s.dependency 'Alamofire', '~> 5.0'
//     s.dependency 'SwiftyJSON', '4.3.0'
//     s.dependency 'Quick'
//   end
//
// The leading receiver name (`s` here) is arbitrary — some specs use `spec`, `pod`,
// or `cocoapod`. We match `<receiver>.dependency`. Multiple version arguments
// (CocoaPods allows up to two) collapse to "latest" when ranged.

var podspecDependencyPattern = regexp.MustCompile(
	`^\s*\w+\.dependency\s+(?:'([^']+)'|"([^"]+)")` +
		`(?:\s*,\s*(?:'([^']*)'|"([^"]*)"))?` +
		`(?:\s*,\s*(?:'([^']*)'|"([^"]*)"))?`,
)

func parsePodspec(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	lines := splitLinesCRLF(string(content))
	var packages []models.Package

	for i, raw := range lines {
		code := stripInlineComment(raw)
		match := podspecDependencyPattern.FindStringSubmatch(code)
		if match == nil {
			continue
		}
		name := firstNonEmpty(match[1], match[2])
		if name == "" {
			continue
		}
		// Use the first version spec; if it's a range and a second spec is present,
		// both will start with operators so the result is still "latest".
		version := firstNonEmpty(match[3], match[4])

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
