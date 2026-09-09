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

// --- uv support: [dependency-groups] (PEP 735) ---

func TestParsePyprojectDependencyGroupsBasic(t *testing.T) {
	content := "[dependency-groups]\ndev = [\n    \"pytest>=7.0\",\n    \"black==22.0\",\n]\n"
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
	if pkgs[0].PackageName != "pytest" || pkgs[0].Version != "latest" {
		t.Errorf("expected pytest@latest (ranged), got %s@%s", pkgs[0].PackageName, pkgs[0].Version)
	}
	if pkgs[1].PackageName != "black" || pkgs[1].Version != "22.0" {
		t.Errorf("expected black@22.0 (exact), got %s@%s", pkgs[1].PackageName, pkgs[1].Version)
	}
	if pkgs[0].PackageManager != "pypi" {
		t.Errorf("expected PackageManager %q, got %q", "pypi", pkgs[0].PackageManager)
	}
}

func TestParsePyprojectDependencyGroupsInlineArray(t *testing.T) {
	content := "[dependency-groups]\ntest = [\"pytest\", \"coverage\"]\n"
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
	if pkgs[0].PackageName != "pytest" || pkgs[1].PackageName != "coverage" {
		t.Errorf("expected [pytest, coverage], got [%s, %s]", pkgs[0].PackageName, pkgs[1].PackageName)
	}
}

func TestParsePyprojectDependencyGroupsIncludeGroupSkipped(t *testing.T) {
	// PEP 735 lets a group reference another via {include-group = "..."}. That entry
	// is not a real package and must be skipped without producing a garbage name;
	// the referenced group's own packages are still captured from its own array.
	content := "[dependency-groups]\n" +
		"test = [\"pytest>=7.0\"]\n" +
		"dev = [\n" +
		"    {include-group = \"test\"},\n" +
		"    \"black==22.0\",\n" +
		"]\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages (pytest from test, black from dev; include-group skipped), got %d", len(pkgs))
	}
	names := []string{pkgs[0].PackageName, pkgs[1].PackageName}
	if names[0] != "pytest" || names[1] != "black" {
		t.Errorf("expected [pytest, black], got %v", names)
	}
}

func TestParsePyprojectDependencyGroupsIncludeGroupSingleLine(t *testing.T) {
	content := "[dependency-groups]\ndev = [\"black==22.0\", {include-group = \"test\"}]\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package (include-group skipped), got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "black" {
		t.Errorf("expected %q, got %q", "black", pkgs[0].PackageName)
	}
}

// --- uv support: uv.lock companion resolution ---

func TestParseUvLockResolvesRangedVersion(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"requests>=2.0\",\n]\n"
	uvLock := "version = 1\nrevision = 2\n\n[[package]]\nname = \"requests\"\nversion = \"2.32.4\"\n" +
		"source = { registry = \"https://pypi.org/simple\" }\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "2.32.4" {
		t.Errorf("expected version resolved from uv.lock %q, got %q", "2.32.4", pkgs[0].Version)
	}
}

func TestParseUvLockResolvesWildcardAndCompatibleRelease(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"numpy==1.*\",\n    \"pandas~=2.0\",\n]\n"
	uvLock := "[[package]]\nname = \"numpy\"\nversion = \"1.26.4\"\nsource = { registry = \"https://pypi.org/simple\" }\n\n" +
		"[[package]]\nname = \"pandas\"\nversion = \"2.2.1\"\nsource = { registry = \"https://pypi.org/simple\" }\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Version != "1.26.4" {
		t.Errorf("expected numpy resolved to %q, got %q", "1.26.4", pkgs[0].Version)
	}
	if pkgs[1].Version != "2.2.1" {
		t.Errorf("expected pandas resolved to %q, got %q", "2.2.1", pkgs[1].Version)
	}
}

func TestParseUvLockMissingEntryFallsBackToLatest(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"unpinned-pkg>=1.0\",\n]\n"
	uvLock := "[[package]]\nname = \"other-pkg\"\nversion = \"1.0.0\"\nsource = { registry = \"https://pypi.org/simple\" }\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected %q, got %q", "latest", pkgs[0].Version)
	}
}

func TestParseUvLockAbsentFileDoesNotError(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"requests>=2.0\",\n]\n"
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 || pkgs[0].Version != "latest" {
		t.Fatalf("expected 1 package at %q, got %d packages, version %q", "latest", len(pkgs), pkgs[0].Version)
	}
}

func TestParseUvLockNameNormalization(t *testing.T) {
	// pyproject.toml may declare a dependency with underscores/mixed case; PEP 503
	// normalization must still match it against uv.lock's hyphenated lowercase name.
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"Flask_SQLAlchemy>=3.0\",\n]\n"
	uvLock := "[[package]]\nname = \"flask-sqlalchemy\"\nversion = \"3.1.1\"\nsource = { registry = \"https://pypi.org/simple\" }\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "3.1.1" {
		t.Errorf("expected PEP 503 normalized lookup to resolve %q, got %q", "3.1.1", pkgs[0].Version)
	}
}

func TestParsePoetryLockTakesPrecedenceOverUvLock(t *testing.T) {
	// A project migrating between tools could transiently have both lock files.
	// poetry.lock wins per-key; uv.lock fills in anything poetry.lock doesn't cover.
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"requests>=2.0\",\n    \"flask>=2.0\",\n]\n"
	poetryLock := "[[package]]\nname = \"requests\"\nversion = \"2.31.0\"\n"
	uvLock := "[[package]]\nname = \"requests\"\nversion = \"2.32.4\"\n\n" +
		"[[package]]\nname = \"flask\"\nversion = \"3.0.0\"\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "poetry.lock"), []byte(poetryLock), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].Version != "2.31.0" {
		t.Errorf("expected poetry.lock version %q to take precedence, got %q", "2.31.0", pkgs[0].Version)
	}
	if pkgs[1].Version != "3.0.0" {
		t.Errorf("expected uv.lock to fill in flask@%q, got %q", "3.0.0", pkgs[1].Version)
	}
}

func TestParsePep621ExtrasStrippedFromName(t *testing.T) {
	// "requests[security]" must report PackageName "requests" — otherwise it
	// matches neither a real PyPI package nor the lock file's unbracketed name.
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"requests[security]>=2.0\",\n    \"uvicorn[standard]\",\n]\n"
	uvLock := "[[package]]\nname = \"requests\"\nversion = \"2.32.4\"\n\n[[package]]\nname = \"uvicorn\"\nversion = \"0.30.1\"\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "requests" || pkgs[0].Version != "2.32.4" {
		t.Errorf("requests[security]: expected name=%q version=%q, got name=%q version=%q",
			"requests", "2.32.4", pkgs[0].PackageName, pkgs[0].Version)
	}
	if pkgs[1].PackageName != "uvicorn" || pkgs[1].Version != "0.30.1" {
		t.Errorf("uvicorn[standard] (no specifier): expected name=%q version=%q, got name=%q version=%q",
			"uvicorn", "0.30.1", pkgs[1].PackageName, pkgs[1].Version)
	}
}

func TestParsePep621BareNameNoLockFallsBackToLatest(t *testing.T) {
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"somepkg\",\n]\n"
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 || pkgs[0].Version != "latest" {
		t.Fatalf("expected 1 package at %q, got %d packages, version %q", "latest", len(pkgs), pkgs[0].Version)
	}
}

func TestParsePep621TrailingCommentInMultilineArray(t *testing.T) {
	// A trailing "# comment" on a multi-line array item must not leak into the
	// parsed version (previously produced Version == `7.0",  # pinned...`).
	tmpDir := t.TempDir()
	pyproject := "[dependency-groups]\ndev = [\n    \"pytest==7.0\",  # pinned for CI stability\n]\n"
	tmpFile := filepath.Join(tmpDir, "pyproject.toml")
	os.WriteFile(tmpFile, []byte(pyproject), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "7.0" {
		t.Errorf("expected clean version %q, got %q", "7.0", pkgs[0].Version)
	}
}

func TestParsePep621ArbitraryEquality(t *testing.T) {
	// PEP 440 "===" (arbitrary equality) is a distinct operator from "==" —
	// it must not be truncated to a malformed version like "=1.2.3.4".
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"legacypkg===1.2.3.4\",\n]\n"
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "legacypkg" || pkgs[0].Version != "1.2.3.4" {
		t.Errorf("expected legacypkg@1.2.3.4, got %s@%s", pkgs[0].PackageName, pkgs[0].Version)
	}
}

func TestParsePep621MarkerOnlyNoVersionSpecifier(t *testing.T) {
	// A marker's own comparison operators (e.g. "<" in "python_version < '3.11'")
	// must not be mistaken for the package's version separator — a very common
	// real-world pattern for conditional backport dependencies.
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"tomli; python_version < '3.11'\",\n    \"typing-extensions; python_version >= '3.8'\",\n]\n"
	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "tomli" {
		t.Errorf("expected %q, got %q", "tomli", pkgs[0].PackageName)
	}
	if pkgs[1].PackageName != "typing-extensions" {
		t.Errorf("expected %q, got %q", "typing-extensions", pkgs[1].PackageName)
	}
}

func TestParseUvLockIgnoresNonPackageTables(t *testing.T) {
	// [package.metadata] (root project's requires-dist) and other non-[[package]]
	// tables must not be mistaken for a package block or corrupt the scan.
	tmpDir := t.TempDir()
	pyproject := "[project]\ndependencies = [\n    \"termcolor>=3.0\",\n]\n"
	uvLock := "[[package]]\n" +
		"name = \"my-project\"\n" +
		"version = \"0.1.0\"\n" +
		"source = { virtual = \".\" }\n" +
		"dependencies = [\n    { name = \"termcolor\" },\n]\n\n" +
		"[package.metadata]\n" +
		"requires-dist = [{ name = \"termcolor\", specifier = \">=3.0\" }]\n\n" +
		"[[package]]\n" +
		"name = \"termcolor\"\n" +
		"version = \"3.3.0\"\n" +
		"source = { registry = \"https://pypi.org/simple\" }\n"

	os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644)
	os.WriteFile(filepath.Join(tmpDir, "uv.lock"), []byte(uvLock), 0644)

	parser := &PoetryPyprojectParser{}
	pkgs, err := parser.Parse(filepath.Join(tmpDir, "pyproject.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "3.3.0" {
		t.Errorf("expected termcolor resolved to %q, got %q", "3.3.0", pkgs[0].Version)
	}
}
