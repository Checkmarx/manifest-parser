package swiftpm

import (
	"path/filepath"
	"testing"
)

func TestPackageSwift_FixtureFile(t *testing.T) {
	parser := &SwiftPmParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Package.swift"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Build a name -> version map for assertion convenience.
	got := make(map[string]string, len(pkgs))
	for _, p := range pkgs {
		got[p.PackageName] = p.Version
		if p.PackageManager != "swift" {
			t.Errorf("%s: PackageManager = %q, want \"swift\"", p.PackageName, p.PackageManager)
		}
		if p.FilePath == "" {
			t.Errorf("%s: FilePath not set", p.PackageName)
		}
		if len(p.Locations) == 0 {
			t.Errorf("%s: no Locations", p.PackageName)
			continue
		}
		// Every Location should have a populated EndIndex > StartIndex.
		for i, loc := range p.Locations {
			if loc.EndIndex <= loc.StartIndex {
				t.Errorf("%s loc[%d]: EndIndex (%d) must be > StartIndex (%d)",
					p.PackageName, i, loc.EndIndex, loc.StartIndex)
			}
		}
	}

	// Expected versions from Package.resolved (lock file), not manifest ranges.
	cases := map[string]string{
		"swift-nio":              "2.42.0", // manifest: 2.0.0 (from:), lock: 2.42.0
		"swift-log":              "1.4.4",  // manifest: 1.4.0 (.upToNextMajor), lock: 1.4.4
		"swift-crypto":           "2.0.6",  // manifest: 2.0.0 (.upToNextMinor), lock: 2.0.6
		"Alamofire":              "5.2.0",  // exact: matches lock
		"SnapKit":                "5.6.0",  // manifest: 5.0.0 (.upToNextMajor), lock: 5.6.0
		"Kingfisher":             "7.6.2",  // manifest: 7.0.0 (from:), lock: 7.6.2
		"swift-collections":      "latest", // branch only, no version in lock
		"swift-syntax":           "latest", // revision only, no version in lock
		"apple.swift-algorithms": "1.0.0",  // exact matches lock
	}
	for name, wantVer := range cases {
		if ver, ok := got[name]; !ok {
			t.Errorf("expected package %q not found", name)
		} else if ver != wantVer {
			t.Errorf("%s: version = %q, want %q", name, ver, wantVer)
		}
	}

	// Local-path .package(path: "../LocalDep") must NOT appear.
	if _, exists := got["LocalDep"]; exists {
		t.Errorf("local .package(path:) should be skipped, but LocalDep was emitted")
	}
}

func TestPackageNameFromURL(t *testing.T) {
	cases := map[string]string{
		"https://github.com/apple/swift-nio.git":     "swift-nio",
		"https://github.com/apple/swift-nio":         "swift-nio",
		"git@github.com:apple/swift-log.git":         "swift-log",
		"https://gitlab.com/group/sub/MyPackage.git": "MyPackage",
		"https://example.com/repo/":                  "repo",
	}
	for in, want := range cases {
		if got := packageNameFromURL(in); got != want {
			t.Errorf("packageNameFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStripLineComment(t *testing.T) {
	cases := []struct{ in, out string }{
		{`.package(url: "https://example.com", from: "1.0.0") // pin`,
			`.package(url: "https://example.com", from: "1.0.0") `},
		{`.package(url: "https://x.com/foo//bar.git")`,
			`.package(url: "https://x.com/foo//bar.git")`},
		{`// whole line`, ``},
	}
	for _, c := range cases {
		if got := stripLineComment(c.in); got != c.out {
			t.Errorf("stripLineComment(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

func TestParenDelta(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{`.package(url: "x", from: "1")`, 0},
		{`.package(`, 1},
		{`)`, -1},
		{`.package(url: "(not a paren)", from: "1")`, 0},
	}
	for _, c := range cases {
		if got := parenDelta(c.in); got != c.want {
			t.Errorf("parenDelta(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestPackageSwift_WithLockFile(t *testing.T) {
	// TC_32: Verify that when both Package.swift and Package.resolved are present,
	// exact versions from the lock file are used instead of ranges from the manifest.
	parser := &SwiftPmParser{}
	pkgs, err := parser.Parse(filepath.Join("..", "..", "..", "test", "resources", "Package.swift"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	got := make(map[string]string, len(pkgs))
	for _, p := range pkgs {
		got[p.PackageName] = p.Version
	}

	// Expected versions from Package.resolved (lock file), not from Package.swift ranges.
	lockExpectations := map[string]string{
		"swift-nio":              "2.42.0", // Lock has 2.42.0, manifest has 2.0.0 (from:)
		"swift-log":              "1.4.4",  // Lock has 1.4.4, manifest has 1.4.0 (.upToNextMajor)
		"swift-crypto":           "2.0.6",  // Lock has 2.0.6, manifest has 2.0.0 (.upToNextMinor)
		"Alamofire":              "5.2.0",  // Lock matches manifest (exact:)
		"SnapKit":                "5.6.0",  // Lock has 5.6.0, manifest has 5.0.0 (.upToNextMajor)
		"Kingfisher":             "7.6.2",  // Lock has 7.6.2, manifest has 7.0.0 (from:)
		"swift-collections":      "latest", // Lock has no version (branch only)
		"swift-syntax":           "latest", // Lock has no version (revision only)
		"apple.swift-algorithms": "1.0.0",  // Lock matches manifest
	}

	for name, wantVer := range lockExpectations {
		if ver, ok := got[name]; !ok {
			t.Errorf("expected package %q not found", name)
		} else if ver != wantVer {
			t.Errorf("%s: version = %q, want %q (lock file should override manifest range)",
				name, ver, wantVer)
		}
	}
}
