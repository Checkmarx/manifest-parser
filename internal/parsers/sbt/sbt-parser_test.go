package sbt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Checkmarx/manifest-parser/internal/testdata"
	"github.com/Checkmarx/manifest-parser/pkg/parser/models"
)

func TestParseSingleDependency(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % "1.0.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	expected := []models.Package{
		{
			PackageManager: "sbt",
			PackageName:    "org.example:test-lib",
			Version:        "1.0.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 23,
				EndIndex:   59,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseSingleDependencyDoublePercent(t *testing.T) {
	content := `libraryDependencies += "org.typelevel" %% "cats-core" % "2.9.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	expected := []models.Package{
		{
			PackageManager: "sbt",
			PackageName:    "org.typelevel:cats-core",
			Version:        "2.9.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 23,
				EndIndex:   63,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseSingleDependencyTriplePercent(t *testing.T) {
	content := `libraryDependencies += "org.scala-js" %%% "scalajs-dom" % "2.4.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	expected := []models.Package{
		{
			PackageManager: "sbt",
			PackageName:    "org.scala-js:scalajs-dom",
			Version:        "2.4.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 23,
				EndIndex:   65,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseSeqBlock(t *testing.T) {
	content := `libraryDependencies ++= Seq(
  "org.example" % "lib-a" % "1.0.0",
  "org.example" % "lib-b" % "2.0.0"
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "org.example:lib-a" {
		t.Errorf("expected pkg[0].PackageName = org.example:lib-a, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "1.0.0" {
		t.Errorf("expected pkg[0].Version = 1.0.0, got %s", pkgs[0].Version)
	}
	if pkgs[1].PackageName != "org.example:lib-b" {
		t.Errorf("expected pkg[1].PackageName = org.example:lib-b, got %s", pkgs[1].PackageName)
	}
	if pkgs[1].Version != "2.0.0" {
		t.Errorf("expected pkg[1].Version = 2.0.0, got %s", pkgs[1].Version)
	}
}

func TestParseWithScope(t *testing.T) {
	content := `libraryDependencies += "org.scalatest" %% "scalatest" % "3.2.15" % "test"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "org.scalatest:scalatest" {
		t.Errorf("expected PackageName = org.scalatest:scalatest, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "3.2.15" {
		t.Errorf("expected Version = 3.2.15, got %s", pkgs[0].Version)
	}
}

func TestParseWithVariableVersion(t *testing.T) {
	content := `val jacksonVersion = "2.13.0"
libraryDependencies += "com.fasterxml.jackson.core" % "jackson-databind" % jacksonVersion
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	if pkgs[0].Version != "2.13.0" {
		t.Errorf("expected Version = 2.13.0, got %s", pkgs[0].Version)
	}
}

func TestParseWithUnresolvableVariable(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % unknownVar
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	if pkgs[0].Version != "latest" {
		t.Errorf("expected Version = latest, got %s", pkgs[0].Version)
	}
}

func TestParseSingleLineComment(t *testing.T) {
	content := `// "org.example" % "should-not-parse" % "1.0.0"
libraryDependencies += "org.example" % "real-lib" % "1.0.0" // inline comment
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "org.example:real-lib" {
		t.Errorf("expected PackageName = org.example:real-lib, got %s", pkgs[0].PackageName)
	}
}

func TestParseBlockComment(t *testing.T) {
	content := `/*
  "org.example" % "should-not-parse" % "1.0.0"
*/
libraryDependencies += "org.example" % "real-lib" % "2.0.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "org.example:real-lib" {
		t.Errorf("expected PackageName = org.example:real-lib, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "2.0.0" {
		t.Errorf("expected Version = 2.0.0, got %s", pkgs[0].Version)
	}
}

func TestParseEmptyFile(t *testing.T) {
	content := ""
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 0 {
		t.Fatalf("expected 0 packages, got %d", len(pkgs))
	}
}

func TestParseDuplicateDependencies(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % "1.0.0"
libraryDependencies += "org.example" % "test-lib" % "2.0.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package (duplicate skipped), got %d", len(pkgs))
	}

	if pkgs[0].Version != "1.0.0" {
		t.Errorf("expected first occurrence version 1.0.0, got %s", pkgs[0].Version)
	}
}

func TestParseLocationAccuracy(t *testing.T) {
	// Line:       "org.example" % "test-lib" % "1.0.0"
	// Positions:   0123456789...
	content := `"org.example" % "test-lib" % "1.0.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	expected := []models.Package{
		{
			PackageManager: "sbt",
			PackageName:    "org.example:test-lib",
			Version:        "1.0.0",
			FilePath:       filePath,
			Locations: []models.Location{{
				Line:       0,
				StartIndex: 0,
				EndIndex:   36,
			}},
		},
	}
	testdata.ValidatePackages(t, pkgs, expected)
}

func TestParseNonExistentFile(t *testing.T) {
	parser := &SbtParser{}
	_, err := parser.Parse("/nonexistent/build.sbt")
	if err == nil {
		t.Error("expected error for non-existent file, got none")
	}
}

func TestParseMixedOperators(t *testing.T) {
	content := `libraryDependencies ++= Seq(
  "org.example" % "lib-a" % "1.0.0",
  "org.typelevel" %% "cats-core" % "2.9.0",
  "org.scala-js" %%% "scalajs-dom" % "2.4.0"
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "org.example:lib-a" {
		t.Errorf("expected pkg[0] = org.example:lib-a, got %s", pkgs[0].PackageName)
	}
	if pkgs[1].PackageName != "org.typelevel:cats-core" {
		t.Errorf("expected pkg[1] = org.typelevel:cats-core, got %s", pkgs[1].PackageName)
	}
	if pkgs[2].PackageName != "org.scala-js:scalajs-dom" {
		t.Errorf("expected pkg[2] = org.scala-js:scalajs-dom, got %s", pkgs[2].PackageName)
	}
}

func TestParseMalformedLine(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib"
libraryDependencies += "org.example" % "real-lib" % "1.0.0"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package (malformed skipped), got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "org.example:real-lib" {
		t.Errorf("expected PackageName = org.example:real-lib, got %s", pkgs[0].PackageName)
	}
}

func TestResolveVersion(t *testing.T) {
	vars := map[string]string{
		"jacksonVersion": "2.13.0",
		"log4jVersion":   "2.14.0",
	}

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"exact version", "1.2.3", "1.2.3"},
		{"variable lookup", "jacksonVersion", "2.13.0"},
		{"another variable", "log4jVersion", "2.14.0"},
		{"missing variable", "unknownVar", "latest"},
		{"empty version", "", "latest"},
		{"semver with pre-release", "2.0.0-RC1", "2.0.0-RC1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveVersion(tt.version, vars)
			if result != tt.expected {
				t.Errorf("resolveVersion(%q) = %q, want %q", tt.version, result, tt.expected)
			}
		})
	}
}

func TestParseWithVersionRanges(t *testing.T) {
	content := `libraryDependencies ++= Seq(
  "org.springframework" % "spring-core" % "[1.0.0,2.0.0)",
  "org.junit" % "junit" % "(1.0,2.0]"
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	if pkgs[0].Version != "latest" {
		t.Errorf("expected version 'latest' for range, got %q", pkgs[0].Version)
	}
	if pkgs[1].Version != "latest" {
		t.Errorf("expected version 'latest' for range, got %q", pkgs[1].Version)
	}
}

func TestParseWithPrefixWildcards(t *testing.T) {
	content := `libraryDependencies ++= Seq(
  "org.springframework" % "spring-core" % "1.0.+",
  "org.junit" % "junit" % "4.12.*"
)
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	if pkgs[0].Version != "latest" {
		t.Errorf("expected version 'latest' for wildcard, got %q", pkgs[0].Version)
	}
	if pkgs[1].Version != "latest" {
		t.Errorf("expected version 'latest' for wildcard, got %q", pkgs[1].Version)
	}
}

func TestStripComments(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		inBlockComment bool
		expected       string
		expectedBlock  bool
	}{
		{"no comments", `"org.example" % "lib" % "1.0"`, false, `"org.example" % "lib" % "1.0"`, false},
		{"single line comment", `"org.example" % "lib" % "1.0" // comment`, false, `"org.example" % "lib" % "1.0" `, false},
		{"full line comment", `// this is a comment`, false, ``, false},
		{"block comment start", `/* start of block`, false, ``, true},
		{"inside block comment", `  some content inside block`, true, ``, true},
		{"block comment end", `end of block */`, true, ``, false},
		{"inline block comment", `before /* inside */ after`, false, `before  after`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inBlock := tt.inBlockComment
			result := stripComments(tt.line, &inBlock)
			if result != tt.expected {
				t.Errorf("stripComments(%q) = %q, want %q", tt.line, result, tt.expected)
			}
			if inBlock != tt.expectedBlock {
				t.Errorf("inBlockComment = %v, want %v", inBlock, tt.expectedBlock)
			}
		})
	}
}

func TestExtractVariables(t *testing.T) {
	lines := []string{
		`val jacksonVersion = "2.13.0"`,
		`lazy val log4jVersion = "2.14.0"`,
		`def strutsVersion = "2.5.20"`,
		`// val commentedOut = "1.0.0"`,
		`name := "my-project"`,
		`val emptyLine`,
		`  val indentedVar = "3.0.0"`,
		`  lazy val indentedLazy = "4.0.0"`,
		`  def indentedDef = "5.0.0"`,
	}

	vars := extractVariables(lines)

	expected := map[string]string{
		"jacksonVersion": "2.13.0",
		"log4jVersion":   "2.14.0",
		"strutsVersion":  "2.5.20",
		"indentedVar":    "3.0.0",
		"indentedLazy":   "4.0.0",
		"indentedDef":    "5.0.0",
	}

	if len(vars) != len(expected) {
		t.Fatalf("expected %d variables, got %d: %v", len(expected), len(vars), vars)
	}

	for key, want := range expected {
		got, exists := vars[key]
		if !exists {
			t.Errorf("expected variable %q not found", key)
			continue
		}
		if got != want {
			t.Errorf("variable %q = %q, want %q", key, got, want)
		}
	}
}

func TestParseAddSbtPlugin(t *testing.T) {
	content := `addSbtPlugin("com.eed3si9n" % "sbt-assembly" % "2.1.0")
addSbtPlugin("org.scalameta" % "sbt-scalafmt" % "2.5.2")
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "plugins.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	if pkgs[0].PackageName != "com.eed3si9n:sbt-assembly" {
		t.Errorf("expected pkg[0].PackageName = com.eed3si9n:sbt-assembly, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "2.1.0" {
		t.Errorf("expected pkg[0].Version = 2.1.0, got %s", pkgs[0].Version)
	}
	if pkgs[1].PackageName != "org.scalameta:sbt-scalafmt" {
		t.Errorf("expected pkg[1].PackageName = org.scalameta:sbt-scalafmt, got %s", pkgs[1].PackageName)
	}
	if pkgs[1].Version != "2.5.2" {
		t.Errorf("expected pkg[1].Version = 2.5.2, got %s", pkgs[1].Version)
	}
}

func TestParseLazyVal(t *testing.T) {
	content := `lazy val myVersion = "3.1.0"
libraryDependencies += "org.example" % "test-lib" % myVersion
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "3.1.0" {
		t.Errorf("expected Version = 3.1.0, got %s", pkgs[0].Version)
	}
}

func TestParseDef(t *testing.T) {
	content := `def myVersion = "4.2.0"
libraryDependencies += "org.example" % "test-lib" % myVersion
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].Version != "4.2.0" {
		t.Errorf("expected Version = 4.2.0, got %s", pkgs[0].Version)
	}
}

func TestParseWithExclude(t *testing.T) {
	content := `libraryDependencies += "org.apache.hadoop" % "hadoop-common" % "3.3.4" exclude("org.slf4j", "slf4j-log4j12")
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "org.apache.hadoop:hadoop-common" {
		t.Errorf("expected PackageName = org.apache.hadoop:hadoop-common, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "3.3.4" {
		t.Errorf("expected Version = 3.3.4, got %s", pkgs[0].Version)
	}
	// EndIndex should NOT include the exclude(...) modifier
	loc := pkgs[0].Locations[0]
	rawLine := `libraryDependencies += "org.apache.hadoop" % "hadoop-common" % "3.3.4" exclude("org.slf4j", "slf4j-log4j12")`
	excludeStart := strings.Index(rawLine, " exclude(")
	if loc.EndIndex > excludeStart {
		t.Errorf("EndIndex %d extends into exclude(...) modifier (starts at %d)", loc.EndIndex, excludeStart)
	}
}

func TestParseWithIntransitive(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" intransitive()
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	// EndIndex should NOT include the intransitive() modifier
	loc := pkgs[0].Locations[0]
	rawLine := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" intransitive()`
	modifierStart := strings.Index(rawLine, " intransitive()")
	if loc.EndIndex > modifierStart {
		t.Errorf("EndIndex %d extends into intransitive() modifier (starts at %d)", loc.EndIndex, modifierStart)
	}
}

func TestParseWithCross(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" cross CrossVersion.full
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	loc := pkgs[0].Locations[0]
	rawLine := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" cross CrossVersion.full`
	modifierStart := strings.Index(rawLine, " cross ")
	if loc.EndIndex > modifierStart {
		t.Errorf("EndIndex %d extends into cross modifier (starts at %d)", loc.EndIndex, modifierStart)
	}
}

func TestParseWithExcludeAll(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" excludeAll(ExclusionRule("org.slf4j"))
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	loc := pkgs[0].Locations[0]
	rawLine := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" excludeAll(ExclusionRule("org.slf4j"))`
	modifierStart := strings.Index(rawLine, " excludeAll(")
	if loc.EndIndex > modifierStart {
		t.Errorf("EndIndex %d extends into excludeAll(...) modifier (starts at %d)", loc.EndIndex, modifierStart)
	}
}

func TestParseDependencyOverrides(t *testing.T) {
	content := `dependencyOverrides += "com.google.guava" % "guava" % "32.1.2-jre"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "com.google.guava:guava" {
		t.Errorf("expected PackageName = com.google.guava:guava, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "32.1.2-jre" {
		t.Errorf("expected Version = 32.1.2-jre, got %s", pkgs[0].Version)
	}
}

func TestParseWithClassifier(t *testing.T) {
	content := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" % "test" classifier "tests"
`
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "build.sbt")
	os.WriteFile(filePath, []byte(content), 0644)

	parser := &SbtParser{}
	pkgs, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}
	if pkgs[0].PackageName != "org.example:test-lib" {
		t.Errorf("expected PackageName = org.example:test-lib, got %s", pkgs[0].PackageName)
	}
	if pkgs[0].Version != "1.0.0" {
		t.Errorf("expected Version = 1.0.0, got %s", pkgs[0].Version)
	}
	loc := pkgs[0].Locations[0]
	rawLine := `libraryDependencies += "org.example" % "test-lib" % "1.0.0" % "test" classifier "tests"`
	modifierStart := strings.Index(rawLine, " classifier ")
	if loc.EndIndex > modifierStart {
		t.Errorf("EndIndex %d extends into classifier modifier (starts at %d)", loc.EndIndex, modifierStart)
	}
}

func TestSbtParser_Parse_RealFile(t *testing.T) {
	parser := &SbtParser{}
	manifestFile := "../../testdata/build.sbt"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify package count: 10 deps (log4j, cats, jackson, struts, commons, snakeyaml, netty, scalajs, hadoop, guava)
	if len(packages) != 10 {
		t.Fatalf("expected 10 packages, got %d", len(packages))
	}

	// Validate key fields for each package
	expected := []struct {
		name    string
		version string
		line    int
	}{
		{"org.apache.logging.log4j:log4j-core", "2.14.0", 10},
		{"org.typelevel:cats-core", "2.9.0", 13},
		{"com.fasterxml.jackson.core:jackson-databind", "2.13.0", 17},
		{"org.apache.struts:struts2-core", "2.5.20", 18},
		{"commons-collections:commons-collections", "3.2.1", 19},
		{"org.yaml:snakeyaml", "1.26", 20},
		{"io.netty:netty-codec-http", "4.1.68.Final", 21},
		{"org.scala-js:scalajs-dom", "2.4.0", 30},
		{"org.apache.hadoop:hadoop-common", "3.3.4", 33},
		{"com.google.guava:guava", "32.1.2-jre", 36},
	}

	for i, exp := range expected {
		if packages[i].PackageManager != "sbt" {
			t.Errorf("pkg[%d].PackageManager = %q, want %q", i, packages[i].PackageManager, "sbt")
		}
		if packages[i].PackageName != exp.name {
			t.Errorf("pkg[%d].PackageName = %q, want %q", i, packages[i].PackageName, exp.name)
		}
		if packages[i].Version != exp.version {
			t.Errorf("pkg[%d].Version = %q, want %q", i, packages[i].Version, exp.version)
		}
		if packages[i].FilePath != manifestFile {
			t.Errorf("pkg[%d].FilePath = %q, want %q", i, packages[i].FilePath, manifestFile)
		}
		if len(packages[i].Locations) != 1 {
			t.Errorf("pkg[%d] has %d locations, want 1", i, len(packages[i].Locations))
			continue
		}
		if packages[i].Locations[0].Line != exp.line {
			t.Errorf("pkg[%d].Location.Line = %d, want %d", i, packages[i].Locations[0].Line, exp.line)
		}
	}

	// Verify hadoop exclude modifier is NOT included in EndIndex
	hadoopPkg := packages[8]
	if hadoopPkg.Locations[0].EndIndex > 71 {
		t.Errorf("hadoop EndIndex %d should not extend into exclude(...) modifier", hadoopPkg.Locations[0].EndIndex)
	}
}

func TestSbtParser_Parse_PluginsFile(t *testing.T) {
	parser := &SbtParser{}
	manifestFile := "../../testdata/plugins.sbt"
	packages, err := parser.Parse(manifestFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expectedPackages := []models.Package{
		{
			PackageManager: "sbt",
			PackageName:    "com.eed3si9n:sbt-assembly",
			Version:        "2.1.0",
			FilePath:       manifestFile,
		},
		{
			PackageManager: "sbt",
			PackageName:    "org.scalameta:sbt-scalafmt",
			Version:        "2.5.2",
			FilePath:       manifestFile,
		},
		{
			PackageManager: "sbt",
			PackageName:    "com.github.sbt:sbt-native-packager",
			Version:        "1.9.16",
			FilePath:       manifestFile,
		},
		{
			PackageManager: "sbt",
			PackageName:    "org.apache.log4j:log4j-core",
			Version:        "2.14.1",
			FilePath:       manifestFile,
		},
		{
			PackageManager: "sbt",
			PackageName:    "org.apache.commons:commons-compress",
			Version:        "1.20",
			FilePath:       manifestFile,
		},
		{
			PackageManager: "sbt",
			PackageName:    "commons-io:commons-io",
			Version:        "2.4",
			FilePath:       manifestFile,
		},
	}

	if len(packages) != len(expectedPackages) {
		t.Fatalf("expected %d packages, got %d", len(expectedPackages), len(packages))
	}

	for i, pkg := range packages {
		if pkg.PackageManager != expectedPackages[i].PackageManager {
			t.Errorf("pkg[%d].PackageManager = %q, want %q", i, pkg.PackageManager, expectedPackages[i].PackageManager)
		}
		if pkg.PackageName != expectedPackages[i].PackageName {
			t.Errorf("pkg[%d].PackageName = %q, want %q", i, pkg.PackageName, expectedPackages[i].PackageName)
		}
		if pkg.Version != expectedPackages[i].Version {
			t.Errorf("pkg[%d].Version = %q, want %q", i, pkg.Version, expectedPackages[i].Version)
		}
	}
}
