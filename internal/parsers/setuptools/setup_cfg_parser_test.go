package setuptools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestSetupCfgParser_ParseExactVersion(t *testing.T) {
	content := "[options]\ninstall_requires =\n    flask==2.0.1\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "flask",
		Version:        "2.0.1",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       2,
			StartIndex: 4,
			EndIndex:   16,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestSetupCfgParser_ParseRangeVersion(t *testing.T) {
	content := "[options]\ninstall_requires =\n    requests>=2.26.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "requests",
		Version:        "latest",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       2,
			StartIndex: 4,
			EndIndex:   20,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestSetupCfgParser_ParseMultipleDependencies(t *testing.T) {
	content := "[options]\ninstall_requires =\n    requests>=2.26.0\n    flask==2.0.1\n    six\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}

	expected := []models.Package{
		{
			PackageManager: "pypi",
			PackageName:    "requests",
			Version:        "latest",
			FilePath:       filePath,
			Locations: []models.Location{{Line: 2, StartIndex: 4, EndIndex: 20}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "flask",
			Version:        "2.0.1",
			FilePath:       filePath,
			Locations: []models.Location{{Line: 3, StartIndex: 4, EndIndex: 16}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "six",
			Version:        "latest",
			FilePath:       filePath,
			Locations: []models.Location{{Line: 4, StartIndex: 4, EndIndex: 7}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestSetupCfgParser_ParseSetupRequires(t *testing.T) {
	content := "[options]\nsetup_requires =\n    setuptools>=42\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "setuptools",
		Version:        "latest",
		FilePath:       filePath,
		Locations: []models.Location{{Line: 2, StartIndex: 4, EndIndex: 18}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestSetupCfgParser_ParseExtrasRequire(t *testing.T) {
	content := "[options.extras_require]\ndev =\n    pytest>=6.0\n    black==22.3.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	expected := []models.Package{
		{
			PackageManager: "pypi",
			PackageName:    "pytest",
			Version:        "latest",
			FilePath:       filePath,
			Locations: []models.Location{{Line: 2, StartIndex: 4, EndIndex: 15}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "black",
			Version:        "22.3.0",
			FilePath:       filePath,
			Locations: []models.Location{{Line: 3, StartIndex: 4, EndIndex: 17}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestSetupCfgParser_ParseSkipCommentLine(t *testing.T) {
	content := "[options]\ninstall_requires =\n    # commented out\n    flask==2.0.1\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "flask",
		Version:        "2.0.1",
		FilePath:       filePath,
		Locations: []models.Location{{Line: 3, StartIndex: 4, EndIndex: 16}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestSetupCfgParser_ParseInlineComment(t *testing.T) {
	content := "[options]\ninstall_requires =\n    flask==2.0.1  # web framework\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "flask",
		Version:        "2.0.1",
		FilePath:       filePath,
		Locations: []models.Location{{Line: 2, StartIndex: 4, EndIndex: 16}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestSetupCfgParser_ParseNoDependencies(t *testing.T) {
	content := "[metadata]\nname = my-package\nversion = 1.0.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.cfg")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestSetupCfgParser_Parse_RealFile(t *testing.T) {
	filePath := "../../testdata/setup.cfg"
	parser := &SetupCfgParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []models.Package{
		{PackageManager: "pypi", PackageName: "requests", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 6, StartIndex: 4, EndIndex: 20}}},
		{PackageManager: "pypi", PackageName: "flask", Version: "2.0.1", FilePath: filePath, Locations: []models.Location{{Line: 7, StartIndex: 4, EndIndex: 16}}},
		{PackageManager: "pypi", PackageName: "six", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 8, StartIndex: 4, EndIndex: 7}}},
		{PackageManager: "pypi", PackageName: "Pillow", Version: "9.0.0", FilePath: filePath, Locations: []models.Location{{Line: 9, StartIndex: 4, EndIndex: 17}}},
		{PackageManager: "pypi", PackageName: "cryptography", Version: "2.9.2", FilePath: filePath, Locations: []models.Location{{Line: 10, StartIndex: 4, EndIndex: 23}}},
		{PackageManager: "pypi", PackageName: "setuptools", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 13, StartIndex: 4, EndIndex: 18}}},
		{PackageManager: "pypi", PackageName: "pytest", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 17, StartIndex: 4, EndIndex: 15}}},
		{PackageManager: "pypi", PackageName: "black", Version: "22.3.0", FilePath: filePath, Locations: []models.Location{{Line: 18, StartIndex: 4, EndIndex: 17}}},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}
