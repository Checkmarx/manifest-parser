package poetry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestParsePyprojectExactVersion(t *testing.T) {
	content := "[tool.poetry.dependencies]\nrequests = \"2.28.2\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "requests",
		Version:        "2.28.2",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       1,
			StartIndex: 0,
			EndIndex:   19,
		}},
	}
	testdata.ValidatePackages(t, pkgs, []models.Package{want})
}

func TestParsePyprojectRangedVersion(t *testing.T) {
	content := "[tool.poetry.dependencies]\nflask = \"^2.3.0\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version %q, got %q", "latest", pkgs[0].Version)
	}
}

func TestParsePyprojectSkipsPython(t *testing.T) {
	content := "[tool.poetry.dependencies]\npython = \"^3.9\"\nrequests = \"2.28.2\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package (python skipped), got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "requests" {
		t.Errorf("expected package %q, got %q", "requests", pkgs[0].PackageName)
	}
}

func TestParsePyprojectDevDependencies(t *testing.T) {
	content := "[tool.poetry.dev-dependencies]\npytest = \"7.2.0\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	want := models.Package{
		PackageManager: "pypi",
		PackageName:    "pytest",
		Version:        "7.2.0",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       1,
			StartIndex: 0,
			EndIndex:   16,
		}},
	}
	testdata.ValidatePackages(t, pkgs, []models.Package{want})
}

func TestParsePyprojectGroupDependencies(t *testing.T) {
	content := "[tool.poetry.group.lint.dependencies]\nblack = \"^22.0.0\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "black" {
		t.Errorf("expected package %q, got %q", "black", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version %q, got %q", "latest", pkgs[0].Version)
	}
}

func TestParsePyprojectGroupExactVersion(t *testing.T) {
	content := "[tool.poetry.group.test.dependencies]\npytest = \"7.4.0\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "7.4.0" {
		t.Errorf("expected version %q, got %q", "7.4.0", pkgs[0].Version)
	}
}

func TestParsePyprojectInlineTable(t *testing.T) {
	content := "[tool.poetry.dependencies]\nnumpy = {version = \"1.24.3\", optional = true}\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "numpy" {
		t.Errorf("expected package %q, got %q", "numpy", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "1.24.3" {
		t.Errorf("expected version %q, got %q", "1.24.3", pkgs[0].Version)
	}
}

func TestParsePyprojectWildcardVersion(t *testing.T) {
	content := "[tool.poetry.dependencies]\nrequests = \"*\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version %q, got %q", "latest", pkgs[0].Version)
	}
}

func TestParsePyprojectPartialWildcard(t *testing.T) {
	content := "[tool.poetry.dependencies]\nrequests = \"2.28.*\"\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version %q, got %q", "latest", pkgs[0].Version)
	}
}

func TestParsePyprojectInlineComment(t *testing.T) {
	content := "[tool.poetry.dependencies]\nrequests = \"2.28.2\"  # pinned for security\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "2.28.2" {
		t.Errorf("expected version %q, got %q", "2.28.2", pkgs[0].Version)
	}
}

func TestParsePyprojectNoDepSection(t *testing.T) {
	content := "[build-system]\nrequires = [\"poetry-core>=1.0.0\"]\n[tool.black]\nline-length = 88\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParsePyprojectPep621Dependencies(t *testing.T) {
	content := "[project]\ndependencies = [\n    \"requests>=2.28.0\",\n    \"flask==2.3.0\",\n]\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "requests" {
		t.Errorf("expected %q, got %q", "requests", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version %q (ranged), got %q", "latest", pkgs[0].Version)
	}
	if pkgs[1].PackageName != "flask" {
		t.Errorf("expected %q, got %q", "flask", pkgs[1].PackageName)
	}
	if pkgs[1].Version != "2.3.0" {
		t.Errorf("expected version %q, got %q", "2.3.0", pkgs[1].Version)
	}
}

func TestParsePyprojectPep621OptionalDeps(t *testing.T) {
	content := "[project.optional-dependencies]\ndev = [\n    \"pytest>=7.0\",\n    \"black>=22.0\",\n]\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "pytest" {
		t.Errorf("expected %q, got %q", "pytest", pkgs[0].PackageName)
	}
	if pkgs[1].PackageName != "black" {
		t.Errorf("expected %q, got %q", "black", pkgs[1].PackageName)
	}
}

func TestParsePyprojectRealFile(t *testing.T) {
	filePath := "../../testdata/pyproject.toml"
	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pkgs) != 7 {
		t.Fatalf("expected 7 packages, got %d", len(pkgs))
	}

	packageNames := make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		packageNames = append(packageNames, pkg.PackageName)
	}

	expectedNames := []string{
		"requests", "flask", "Pillow", "cryptography", "pytest", "numpy", "pandas",
	}

	for i, expectedName := range expectedNames {
		if i < len(packageNames) {
			if packageNames[i] != expectedName {
				t.Errorf("package %d: expected %q, got %q", i, expectedName, packageNames[i])
			}
		}
	}
}
