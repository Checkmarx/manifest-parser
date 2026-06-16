package carthage

import (
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestCartfile_Fixture(t *testing.T) {
	parser := &CarthageParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Cartfile"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)
	cases := map[string]string{
		"Alamofire/Alamofire":                         "5.2.0",  // quoted exact tag
		"ReactiveCocoa/ReactiveSwift":                 "6.7.0",  // == operator
		"onevcat/Kingfisher":                          "latest", // ~> range
		"SnapKit/SnapKit":                             "latest", // >= range
		"Quick/Quick":                                 "latest", // no spec
		"ReactiveX/RxSwift":                           "latest", // branch ref
		"MyFramework":                                 "1.4.2",  // git origin (last path component, .git stripped)
	}
	// Binary origin uses .json suffix stripped — should also appear.
	cases["MyFramework"] = "1.4.2" // already set; binary appears separately via different parse

	// Verify each expected package
	for name, wantVer := range cases {
		assertPkg(t, got, name, "carthage", wantVer)
	}

	// Binary entry — note both git and binary in fixture happen to produce the same
	// derived name "MyFramework". The git one is pinned 1.4.2; the binary one is
	// ranged ~> 2.3 -> latest. The parser emits both Locations as separate Packages.
	binaryCount := 0
	gitCount := 0
	for _, p := range pkgs {
		if p.PackageName == "MyFramework" {
			switch p.Version {
			case "1.4.2":
				gitCount++
			case "latest":
				binaryCount++
			}
		}
	}
	if gitCount != 1 {
		t.Errorf("expected 1 MyFramework@1.4.2 (git origin), got %d", gitCount)
	}
	if binaryCount != 1 {
		t.Errorf("expected 1 MyFramework@latest (binary origin), got %d", binaryCount)
	}
}

func TestCartfilePrivate_Fixture(t *testing.T) {
	parser := &CarthageParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Cartfile.private"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)
	cases := map[string]string{
		"Quick/Nimble":                            "9.2.0",
		"pointfreeco/swift-snapshot-testing":      "latest",
		"TestUtils":                               "latest", // git origin, branch ref "main"
	}

	if len(pkgs) != len(cases) {
		t.Errorf("expected %d entries, got %d", len(cases), len(pkgs))
	}

	for name, wantVer := range cases {
		assertPkg(t, got, name, "carthage", wantVer)
	}
}

func TestCartfileResolved_Fixture(t *testing.T) {
	parser := &CarthageParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Cartfile.resolved"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)
	cases := map[string]string{
		"Alamofire/Alamofire":         "5.2.0",
		"ReactiveCocoa/ReactiveSwift": "6.7.0",
		"onevcat/Kingfisher":          "7.6.2",
		"SnapKit/SnapKit":             "5.6.0",
		"Quick/Quick":                 "v4.0.0",                  // tag-prefixed semver — accepted as-is by looksLikeSemver
		"ReactiveX/RxSwift":           "latest",                  // 19-hex commit SHA — not semver
	}
	// MyFramework appears twice (git + binary), both pinned.
	for name, wantVer := range cases {
		assertPkg(t, got, name, "carthage", wantVer)
	}

	// MyFramework: count distinct lines
	myFrameworkCount := 0
	for _, p := range pkgs {
		if p.PackageName == "MyFramework" && (p.Version == "1.4.2" || p.Version == "2.3.1") {
			myFrameworkCount++
		}
	}
	if myFrameworkCount != 2 {
		t.Errorf("expected 2 MyFramework entries in resolved file, got %d", myFrameworkCount)
	}
}

func TestPackageNameFromSource(t *testing.T) {
	cases := []struct {
		origin, source, want string
	}{
		{"github", "Alamofire/Alamofire", "Alamofire/Alamofire"},
		{"git", "https://example.com/internal/MyFramework.git", "MyFramework"},
		{"git", "git@server:group/SubMod.git", "SubMod"},
		{"binary", "https://my.domain.com/release/MyFramework.json", "MyFramework"},
	}
	for _, c := range cases {
		if got := packageNameFromSource(c.origin, c.source); got != c.want {
			t.Errorf("packageNameFromSource(%q, %q) = %q, want %q",
				c.origin, c.source, got, c.want)
		}
	}
}

func TestResolveVersion(t *testing.T) {
	cases := map[string]string{
		"":            "latest",
		`"1.2.3"`:     "1.2.3",
		`"main"`:      "latest",
		`"abc12345"`:  "latest", // looks like a SHA, not semver
		"== 1.2.3":    "1.2.3",
		"==1.2.3":     "1.2.3",
		"~> 1.2":      "latest",
		">= 1.0":      "latest",
		`"v4.0.0"`:    "v4.0.0", // v-prefix is a common Git tag convention — accepted
		`"1.0.0-rc1"`: "1.0.0-rc1",
	}
	for in, want := range cases {
		if got := resolveVersion(in); got != want {
			t.Errorf("resolveVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

// --- helpers ---------------------------------------------------------------

func indexByName(pkgs []models.Package) map[string]models.Package {
	out := make(map[string]models.Package, len(pkgs))
	for _, p := range pkgs {
		// For duplicate names (e.g. MyFramework via git+binary), keep first.
		if _, exists := out[p.PackageName]; !exists {
			out[p.PackageName] = p
		}
	}
	return out
}

func assertPkg(t *testing.T, got map[string]models.Package, name, pm, ver string) {
	t.Helper()
	p, ok := got[name]
	if !ok {
		t.Errorf("expected package %q not found", name)
		return
	}
	if p.PackageManager != pm {
		t.Errorf("%s: PackageManager = %q, want %q", name, p.PackageManager, pm)
	}
	if p.Version != ver {
		t.Errorf("%s: Version = %q, want %q", name, p.Version, ver)
	}
	if p.FilePath == "" {
		t.Errorf("%s: FilePath not set", name)
	}
	if len(p.Locations) == 0 {
		t.Errorf("%s: no Locations", name)
		return
	}
	for i, loc := range p.Locations {
		if loc.EndIndex <= loc.StartIndex {
			t.Errorf("%s loc[%d]: EndIndex (%d) must be > StartIndex (%d)",
				name, i, loc.EndIndex, loc.StartIndex)
		}
	}
}
