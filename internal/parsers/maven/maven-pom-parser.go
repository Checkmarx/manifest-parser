package maven

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

// MavenPomParser implements parsing of Maven POM files
type MavenPomParser struct{}

// MavenDependency represents a dependency in the POM file
type MavenDependency struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
}

// xmlProperty represents a single property under <properties>
type xmlProperty struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// MavenProject represents the POM file structure
type MavenProject struct {
	GroupId              string            `xml:"groupId"`
	ArtifactId           string            `xml:"artifactId"`
	Version              string            `xml:"version"`
	Dependencies         []MavenDependency `xml:"dependencies>dependency"`
	DependencyManagement struct {
		Dependencies []MavenDependency `xml:"dependencies>dependency"`
	} `xml:"dependencyManagement"`
	Properties struct {
		Entries []xmlProperty `xml:",any"`
	} `xml:"properties"`
}

// resolveVersion replaces ${...} variables from <properties>, handles version ranges,
// and looks up managed versions
func resolveVersion(raw string, props map[string]string, managedDeps []MavenDependency, groupId, artifactId string) string {
	// First, resolve property variables
	if strings.HasPrefix(raw, "${") && strings.HasSuffix(raw, "}") {
		key := strings.TrimSuffix(strings.TrimPrefix(raw, "${"), "}")
		if resolved, exists := props[key]; exists {
			raw = resolved
		}
	}

	// If version is empty or contains range chars, try to find in managed dependencies
	if raw == "" || strings.ContainsAny(raw, "[]()^~*><") {
		for _, managedDep := range managedDeps {
			if managedDep.GroupId == groupId && managedDep.ArtifactId == artifactId {
				if managedDep.Version != "" {
					return resolveVersion(managedDep.Version, props, nil, "", "")
				}
			}
		}
		return "latest"
	}

	return raw
}

// findDependencyLocation finds the exact location of a dependency in the POM file
// Returns line number (0-based), start index, and end index of the artifactId line
func findDependencyLocation(lines []string, dep MavenDependency) (lineNum, startIdx, endIdx int) {
	// Search for the dependency block
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for the beginning of a dependency block
		if strings.Contains(line, "<dependency>") {
			// Now search for the groupId in the following lines
			for j := i + 1; j < len(lines) && j < i+10; j++ { // Limit to 10 lines
				innerLine := strings.TrimSpace(lines[j])

				// Check if this is the end of the dependency
				if strings.Contains(innerLine, "</dependency>") {
					break
				}

				// Look for matching groupId
				if strings.Contains(innerLine, "<groupId>") && strings.Contains(innerLine, dep.GroupId) {
					// Now verify that the artifactId also matches in the same dependency block
					for k := j + 1; k < len(lines) && k < i+10; k++ {
						artifactLine := strings.TrimSpace(lines[k])

						if strings.Contains(artifactLine, "</dependency>") {
							break
						}

						if strings.Contains(artifactLine, "<artifactId>") && strings.Contains(artifactLine, dep.ArtifactId) {
							// Found it! Return the location of the artifactId line
							lineNum = k // 0-based line numbers

							// Find the start of the <artifactId> tag (0-based column)
							startIdx = strings.Index(lines[k], "<artifactId>")
							if startIdx == -1 {
								// Fallback to start of line if tag not found
								startIdx = 0
							}

							// Find the end of the closing </artifactId> tag
							if closingTagPos := strings.Index(lines[k], "</artifactId>"); closingTagPos != -1 {
								endIdx = closingTagPos + len("</artifactId>")
							} else {
								// Fallback to end of line
								endIdx = len(lines[k])
							}

							return lineNum, startIdx, endIdx
						}
					}
					break
				}
			}
		}
	}

	// If not found, return 0
	return 0, 0, 0
}

// Parse implements the Parser interface for Maven POM files
func (p *MavenPomParser) Parse(manifestFile string) ([]models.Package, error) {
	// Read the POM file content
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse XML content into MavenProject struct
	var project MavenProject
	if err := xml.Unmarshal(content, &project); err != nil {
		return nil, fmt.Errorf("failed to parse POM file: %w", err)
	}

	// Extract properties to map for variable resolution
	props := make(map[string]string)
	for _, entry := range project.Properties.Entries {
		props[entry.XMLName.Local] = strings.TrimSpace(entry.Value)
	}

	var packages []models.Package
	lines := strings.Split(string(content), "\n")

	// Process only direct dependencies (not managed ones to avoid duplicates)
	allDeps := project.Dependencies

	// Process each dependency
	for _, dep := range allDeps {
		// Use the enhanced location finding function
		lineNum, startIdx, endIdx := findDependencyLocation(lines, dep)

		// Create package entry
		packages = append(packages, models.Package{
			PackageManager: "maven",
			PackageName:    dep.GroupId + ":" + dep.ArtifactId,
			Version:        resolveVersion(dep.Version, props, project.DependencyManagement.Dependencies, dep.GroupId, dep.ArtifactId),
			FilePath:       manifestFile,
			Locations: []models.Location{{
				Line:       lineNum,
				StartIndex: startIdx,
				EndIndex:   endIdx,
			}},
		})
	}

	return packages, nil
}
