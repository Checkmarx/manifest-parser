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
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     1,
		EndIndex:       12,
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
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     4,
		EndIndex:       19,
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
		LineStart:      1,
		LineEnd:        1,
		StartIndex:     1,
		EndIndex:       16,
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
			LineStart:      4,
			LineEnd:        4,
			StartIndex:     1,
			EndIndex:       18,
		},
		{
			PackageManager: "pypi",
			PackageName:    "setuptools",
			Version:        "latest",
			FilePath:       filePath,
			LineStart:      5,
			LineEnd:        5,
			StartIndex:     1,
			EndIndex:       27,
		},
		{
			PackageManager: "pypi",
			PackageName:    "types-setuptools",
			Version:        "latest",
			FilePath:       filePath,
			LineStart:      6,
			LineEnd:        6,
			StartIndex:     1,
			EndIndex:       29,
		},
		{
			PackageManager: "pypi",
			PackageName:    "pytest",
			Version:        "7.1.3",
			FilePath:       filePath,
			LineStart:      7,
			LineEnd:        7,
			StartIndex:     1,
			EndIndex:       14,
		},
	}

	testdata.ValidatePackages(t, pkgs, expected)
}
