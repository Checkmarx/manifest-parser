// Package swiftpm parses Swift Package Manager manifests.
//
// Two file formats are supported:
//   - Package.swift / Package@swift-X.Y.swift — Swift DSL manifest
//   - Package.resolved                        — JSON lock file (v1 and v2)
package swiftpm

import (
	"fmt"
	"path/filepath"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

const packageManagerName = "swift"

// SwiftPmParser dispatches to the manifest or lock-file parser based on the filename.
type SwiftPmParser struct{}

// Parse implements the Parser interface.
func (p *SwiftPmParser) Parse(manifestFile string) ([]models.Package, error) {
	name := filepath.Base(manifestFile)
	if name == "Package.resolved" {
		return parseResolved(manifestFile)
	}
	return parsePackageSwift(manifestFile)
}

// errReadFile wraps file-read errors with a consistent prefix.
func errReadFile(err error) error {
	return fmt.Errorf("failed to read manifest file: %w", err)
}
