package swiftpm

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

type resolvedFile struct {
	Object  *resolvedObject `json:"object,omitempty"`
	Pins    []resolvedPin   `json:"pins,omitempty"`
	Version int             `json:"version"`
}

type resolvedObject struct {
	Pins []resolvedPin `json:"pins"`
}

type resolvedPin struct {
	Package       string `json:"package,omitempty"`
	RepositoryURL string `json:"repositoryURL,omitempty"`
	Identity      string `json:"identity,omitempty"`
	Kind          string `json:"kind,omitempty"`
	Location      string `json:"location,omitempty"`
	State         resolvedState `json:"state"`
}

type resolvedState struct {
	Version  string `json:"version,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Revision string `json:"revision,omitempty"`
}

func (p resolvedPin) name() string {
	if p.Identity != "" {
		return p.Identity
	}
	return p.Package
}

func (p resolvedPin) version() string {
	if p.State.Version != "" {
		return p.State.Version
	}
	return "latest"
}

// parseResolved decodes a Package.resolved file and returns one Package per pin.
func parseResolved(manifestFile string) ([]models.Package, error) {
	content, err := os.ReadFile(manifestFile)
	if err != nil {
		return nil, errReadFile(err)
	}

	var doc resolvedFile
	if err := json.Unmarshal(content, &doc); err != nil {
		return nil, err
	}

	pins := doc.Pins
	if doc.Object != nil && len(doc.Object.Pins) > 0 {
		pins = doc.Object.Pins
	}

	lines := splitLinesCRLF(string(content))
	identityField := regexp.MustCompile(`"(?:identity|package)"\s*:\s*"([^"]+)"`)

	var packages []models.Package
	for _, pin := range pins {
		name := pin.name()
		if name == "" {
			continue
		}
		loc := findPinLocation(lines, name, identityField)
		packages = append(packages, models.Package{
			PackageManager: packageManagerName,
			PackageName:    name,
			Version:        pin.version(),
			FilePath:       manifestFile,
			Locations:      []models.Location{loc},
		})
	}

	return packages, nil
}

// findPinLocation scans the raw JSON for the line that declares this pin's
// identity/package field and returns a Location pointing at it. Falls back to
// a zero-value Location if nothing matches (shouldn't happen for well-formed input).
func findPinLocation(lines []string, name string, identityField *regexp.Regexp) models.Location {
	for i, raw := range lines {
		if !strings.Contains(raw, name) {
			continue
		}
		match := identityField.FindStringSubmatchIndex(raw)
		if match == nil {
			continue
		}
		// match[2]/match[3] are the bounds of the captured name.
		if raw[match[2]:match[3]] != name {
			continue
		}
		startIdx := len(raw) - len(strings.TrimLeft(raw, " \t"))
		endIdx := len(strings.TrimRight(raw, " \t,"))
		return models.Location{Line: i, StartIndex: startIdx, EndIndex: endIdx}
	}
	return models.Location{}
}

// splitLinesCRLF splits on \n and strips trailing \r so byte offsets stay correct
// on Windows CRLF files.
func splitLinesCRLF(s string) []string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines
}
