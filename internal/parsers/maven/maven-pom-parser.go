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

// findDependencyLocations finds all locations for a dependency in the POM file
// Returns all lines from <dependency> to </dependency> inclusive, excluding comments
func findDependencyLocations(lines []string, dep MavenDependency) []models.Location {
	var locations []models.Location

	// Search for the dependency block
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for the beginning of a dependency block
		if strings.Contains(line, "<dependency>") {
			dependencyStartLine := i
			var foundGroupId, foundArtifactId bool
			var dependencyEndLine int

			// First, find the end of this dependency block
			for j := i + 1; j < len(lines) && j < i+15; j++ { // Limit to 15 lines
				innerLine := strings.TrimSpace(lines[j])
				if strings.Contains(innerLine, "</dependency>") {
					dependencyEndLine = j
					break
				}
			}

			// If we didn't find the end, skip this block
			if dependencyEndLine == 0 {
				continue
			}

			// Now scan only within this dependency block to check if it matches our target dependency
			for j := i + 1; j < dependencyEndLine; j++ {
				innerLine := strings.TrimSpace(lines[j])

				// Look for matching groupId within this block only
				if strings.Contains(innerLine, "<groupId>") && strings.Contains(innerLine, "</groupId>") {
					// Extract the actual groupId value
					start := strings.Index(innerLine, "<groupId>") + len("<groupId>")
					end := strings.Index(innerLine, "</groupId>")
					if start < end {
						actualGroupId := strings.TrimSpace(innerLine[start:end])
						if actualGroupId == dep.GroupId {
							foundGroupId = true
						}
					}
				}

				// Look for matching artifactId within this block only
				if strings.Contains(innerLine, "<artifactId>") && strings.Contains(innerLine, "</artifactId>") {
					// Extract the actual artifactId value
					start := strings.Index(innerLine, "<artifactId>") + len("<artifactId>")
					end := strings.Index(innerLine, "</artifactId>")
					if start < end {
						actualArtifactId := strings.TrimSpace(innerLine[start:end])
						if actualArtifactId == dep.ArtifactId {
							foundArtifactId = true
						}
					}
				}
			}

			// If both groupId and artifactId match in this dependency block, collect all lines
			if foundGroupId && foundArtifactId {
				// Add the opening <dependency> line
				startIdx := strings.Index(lines[dependencyStartLine], "<dependency>")
				if startIdx == -1 {
					startIdx = 0
				}
				locations = append(locations, models.Location{
					Line:       dependencyStartLine,
					StartIndex: startIdx,
					EndIndex:   len(lines[dependencyStartLine]),
				})

				// Add all intermediate lines (skip comments)
				for m := dependencyStartLine + 1; m <= dependencyEndLine; m++ {
					currentLine := lines[m]
					trimmedLine := strings.TrimSpace(currentLine)

					// Check if this is the closing </dependency> line
					if strings.Contains(currentLine, "</dependency>") {
						closingStartIdx := strings.Index(currentLine, "</dependency>")
						closingEndIdx := strings.Index(currentLine, "</dependency>") + len("</dependency>")
						locations = append(locations, models.Location{
							Line:       m,
							StartIndex: closingStartIdx,
							EndIndex:   closingEndIdx,
						})
						break
					}

					// Skip comment lines
					if strings.HasPrefix(trimmedLine, "<!--") {
						continue
					}

					// Add intermediate line (find meaningful content, not just whitespace)
					if trimmedLine != "" {
						// Find the start of actual content (skip leading whitespace)
						contentStartIdx := strings.Index(currentLine, strings.TrimLeft(trimmedLine, " \t"))
						if contentStartIdx == -1 {
							contentStartIdx = 0
						}
						locations = append(locations, models.Location{
							Line:       m,
							StartIndex: contentStartIdx,
							EndIndex:   len(currentLine),
						})
					}
				}

				return locations
			}

			// Move to the end of this dependency block to continue searching
			i = dependencyEndLine
		}
	}

	// If not found, return empty slice
	return []models.Location{}
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
		locations := findDependencyLocations(lines, dep)

		// Create package entry
		packages = append(packages, models.Package{
			PackageManager: "mvn",
			PackageName:    dep.GroupId + ":" + dep.ArtifactId,
			Version:        resolveVersion(dep.Version, props, project.DependencyManagement.Dependencies, dep.GroupId, dep.ArtifactId),
			FilePath:       manifestFile,
			Locations:      locations,
		})
	}

	return packages, nil
}
