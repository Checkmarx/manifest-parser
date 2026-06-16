package cocoapods

import (
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestPodfile_Fixture(t *testing.T) {
	parser := &CocoaPodsParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Podfile"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)

	cases := map[string]string{
		"Alamofire":     "5.2.0",  // pinned single-quote
		"Nimble":        "9.2.0",  // pinned double-quote
		"Firebase/Core": "7.0.0",  // subspec form
		"Realm":         "latest", // ~> range
		"SwiftyJSON":    "latest", // >= range
		"Quick":         "latest", // no version specified
		"Kingfisher":    "latest", // :git => ref
		"SDWebImage":    "latest", // modern hash syntax git ref
	}

	if len(pkgs) != len(cases) {
		t.Errorf("expected %d pods, got %d", len(cases), len(pkgs))
	}

	for name, wantVer := range cases {
		assertPkg(t, got, name, "cocoapods", wantVer)
	}
}

func TestPodfileLock_Fixture(t *testing.T) {
	parser := &CocoaPodsParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Podfile.lock"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)

	cases := map[string]string{
		"Alamofire":         "5.2.0",
		"Firebase/Core":     "7.0.0",
		"FirebaseAnalytics": "7.0.0",
		"Kingfisher":        "7.6.2",
		"Nimble":            "9.2.0",
		"Quick":             "4.0.0",
		"Realm":             "10.20.0",
		"SDWebImage":        "5.12.0",
		"SwiftyJSON":        "4.3.0",
	}

	if len(pkgs) != len(cases) {
		t.Errorf("expected %d pods, got %d", len(cases), len(pkgs))
	}

	for name, wantVer := range cases {
		assertPkg(t, got, name, "cocoapods", wantVer)
	}

	// Realm/Headers (sub-pod) must NOT appear separately — it would duplicate Realm.
	if _, exists := got["Realm/Headers"]; exists {
		t.Errorf("sub-pod Realm/Headers should not be emitted separately")
	}
}

func TestPodspec_Fixture(t *testing.T) {
	parser := &CocoaPodsParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "internal", "testdata", "TestPod.podspec"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)

	cases := map[string]string{
		"Alamofire":  "latest", // ~> range
		"SwiftyJSON": "4.3.0",  // pinned double-quote
		"Quick":      "latest", // no version
		"Nimble":     "latest", // two ranged args
	}

	if len(pkgs) != len(cases) {
		t.Errorf("expected %d dependencies, got %d", len(cases), len(pkgs))
	}

	for name, wantVer := range cases {
		assertPkg(t, got, name, "cocoapods", wantVer)
	}
}

func TestPodspecJSON_Fixture(t *testing.T) {
	parser := &CocoaPodsParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "internal", "testdata", "TestPod.podspec.json"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)

	cases := map[string]string{
		"Alamofire":  "latest", // ~> range
		"SwiftyJSON": "4.3.0",  // pinned
		"Quick":      "latest", // empty array
		"Nimble":     "latest", // all-ranged array
	}

	if len(pkgs) != len(cases) {
		t.Errorf("expected %d dependencies, got %d", len(cases), len(pkgs))
	}

	for name, wantVer := range cases {
		assertPkg(t, got, name, "cocoapods", wantVer)
	}
}

func TestResolveVersion(t *testing.T) {
	cases := map[string]string{
		"":        "latest",
		"5.2.0":   "5.2.0",
		"~> 5.0":  "latest",
		">= 4.0":  "latest",
		"< 10.0":  "latest",
		"= 5.2.0": "latest",
		"!= 6.0":  "latest",
	}
	for in, want := range cases {
		if got := resolveVersion(in); got != want {
			t.Errorf("resolveVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStripInlineComment(t *testing.T) {
	cases := []struct{ in, out string }{
		{`pod 'Alamofire', '5.0' # pinned`, `pod 'Alamofire', '5.0' `},
		{`pod 'Nimble', '~> 9.0' # ranged # second hash`, `pod 'Nimble', '~> 9.0' `},
		{`pod 'NameWith#Hash'`, `pod 'NameWith#Hash'`},
		{`# whole line is a comment`, ``},
	}
	for _, c := range cases {
		if got := stripInlineComment(c.in); got != c.out {
			t.Errorf("stripInlineComment(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

// --- helpers ---------------------------------------------------------------

func indexByName(pkgs []models.Package) map[string]models.Package {
	out := make(map[string]models.Package, len(pkgs))
	for _, p := range pkgs {
		out[p.PackageName] = p
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
