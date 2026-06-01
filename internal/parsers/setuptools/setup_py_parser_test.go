package setuptools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestSetupPyParser_ParseSingleInstallRequire(t *testing.T) {
	content := "from setuptools import setup\n\nsetup(\n    install_requires=['flask==2.0.1'],\n)\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	if got.PackageManager != "pypi" || got.PackageName != "flask" || got.Version != "2.0.1" {
		t.Errorf("got package %s %s, want flask 2.0.1", got.PackageName, got.Version)
	}
}

func TestSetupPyParser_ParseRangeInstallRequire(t *testing.T) {
	content := "from setuptools import setup\n\nsetup(\n    install_requires=['requests>=2.26.0'],\n)\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	if got.PackageManager != "pypi" || got.PackageName != "requests" || got.Version != "latest" {
		t.Errorf("got package %s %s, want requests latest", got.PackageName, got.Version)
	}
}

func TestSetupPyParser_ParseMultipleDependencies(t *testing.T) {
	content := `from setuptools import setup

setup(
    install_requires=[
        'requests>=2.26.0',
        'flask==2.0.1',
        'six',
    ],
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}

	packageNames := make(map[string]string)
	for _, pkg := range pkgs {
		packageNames[pkg.PackageName] = pkg.Version
	}

	if packageNames["requests"] != "latest" {
		t.Errorf("requests: expected latest, got %s", packageNames["requests"])
	}
	if packageNames["flask"] != "2.0.1" {
		t.Errorf("flask: expected 2.0.1, got %s", packageNames["flask"])
	}
	if packageNames["six"] != "latest" {
		t.Errorf("six: expected latest, got %s", packageNames["six"])
	}
}

func TestSetupPyParser_ParseExtrasRequire(t *testing.T) {
	content := `from setuptools import setup

setup(
    install_requires=['requests>=2.26.0'],
    extras_require={
        'dev': ['pytest>=6.0', 'black==22.3.0'],
    },
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) < 3 {
		t.Fatalf("expected at least 3 packages, got %d", len(pkgs))
	}

	packageNames := make(map[string]string)
	for _, pkg := range pkgs {
		packageNames[pkg.PackageName] = pkg.Version
	}

	if packageNames["requests"] != "latest" {
		t.Errorf("requests: expected latest, got %s", packageNames["requests"])
	}
	if packageNames["pytest"] != "latest" {
		t.Errorf("pytest: expected latest, got %s", packageNames["pytest"])
	}
	if packageNames["black"] != "22.3.0" {
		t.Errorf("black: expected 22.3.0, got %s", packageNames["black"])
	}
}

func TestSetupPyParser_ParseTestsRequire(t *testing.T) {
	content := "from setuptools import setup\n\nsetup(\n    tests_require=['pytest'],\n)\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	got := pkgs[0]
	if got.PackageManager != "pypi" || got.PackageName != "pytest" {
		t.Errorf("got package %s, want pytest", got.PackageName)
	}
}

func TestSetupPyParser_ParseNoRequires(t *testing.T) {
	content := "from setuptools import setup\n\nsetup(\n    name='my-package',\n)\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestSetupPyParser_ParseWithDoubleQuotes(t *testing.T) {
	content := `from setuptools import setup

setup(
    install_requires=[
        "requests>=2.26.0",
        "flask==2.0.1",
    ],
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "setup.py")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	packageNames := make(map[string]string)
	for _, pkg := range pkgs {
		packageNames[pkg.PackageName] = pkg.Version
	}

	if packageNames["requests"] != "latest" {
		t.Errorf("requests: expected latest, got %s", packageNames["requests"])
	}
	if packageNames["flask"] != "2.0.1" {
		t.Errorf("flask: expected 2.0.1, got %s", packageNames["flask"])
	}
}

func TestSetupPyParser_Parse_RealFile(t *testing.T) {
	filePath := "../../testdata/setup.py"
	parser := &SetupPyParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []models.Package{
		{PackageManager: "pypi", PackageName: "requests", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 6, StartIndex: 9, EndIndex: 25}}},
		{PackageManager: "pypi", PackageName: "flask", Version: "2.0.1", FilePath: filePath, Locations: []models.Location{{Line: 7, StartIndex: 9, EndIndex: 21}}},
		{PackageManager: "pypi", PackageName: "six", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 8, StartIndex: 9, EndIndex: 12}}},
		{PackageManager: "pypi", PackageName: "Pillow", Version: "9.0.0", FilePath: filePath, Locations: []models.Location{{Line: 9, StartIndex: 9, EndIndex: 22}}},
		{PackageManager: "pypi", PackageName: "cryptography", Version: "2.9.2", FilePath: filePath, Locations: []models.Location{{Line: 10, StartIndex: 9, EndIndex: 28}}},
		{PackageManager: "pypi", PackageName: "pytest", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 19, StartIndex: 9, EndIndex: 15}}},
		{PackageManager: "pypi", PackageName: "pytest", Version: "latest", FilePath: filePath, Locations: []models.Location{{Line: 14, StartIndex: 13, EndIndex: 24}}},
		{PackageManager: "pypi", PackageName: "black", Version: "22.3.0", FilePath: filePath, Locations: []models.Location{{Line: 15, StartIndex: 13, EndIndex: 26}}},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}
