package swiftpm

import (
	"path/filepath"
	"testing"
)

func TestPackageResolved_V2(t *testing.T) {
	parser := &SwiftPmParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Package.resolved"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := make(map[string]string, len(pkgs))
	for _, p := range pkgs {
		got[p.PackageName] = p.Version
		if p.PackageManager != "swift" {
			t.Errorf("%s: PackageManager = %q, want \"swift\"", p.PackageName, p.PackageManager)
		}
		if len(p.Locations) == 0 || p.Locations[0].Line == 0 {
			t.Errorf("%s: Location.Line should be populated, got %+v", p.PackageName, p.Locations)
		}
	}

	cases := map[string]string{
		"alamofire":         "5.2.0",
		"kingfisher":        "7.6.2",
		"snapkit":           "5.6.0",
		"swift-collections": "latest", // branch only, no version
		"swift-crypto":      "2.0.6",
		"swift-log":         "1.4.4",
		"swift-nio":         "2.42.0",
	}
	for name, wantVer := range cases {
		if ver, ok := got[name]; !ok {
			t.Errorf("expected pin %q not found", name)
		} else if ver != wantVer {
			t.Errorf("%s: version = %q, want %q", name, ver, wantVer)
		}
	}
}

func TestPackageResolved_V1(t *testing.T) {
	parser := &SwiftPmParser{}
	// Path the parser is given matters only for FilePath; the v1 fixture is
	// named `.json` so it does not collide with the routing rule for
	// "Package.resolved". We can still parse it via parseResolved directly.
	pkgs, err := parseResolved(filepath.Join("..", "..", "..", "internal", "testdata", "Package.resolved.v1.json"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	_ = parser // silence "declared but not used"

	got := make(map[string]string, len(pkgs))
	for _, p := range pkgs {
		got[p.PackageName] = p.Version
	}

	cases := map[string]string{
		"Alamofire":        "5.2.0",
		"swift-nio":        "2.42.0",
		"SwiftCollections": "latest", // branch only, no version
	}
	for name, wantVer := range cases {
		if ver, ok := got[name]; !ok {
			t.Errorf("expected pin %q not found", name)
		} else if ver != wantVer {
			t.Errorf("%s: version = %q, want %q", name, ver, wantVer)
		}
	}
}
