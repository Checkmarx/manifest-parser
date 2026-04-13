package pypi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"

	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestParseExactVersion(t *testing.T) {
	content := "flask==1.1.2\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
		Version:        "1.1.2",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       0,
			StartIndex: 0,
			EndIndex:   12,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestParseInlineComment(t *testing.T) {
	content := "   requests==2.25.1  # pinned for compatibility\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
		Version:        "2.25.1",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       0,
			StartIndex: 3,
			EndIndex:   19,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestParseRequirementLineEndIndex(t *testing.T) {
	content := "requests==2.25.1  # pinned for compatibility\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
		Version:        "2.25.1",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       0,
			StartIndex: 0,
			EndIndex:   16,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestParseSkipCommentLine(t *testing.T) {
	content := "# just a comment\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestPypiParser_Parse_RealFile(t *testing.T) {
	filePath := "../../testdata/requirements.txt"
	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []models.Package{
		{
			PackageManager: "pypi",
			PackageName:    "ansicolors",
			Version:        "1.1.8",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       3,
				StartIndex: 0,
				EndIndex:   17,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "setuptools",
			Version:        "latest",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       4,
				StartIndex: 1,
				EndIndex:   23,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "types-setuptools",
			Version:        "latest",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       5,
				StartIndex: 2,
				EndIndex:   30,
			}},
		},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseLineContinuationWithHashes(t *testing.T) {
	content := "asgiref==3.8.1 \\\n    --hash=sha256:89b2ef22 \\\n    --hash=sha256:9e0ce3aa\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
		PackageName:    "asgiref",
		Version:        "3.8.1",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       0,
			StartIndex: 0,
			EndIndex:   14,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestParsePipOptionLinesSkipped(t *testing.T) {
	content := "--index-url https://pypi.org/simple\n-r base-requirements.txt\nflask==3.1.0\n-e git+https://github.com/foo/bar.git#egg=bar\nrequests==2.32.3\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
			PackageName:    "flask",
			Version:        "3.1.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       2,
				StartIndex: 0,
				EndIndex:   12,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "requests",
			Version:        "2.32.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       4,
				StartIndex: 0,
				EndIndex:   16,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseEnvMarkerWithContinuation(t *testing.T) {
	content := "tzdata==2025.3 ; sys_platform == 'win32' \\\n    --hash=sha256:06a47e57 \\\n    --hash=sha256:de39c2ca\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
		PackageName:    "tzdata",
		Version:        "2025.3",
		FilePath:       filePath,
		Locations: []models.Location{{
			Line:       0,
			StartIndex: 0,
			EndIndex:   14,
		}},
	}
	testdata.ValidatePackages(t, []models.Package{got}, []models.Package{want})
}

func TestParseViaCommentsIgnored(t *testing.T) {
	content := "asgiref==3.8.1\n    # via django\ndjango==5.1.7\n    # via sample-app\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
			PackageName:    "asgiref",
			Version:        "3.8.1",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "django",
			Version:        "5.1.7",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       2,
				StartIndex: 0,
				EndIndex:   13,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseLineContinuationLocationTracking(t *testing.T) {
	content := "# comment\nasgiref==3.8.1 \\\n    --hash=sha256:abc123 \\\n    --hash=sha256:def456\ndjango==5.1.7\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	// asgiref starts at physical line 1 (0-indexed), after the comment on line 0
	if pkgs[0].Locations[0].Line != 1 {
		t.Errorf("asgiref: expected line 1, got %d", pkgs[0].Locations[0].Line)
	}
	// django is at physical line 4
	if pkgs[1].Locations[0].Line != 4 {
		t.Errorf("django: expected line 4, got %d", pkgs[1].Locations[0].Line)
	}
}

func TestPypiParser_Parse_UvExportFile(t *testing.T) {
	filePath := "../../testdata/requirements-uv-export.txt"
	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []models.Package{
		{
			PackageManager: "pypi",
			PackageName:    "asgiref",
			Version:        "3.8.1",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       2,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "django",
			Version:        "5.1.7",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       8,
				StartIndex: 0,
				EndIndex:   13,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "pycryptodome",
			Version:        "3.21.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       12,
				StartIndex: 0,
				EndIndex:   20,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "sqlparse",
			Version:        "0.5.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       17,
				StartIndex: 0,
				EndIndex:   15,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "typing-extensions",
			Version:        "4.12.2",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       23,
				StartIndex: 0,
				EndIndex:   25,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "tzdata",
			Version:        "2025.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       29,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}

func TestPypiParser_Parse_PipFreezeFile(t *testing.T) {
	filePath := "../../testdata/requirements-pip-freeze.txt"
	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []models.Package{
		{
			PackageManager: "pypi",
			PackageName:    "asgiref",
			Version:        "3.8.1",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "Django",
			Version:        "5.1.7",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       1,
				StartIndex: 0,
				EndIndex:   13,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "sqlparse",
			Version:        "0.5.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       2,
				StartIndex: 0,
				EndIndex:   15,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "tzdata",
			Version:        "2025.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       3,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}

func TestPypiParser_Parse_PipCompileFile(t *testing.T) {
	filePath := "../../testdata/requirements-pip-compile.txt"
	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []models.Package{
		{
			PackageManager: "pypi",
			PackageName:    "asgiref",
			Version:        "3.8.1",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       6,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "django",
			Version:        "5.1.7",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       8,
				StartIndex: 0,
				EndIndex:   13,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "sqlparse",
			Version:        "0.5.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       10,
				StartIndex: 0,
				EndIndex:   15,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "tzdata",
			Version:        "2025.3",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       12,
				StartIndex: 0,
				EndIndex:   14,
			}},
		},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseArbitraryEquality(t *testing.T) {
	content := "mypackage===1.0.dev1\nflask==3.1.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
			PackageName:    "mypackage",
			Version:        "1.0.dev1",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 0,
				EndIndex:   20,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "flask",
			Version:        "3.1.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       1,
				StartIndex: 0,
				EndIndex:   12,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseURLRequirement(t *testing.T) {
	content := "requests @ https://example.com/requests-2.32.3.tar.gz\nflask==3.1.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
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
			PackageName:    "requests",
			Version:        "latest",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 0,
				EndIndex:   53,
			}},
		},
		{
			PackageManager: "pypi",
			PackageName:    "flask",
			Version:        "3.1.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       1,
				StartIndex: 0,
				EndIndex:   12,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseURLRequirementWithExtras(t *testing.T) {
	content := "requests[security] @ https://example.com/requests-2.32.3.tar.gz\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "requests" {
		t.Errorf("expected package name 'requests', got %q", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version 'latest', got %q", pkgs[0].Version)
	}
}

func TestParseVCSRequirement(t *testing.T) {
	content := "git+https://github.com/user/repo.git@v1.0#egg=mypackage\nflask==3.1.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "mypackage" {
		t.Errorf("expected package name 'mypackage', got %q", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "latest" {
		t.Errorf("expected version 'latest', got %q", pkgs[0].Version)
	}
	if pkgs[1].PackageName != "flask" {
		t.Errorf("expected package name 'flask', got %q", pkgs[1].PackageName)
	}
	if pkgs[1].Version != "3.1.0" {
		t.Errorf("expected version '3.1.0', got %q", pkgs[1].Version)
	}
}

func TestParseVCSRequirementNoEgg(t *testing.T) {
	content := "git+https://github.com/user/repo.git@v1.0\nflask==3.1.0\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// VCS line without #egg= should be skipped
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package (VCS without egg skipped), got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "flask" {
		t.Errorf("expected package name 'flask', got %q", pkgs[0].PackageName)
	}
}

func TestParseVCSSchemes(t *testing.T) {
	content := "git+https://github.com/user/repo1.git#egg=pkg1\nhg+https://hg.example.com/repo2#egg=pkg2\nsvn+svn://svn.example.com/repo3#egg=pkg3\nbzr+lp:repo4#egg=pkg4\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(pkgs))
	}

	expectedNames := []string{"pkg1", "pkg2", "pkg3", "pkg4"}
	for i, name := range expectedNames {
		if pkgs[i].PackageName != name {
			t.Errorf("package %d: expected name %q, got %q", i, name, pkgs[i].PackageName)
		}
		if pkgs[i].Version != "latest" {
			t.Errorf("package %d: expected version 'latest', got %q", i, pkgs[i].Version)
		}
	}
}

func TestParseMixedFormats(t *testing.T) {
	content := "# Mixed format requirements file\nflask==3.1.0\nrequests @ https://example.com/requests-2.32.3.tar.gz\ngit+https://github.com/user/repo.git@main#egg=custom-pkg\ndjango>=4.2,<6.0\nmylib===1.0.dev5\n-r other-requirements.txt\n--index-url https://pypi.org/simple\n"
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "requirements.txt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &PypiParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 5 {
		t.Fatalf("expected 5 packages, got %d", len(pkgs))
	}

	// flask==3.1.0
	if pkgs[0].PackageName != "flask" || pkgs[0].Version != "3.1.0" {
		t.Errorf("pkg 0: got %q==%q, want flask==3.1.0", pkgs[0].PackageName, pkgs[0].Version)
	}
	// requests @ URL
	if pkgs[1].PackageName != "requests" || pkgs[1].Version != "latest" {
		t.Errorf("pkg 1: got %q==%q, want requests==latest", pkgs[1].PackageName, pkgs[1].Version)
	}
	// git+...#egg=custom-pkg
	if pkgs[2].PackageName != "custom-pkg" || pkgs[2].Version != "latest" {
		t.Errorf("pkg 2: got %q==%q, want custom-pkg==latest", pkgs[2].PackageName, pkgs[2].Version)
	}
	// django>=3.2,<4.0
	if pkgs[3].PackageName != "django" || pkgs[3].Version != "latest" {
		t.Errorf("pkg 3: got %q==%q, want django==latest", pkgs[3].PackageName, pkgs[3].Version)
	}
	// mylib===1.0.dev5
	if pkgs[4].PackageName != "mylib" || pkgs[4].Version != "1.0.dev5" {
		t.Errorf("pkg 4: got %q==%q, want mylib==1.0.dev5", pkgs[4].PackageName, pkgs[4].Version)
	}
}
