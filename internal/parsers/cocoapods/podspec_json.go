package cocoapods

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// *.podspec.json shape (CocoaPods trunk-published spec):
//
//   {
//     "name": "MyPod",
//     "version": "1.0.0",
//     "dependencies": {
//       "Alamofire": ["~> 5.0"],
//       "SwiftyJSON": ["4.3.0"],
//       "Quick": []
//     }
//   }
//
// Dependency values are arrays of version specifiers (often a single element,
// sometimes empty). When any specifier in the array is ranged or the array is
// empty, we resolve to "latest".

type podspecJSON struct {
	Dependencies map[string][]string `json:"dependencies"`
}

func parsePodspecJSON(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	var doc podspecJSON
	if err := json.Unmarshal(content, &doc); err != nil {
		return nil, err
	}

	lines := splitLinesCRLF(string(content))
	keyPattern := regexp.MustCompile(`"([^"]+)"\s*:`)

	var packages []models.Package
	for name, specs := range doc.Dependencies {
		version := pickConcreteVersion(specs)
		loc := findDependencyKeyLocation(lines, name, keyPattern)
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        version,
			FilePath:       manifestFile,
			Locations:      []models.Location{loc},
		})
	}
	return packages, nil
}

// pickConcreteVersion returns the first version specifier that isn't ranged.
// Empty array or all-ranged specifiers resolve to "latest".
func pickConcreteVersion(specs []string) string {
	for _, s := range specs {
		if v := resolveVersion(s); v != "latest" {
			return v
		}
	}
	return "latest"
}

// findDependencyKeyLocation finds the line where `"name":` appears inside the
// dependencies object and returns a precise Location.
func findDependencyKeyLocation(lines []string, name string, keyPattern *regexp.Regexp) models.Location {
	for i, raw := range lines {
		if !strings.Contains(raw, name) {
			continue
		}
		match := keyPattern.FindStringSubmatchIndex(raw)
		if match == nil {
			continue
		}
		if raw[match[2]:match[3]] != name {
			continue
		}
		startIdx := len(raw) - len(strings.TrimLeft(raw, " \t"))
		endIdx := len(strings.TrimRight(raw, " \t,"))
		return models.Location{Line: i, StartIndex: startIdx, EndIndex: endIdx}
	}
	return models.Location{}
}
