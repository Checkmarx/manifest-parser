package package_json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Checkmarx/manifest-parser/internal"
	"os"
	"strings"
)

type NpmPackageJsonParser struct{}

func (p *NpmPackageJsonParser) Parse(manifestFile string) ([]internal.Package, error) {
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Read the file line by line and build a map of line numbers.
	lineNumbers := make(map[string]int)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		// Extract key from lines like "  \"<package-name>\": \"<version>\","
		if strings.Contains(line, ":") && strings.Contains(line, "\"") {
			line = strings.TrimSpace(line)
			parts := strings.Split(line, ":")
			packageName := strings.Trim(parts[0], "\"")
			lineNumbers[packageName] = lineNum
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Reset the file pointer for JSON decoding.
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Decode JSON
	var packageJSON map[string]interface{}
	if err := json.NewDecoder(file).Decode(&packageJSON); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract dependencies and devDependencies.
	var packages []internal.Package
	for _, key := range []string{"dependencies", "devDependencies"} {
		if deps, ok := packageJSON[key].(map[string]interface{}); ok {
			for pkg, ver := range deps {
				packages = append(packages, internal.Package{
					PackageName: pkg,
					Version:     fmt.Sprintf("%v", ver),
					LineStart:   lineNumbers[pkg],
					LineEnd:     lineNumbers[pkg],
					Filepath:    manifestFile,
				})
			}
		}
	}

	return packages, nil
}
