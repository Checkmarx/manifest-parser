package setuptools

import (
	"os"
	"path/filepath"
	"testing"
)

// Edge case tests for robustness and error handling

func TestSetupCfgParser_EmptyFile(t *testing.T) {
	content := ""
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages from empty file, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_OnlyComments(t *testing.T) {
	content := "# Comment 1\n# Comment 2\n# Comment 3\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages from comment-only file, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_MalformedSection(t *testing.T) {
	// Section without closing bracket should be ignored
	content := "[options\ninstall_requires =\n    flask==2.0.1\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still parse the dependency
	if len(pkgs) == 0 {
		t.Fatalf("expected packages from partially malformed file, got 0")
	}
}

func TestSetupCfgParser_MixedIndentation(t *testing.T) {
	// Mix of tabs and spaces - should still work
	content := "[options]\ninstall_requires =\n\trequests>=2.0\n    flask==2.0.1\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages with mixed indentation, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_PackageNameWithNumbers(t *testing.T) {
	// Package names can start with numbers
	content := "[options]\ninstall_requires =\n    py2exe\n    3to2\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages with numeric names, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_DuplicateSections(t *testing.T) {
	// Same section defined twice - second one should override
	content := "[options]\ninstall_requires =\n    flask==1.0.0\n[options]\ninstall_requires =\n    flask==2.0.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should find both (no dedup in parser)
	if len(pkgs) < 1 {
		t.Fatalf("expected at least 1 package with duplicate sections, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_VeryLongLine(t *testing.T) {
	// Create a very long line with many dependencies
	longDeps := "[options]\ninstall_requires =\n"
	for i := 0; i < 100; i++ {
		longDeps += "    package" + string(rune(48+i%10)) + "\n"
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(longDeps), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 100 {
		t.Fatalf("expected 100 packages from long dependency list, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_UnicodePackageName(t *testing.T) {
	// Package names with unicode (should be skipped as invalid)
	content := "[options]\ninstall_requires =\n    café-package\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Unicode shouldn't match the regex, so should be skipped
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages with unicode name, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_VersionSpecifierEdgeCases(t *testing.T) {
	content := "[options]\ninstall_requires =\n    package1==1.0.0\n    package2!=1.0.0\n    package3~=1.0\n    package4>1.0\n    package5<2.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 5 {
		t.Fatalf("expected 5 packages with various specifiers, got %d", len(pkgs))
	}

	versions := make(map[string]string)
	for _, pkg := range pkgs {
		versions[pkg.PackageName] = pkg.Version
	}

	if versions["package1"] != "1.0.0" {
		t.Errorf("package1: expected exact version, got %s", versions["package1"])
	}
	if versions["package2"] != "latest" {
		t.Errorf("package2 (!=): expected latest, got %s", versions["package2"])
	}
	if versions["package3"] != "latest" {
		t.Errorf("package3 (~=): expected latest, got %s", versions["package3"])
	}
	if versions["package4"] != "latest" {
		t.Errorf("package4 (>): expected latest, got %s", versions["package4"])
	}
	if versions["package5"] != "latest" {
		t.Errorf("package5 (<): expected latest, got %s", versions["package5"])
	}
}
