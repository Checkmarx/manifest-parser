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

	// Expected versions from Cartfile.resolved (lock file), not manifest ranges.
	cases := map[string]string{
		"Alamofire/Alamofire":         "5.2.0",  // quoted exact tag, matches lock
		"ReactiveCocoa/ReactiveSwift": "6.7.0",  // == operator, matches lock
		"onevcat/Kingfisher":          "7.6.2",  // manifest: ~> 7.0 (latest), lock: 7.6.2
		"SnapKit/SnapKit":             "5.6.0",  // manifest: >= 5.0.0 (latest), lock: 5.6.0
		"Quick/Quick":                 "v4.0.0", // manifest: no spec (latest), lock: v4.0.0
		"ReactiveX/RxSwift":           "latest", // branch ref, lock has SHA (not semver)
		"MyFramework":                 "1.4.2",  // Git origin (exact version from manifest, not overridden by lock)
	}

	// Verify each expected package
	for name, wantVer := range cases {
		assertPkg(t, got, name, "carthage", wantVer)
	}

	// Note: Both git and binary origins produce "MyFramework" as the package name.
	// The git version is 1.4.2 (pinned, matches lock).
	// The binary version is ranged ~> 2.3, which resolves to 2.3.1 from lock.
	// indexByName keeps only the first occurrence, so we count separately.
	gitCount := 0
	binaryCount := 0
	for _, p := range pkgs {
		if p.PackageName == "MyFramework" {
			switch p.Version {
			case "1.4.2":
				gitCount++
			case "2.3.1":
				binaryCount++
			}
		}
	}
	if gitCount != 1 {
		t.Errorf("expected 1 MyFramework@1.4.2 (git origin), got %d", gitCount)
	}
	if binaryCount != 1 {
		t.Errorf("expected 1 MyFramework@2.3.1 (binary origin), got %d", binaryCount)
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
		"Quick/Nimble":                       "9.2.0",
		"pointfreeco/swift-snapshot-testing": "latest",
		"TestUtils":                          "latest", // git origin, branch ref "main"
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
		"Quick/Quick":                 "v4.0.0", // tag-prefixed semver — accepted as-is by looksLikeSemver
		"ReactiveX/RxSwift":           "latest", // 19-hex commit SHA — not semver
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

func TestCartfile_WithLockFile(t *testing.T) {
	// Verify that when both Cartfile and Cartfile.resolved are present,
	// exact versions from the lock file override ranges from the manifest.
	parser := &CarthageParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Cartfile"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := indexByName(pkgs)

	// Expected versions from Cartfile.resolved (lock file), not from Cartfile ranges.
	// Cartfile has: ~> ranges, >= ranges, no spec, branch refs
	// Cartfile.resolved has exact versions for all.
	lockExpectations := map[string]string{
		"Alamofire/Alamofire":         "5.2.0",  // quoted exact in both
		"ReactiveCocoa/ReactiveSwift": "6.7.0",  // == in manifest, lock has 6.7.0
		"onevcat/Kingfisher":          "7.6.2",  // ~> range in manifest, lock has 7.6.2
		"SnapKit/SnapKit":             "5.6.0",  // >= range in manifest, lock has 5.6.0
		"Quick/Quick":                 "v4.0.0", // no spec in manifest, lock has v4.0.0
		"ReactiveX/RxSwift":           "latest", // branch ref in manifest, lock has SHA (not semver -> latest)
		// MyFramework: both git and binary origins are parsed separately
	}

	for name, wantVer := range lockExpectations {
		if p, ok := got[name]; !ok {
			t.Errorf("expected package %q not found", name)
		} else if p.Version != wantVer {
			t.Errorf("%s: version = %q, want %q (lock file should override manifest range)",
				name, p.Version, wantVer)
		}
	}

	// MyFramework from both git and binary origins should use lock file versions
	myFrameworkVersions := make(map[string]bool)
	for _, p := range pkgs {
		if p.PackageName == "MyFramework" {
			myFrameworkVersions[p.Version] = true
		}
	}

	// Cartfile.resolved has both: git "1.4.2" and binary "2.3.1"
	if !myFrameworkVersions["1.4.2"] {
		t.Errorf("MyFramework git version should be 1.4.2 from lock file, got versions: %v",
			myFrameworkVersions)
	}
	if !myFrameworkVersions["2.3.1"] {
		t.Errorf("MyFramework binary version should be 2.3.1 from lock file, got versions: %v",
			myFrameworkVersions)
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
