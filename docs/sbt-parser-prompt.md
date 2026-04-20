# Reproducible Prompt: Add SBT Parser to manifest-parser

Use the following prompt with an AI coding assistant to reproduce the exact same SBT parser implementation. Copy everything below the line.

---

## Prompt

I have a Go repository (`github.com/Checkmarx/manifest-parser`) that parses package manifest files and extracts dependency information. It already supports Maven, npm, PyPI, Go modules, and .NET. It is used as an internal Go module consumed by a parent package.

### Existing Architecture (do NOT modify these files — only extend)

**Common types** (`pkg/parser/models/package_model.go`):
```go
type Location struct {
    Line       int
    StartIndex int
    EndIndex   int
}
type Package struct {
    PackageManager string
    PackageName    string
    Version        string
    FilePath       string
    Locations      []Location
}
```

**Parser interface** (`pkg/parser/parser.go`):
```go
type Parser interface {
    Parse(manifestFile string) ([]models.Package, error)
}
```

**File detection** (`pkg/parser/manifest-file-selector.go`): Uses a `Manifest` iota enum and `selectManifestFile()` function that matches filenames/extensions. Example: `.csproj` extension match for dotnet, `"pom.xml"` exact match for Maven.

**Factory** (`pkg/parser/parser_factory.go`): `ParsersFactory(manifest string) Parser` with a switch on the Manifest type. Each parser lives in its own package under `internal/parsers/<name>/`.

**Test patterns**: Table-driven tests with `t.TempDir()`, inline content strings, `testdata.ValidatePackages()` helper from `internal/testdata/helper.go`. Real file tests against fixtures in `internal/testdata/`.

**Existing PackageManager identifiers**: `"mvn"`, `"npm"`, `"pypi"`, `"go"`, `"nuget"`.

**Version resolution convention**: Exact version returned as-is. Ranges/specifiers/empty return `"latest"`.

### Task: Add SBT (Scala Build Tool) Parser Support

Implement a production-grade SBT parser following these exact specifications:

#### 1. Create `internal/parsers/sbt/sbt-parser.go`

- **Package**: `sbt` | **Struct**: `SbtParser{}` | **PackageManager string**: `"sbt"`
- **Two-pass, regex-based parsing** (similar to PyPI parser style):
  - **Pass 1**: Extract variable definitions (`val`, `lazy val`, `def`) into a `map[string]string`
  - **Pass 2**: Line-by-line dependency extraction using regex
- **Variable regex** (combined `val`, `lazy val`, `def`):
  ```
  ^\s*(?:lazy\s+)?(?:val|def)\s+(\w+)\s*=\s*"([^"]+)"
  ```
- **Dependency regex**:
  ```
  "([^"]+)"\s+(%{1,3})\s+"([^"]+)"\s+%\s+(?:"([^"]+)"|(\w+))(?:\s+%\s+(?:"[^"]*"|\w+))?
  ```
  Captures: groupId, operator (`%`/`%%`/`%%%`), artifactId, version (quoted string or bare variable name), optional scope (ignored). The `%%`/`%%%` operators are captured but NOT used — PackageName is always `"groupId:artifactId"` without Scala version suffix.
- **Helper functions**:
  - `extractVariables(lines []string) map[string]string` — scans for `val`/`lazy val`/`def` declarations, respects block comment state
  - `resolveVersion(version string, vars map[string]string) string` — returns exact version if starts with digit, looks up variable map otherwise, returns `"latest"` if unresolvable or empty
  - `stripComments(line string, inBlockComment *bool) string` — handles `//` single-line and `/* */` multi-line block comments (including inline `/* ... */` on same line)
  - `computeLocationIndices(rawLine string, groupId string) (int, int)` — calculates StartIndex (position of first `"` of groupId) and EndIndex (end of core dependency, EXCLUDING modifiers)
- **Modifier-aware EndIndex**: The `computeLocationIndices` function must trim these modifier keywords from the end BEFORE trimming trailing punctuation (order matters — `intransitive()` has parens that get stripped by punctuation trimmer):
  ```go
  var modifierKeywords = []string{
      "exclude(", "excludeAll(", "intransitive()", "withSources()",
      "withJavadoc()", "classifier ", "classifier(", "cross ", "cross(",
  }
  ```
  Also trim trailing `)`, `,`, whitespace via `trimTrailingPunctuation`.
- **Duplicate detection**: `map[string]bool` keyed by `"groupId:artifactId"`. Skip duplicates silently (NO `log.Printf` — this is a library, not a CLI).
- **Line ending normalization**: `strings.ReplaceAll(string(content), "\r\n", "\n")` before splitting.
- **Imports**: Only `fmt`, `os`, `regexp`, `strings` + `models` package. No `log` import.
- **Location tracking**: Single `Location` per package, `Line` is 0-indexed.

#### 2. Create `internal/testdata/build.sbt`

Test fixture exercising all parsing scenarios with known-vulnerable packages:
```scala
// Project settings
name := "vulnerable-test-project"
version := "1.0.0"
scalaVersion := "2.13.12"

val jacksonVersion = "2.13.0"
lazy val log4jVersion = "2.14.0"
def strutsVersion = "2.5.20"

// Single dependency with % — CVE-2021-44228 (Log4Shell)
libraryDependencies += "org.apache.logging.log4j" % "log4j-core" % log4jVersion

// Single dependency with %% — safe dependency
libraryDependencies += "org.typelevel" %% "cats-core" % "2.9.0"

// Seq block with mixed operators and vulnerable packages
libraryDependencies ++= Seq(
  "com.fasterxml.jackson.core" % "jackson-databind" % jacksonVersion,
  "org.apache.struts" % "struts2-core" % strutsVersion,
  "commons-collections" % "commons-collections" % "3.2.1",
  "org.yaml" % "snakeyaml" % "1.26",
  "io.netty" %% "netty-codec-http" % "4.1.68.Final" % "test"
)

/*
  This is a block comment — dependencies here should NOT be parsed
  "org.example" % "should-not-parse" % "1.0.0"
*/

// Scala.js dependency with %%%
libraryDependencies += "org.scala-js" %%% "scalajs-dom" % "2.4.0"

// Dependency with exclude modifier
libraryDependencies += "org.apache.hadoop" % "hadoop-common" % "3.3.4" exclude("org.slf4j", "slf4j-log4j12")

// Dependency override
dependencyOverrides += "com.google.guava" % "guava" % "32.1.2-jre"
```

#### 3. Create `internal/testdata/plugins.sbt`

```scala
// SBT plugins
addSbtPlugin("com.eed3si9n" % "sbt-assembly" % "2.1.0")
addSbtPlugin("org.scalameta" % "sbt-scalafmt" % "2.5.2")
addSbtPlugin("com.github.sbt" % "sbt-native-packager" % "1.9.16")
```

#### 4. Create `internal/parsers/sbt/sbt-parser_test.go`

Write 29 comprehensive tests following existing Maven/PyPI test patterns. Include:

**Core parsing tests** (use `t.TempDir()` with inline content):
1. `TestParseSingleDependency` — basic `"g" % "a" % "v"` with exact location validation via `testdata.ValidatePackages`
2. `TestParseSingleDependencyDoublePercent` — `%%` operator, verify PackageName is `g:a` (no Scala suffix)
3. `TestParseSingleDependencyTriplePercent` — `%%%` operator (Scala.js), same as `%%`
4. `TestParseSeqBlock` — `libraryDependencies ++= Seq(...)` with multiple deps
5. `TestParseWithScope` — trailing `% "test"`, verify scope is ignored
6. `TestParseWithVariableVersion` — `val v = "1.0"` then `% v`, verify resolution
7. `TestParseWithUnresolvableVariable` — missing variable, version should be `"latest"`
8. `TestParseSingleLineComment` — `//` comments skipped, inline comment after dep works
9. `TestParseBlockComment` — `/* ... */` spanning lines, deps inside skipped
10. `TestParseEmptyFile` — returns empty slice, no error
11. `TestParseDuplicateDependencies` — same `g:a` twice, first wins, second skipped
12. `TestParseLocationAccuracy` — bare `"g" % "a" % "v"` line, verify exact indices
13. `TestParseNonExistentFile` — returns error
14. `TestParseMixedOperators` — mix of `%`, `%%`, `%%%` in same Seq
15. `TestParseMalformedLine` — missing version part, gracefully skipped

**Production hardening tests**:
16. `TestParseAddSbtPlugin` — `addSbtPlugin(...)` syntax, 2 plugins parsed correctly
17. `TestParseLazyVal` — `lazy val v = "3.1.0"`, variable resolved
18. `TestParseDef` — `def v = "4.2.0"`, variable resolved
19. `TestParseWithExclude` — `exclude(...)` modifier, EndIndex must NOT include it
20. `TestParseWithIntransitive` — `intransitive()` modifier, EndIndex must NOT include it
21. `TestParseWithCross` — `cross CrossVersion.full` modifier, EndIndex must NOT include it
22. `TestParseWithExcludeAll` — `excludeAll(...)` modifier, EndIndex must NOT include it
23. `TestParseDependencyOverrides` — `dependencyOverrides +=` parsed correctly
24. `TestParseWithClassifier` — `classifier "tests"` modifier, EndIndex must NOT include it

**Helper function tests**:
25. `TestResolveVersion` — table-driven: exact, variable lookup, missing, empty, semver with pre-release
26. `TestStripComments` — table-driven: no comments, single-line, full-line, block start/end, inline block
27. `TestExtractVariables` — `val`, `lazy val`, `def`, commented-out (skipped), indented variants

**Real file tests**:
28. `TestSbtParser_Parse_RealFile` — parse `../../testdata/build.sbt`, validate all 10 packages (log4j, cats, jackson, struts, commons-collections, snakeyaml, netty, scalajs, hadoop, guava) with correct names, versions, and line numbers
29. `TestSbtParser_Parse_PluginsFile` — parse `../../testdata/plugins.sbt`, validate 3 plugin packages

#### 5. Modify `pkg/parser/manifest-file-selector.go`

- Add `SbtBuild` to the `Manifest` iota enum after `GoMod`
- Add **extension-based** detection (NOT exact filename): `if manifestFileExtension == ".sbt" { return SbtBuild }` — this matches all `.sbt` files (`build.sbt`, `plugins.sbt`, `dependencies.sbt`, etc.), following the same pattern used for `.csproj`

#### 6. Modify `pkg/parser/parser_factory.go`

- Add import: `"github.com/Checkmarx/manifest-parser/internal/parsers/sbt"`
- Add case: `case SbtBuild: return &sbt.SbtParser{}`

#### 7. Modify `pkg/parser/manifest-file-selector_test.go`

Add 3 tests:
- `TestManifestFileSelector_ExpectSbtBuild` — `"build.sbt"` → `SbtBuild`
- `TestManifestFileSelector_ExpectSbtPlugins` — `"plugins.sbt"` → `SbtBuild`
- `TestManifestFileSelector_ExpectSbtCustom` — `"dependencies.sbt"` → `SbtBuild`

#### 8. Update `README.md`

Add SBT to the supported package managers table, usage examples, and project structure.

### Critical Implementation Details

1. **Order in `computeLocationIndices`**: Run `trimModifiers` BEFORE `trimTrailingPunctuation`. If you reverse the order, `intransitive()` will have its `)` stripped first and the modifier keyword won't match.
2. **No `log` package**: This is a library module. Do NOT use `log.Printf` for duplicate warnings or any other logging. Skip duplicates silently.
3. **`\r\n` normalization**: Add `strings.ReplaceAll(string(content), "\r\n", "\n")` before `strings.Split` — required for Windows compatibility.
4. **Regex is context-free**: The dependency regex matches `"g" % "a" % "v"` anywhere on a line, regardless of prefix (`libraryDependencies +=`, `addSbtPlugin(`, `dependencyOverrides +=`, bare declaration). This is intentional — no need to check the prefix.
5. **`%%`/`%%%` operators**: Captured by regex but ignored in output. PackageName is always `groupId:artifactId`.

### Verification

After implementation, run:
```bash
go build ./...
go test ./internal/parsers/sbt/ -v -cover   # expect 29 tests, ~97.8% coverage
go test ./pkg/parser/ -v                      # expect 10 selector tests pass
go run cmd/main.go internal/testdata/build.sbt
go run cmd/main.go internal/testdata/plugins.sbt
```
