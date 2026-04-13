# SBT Parser Implementation Plan

## Context

The manifest-parser repository supports Maven, npm, PyPI, Go modules, and .NET. The user needs to extend it with SBT (Scala Build Tool) support to parse SBT manifest files and extract dependencies. The implementation must follow existing patterns exactly, add duplicate detection, include comprehensive tests with vulnerable packages, and integrate cleanly without modifying existing parsers.

### Supported SBT File Types

SBT uses multiple file types that can declare dependencies. The parser supports **all `.sbt` files** via extension-based matching (like `.csproj` for dotnet):

| File | Purpose | Syntax |
|------|---------|--------|
| `build.sbt` | Primary build definition | `libraryDependencies += "g" % "a" % "v"` |
| `plugins.sbt` | SBT plugin dependencies (in `project/`) | `addSbtPlugin("g" % "a" % "v")` |
| `dependencies.sbt` | Separated dependency definitions | Same as `build.sbt` |
| Any other `*.sbt` | SBT auto-loads all `.sbt` files in project root | Same as `build.sbt` |

The core dependency regex `"g" % "a" % "v"` matches inside any wrapper (`addSbtPlugin(...)`, `libraryDependencies +=`, bare declarations), so all these file types are handled by the same parser with no special-casing needed.

---

## Files to Create (3)

### 1. `internal/parsers/sbt/sbt-parser.go` — Core Parser

**Package:** `sbt` | **Struct:** `SbtParser{}` | **PackageManager string:** `"sbt"`

**Parsing Strategy — Two-pass, regex-based (like PyPI parser but with Seq-block state tracking):**

- **Pass 1:** Extract variable definitions into `map[string]string`
- **Pass 2:** Line-by-line dependency extraction with state machine for `Seq(...)` blocks

#### Variable Extraction (Pass 1)

Supports all Scala variable declaration forms used in SBT files:

| Pattern | Example | Regex |
|---------|---------|-------|
| `val` | `val v = "1.0"` | `^\s*val\s+(\w+)\s*=\s*"([^"]+)"` |
| `lazy val` | `lazy val v = "1.0"` | `^\s*lazy\s+val\s+(\w+)\s*=\s*"([^"]+)"` |
| `def` | `def v = "1.0"` | `^\s*def\s+(\w+)\s*=\s*"([^"]+)"` |

All three patterns are combined into a single regex:
```
^\s*(?:lazy\s+)?(?:val|def)\s+(\w+)\s*=\s*"([^"]+)"
```

#### Dependency Extraction (Pass 2)

**Core dependency regex:**
```
"([^"]+)"\s+(%{1,3})\s+"([^"]+)"\s+%\s+(?:"([^"]+)"|(\w+))(?:\s+%\s+(?:"[^"]*"|\w+))?
```
Captures: groupId, operator (`%`/`%%`/`%%%`), artifactId, version (quoted or variable name), optional scope (ignored).

#### Helper functions:
- `extractVariables(lines []string) map[string]string` — supports `val`, `lazy val`, and `def`
- `resolveVersion(version string, vars map[string]string) string` — exact version as-is, variable lookup, unresolvable → `"latest"`
- `stripComments(line string, inBlockComment *bool) string` — handles `//` and `/* */`
- `computeLocationIndices(rawLine, groupId) (int, int)` — calculates start/end with modifier-aware trimming

#### Duplicate detection:
`map[string]bool` keyed by `"groupId:artifactId"`. Skip duplicates silently (no `log.Printf` — this is a library, not a CLI; callers control their own logging).

#### Comment handling:
Strip `//` inline comments; track `/* */` block comment state across lines. The `//` stripping is applied **after** the dependency regex match on the raw line, so `//` inside quoted strings in dependency declarations won't cause false truncation.

#### Location tracking:
Single `Location` per package (like PyPI), `Line` is 0-indexed:
- `StartIndex` = position of first `"` of groupId in the raw line
- `EndIndex` = end of the dependency declaration, **excluding** trailing modifiers

**Modifier-aware EndIndex calculation:** The `computeLocationIndices` function trims the following patterns from the end of the line when computing `EndIndex`:
- Trailing commas and whitespace
- Dependency modifiers: `exclude(...)`, `excludeAll(...)`, `classifier(...)`, `intransitive()`, `withSources()`, `withJavadoc()`, `cross(...)`
- Closing parentheses from `addSbtPlugin(...)` or `Seq(...)` wrappers
- Inline comments (`// ...`)

This ensures the location span covers only the `"g" % "a" % "v"` core declaration.

#### Imports:
Only stdlib — `os`, `regexp`, `strings`, `fmt` + `models` package. No `log` import (library code should not write to stderr).

### 2. `internal/testdata/build.sbt` and `internal/testdata/plugins.sbt` — Test Fixtures

**`build.sbt`** — Contains known-vulnerable dependencies:
- **log4j-core 2.14.0** (CVE-2021-44228 — Log4Shell)
- **jackson-databind 2.13.0** (multiple CVEs)
- **struts2-core 2.5.20** (CVE-2020-17530)
- **commons-collections 3.2.1** (deserialization vulnerability)
- **snakeyaml 1.26** (CVE-2022-1471)

Exercises all parsing scenarios: `%`, `%%`, `%%%`, `Seq(...)`, variable-based versions, inline comments, block comments, scope annotations.

**`plugins.sbt`** — Contains SBT plugin dependencies using `addSbtPlugin(...)` syntax to validate that the parser handles `plugins.sbt` files correctly.

### 3. `internal/parsers/sbt/sbt-parser_test.go` — Comprehensive Tests

**Table-driven + individual tests following Maven/PyPI patterns:**

| # | Test | What it validates |
|---|------|-------------------|
| 1 | TestParseSingleDependency | Basic `libraryDependencies += "g" % "a" % "v"` |
| 2 | TestParseSingleDependencyDoublePercent | `%%` operator → PackageName is `g:a` (no Scala suffix) |
| 3 | TestParseSingleDependencyTriplePercent | `%%%` operator (Scala.js) → same as `%%` |
| 4 | TestParseSeqBlock | `libraryDependencies ++= Seq(...)` with multiple deps |
| 5 | TestParseWithScope | Trailing `% "test"` or `% Test` → parsed correctly, scope ignored |
| 6 | TestParseWithVariableVersion | `val v = "1.0"` then `% v` → resolves to `"1.0"` |
| 7 | TestParseWithUnresolvableVariable | Missing variable → version is `"latest"` |
| 8 | TestParseSingleLineComment | `//` comments are skipped |
| 9 | TestParseBlockComment | `/* ... */` spanning lines → deps inside skipped |
| 10 | TestParseEmptyFile | Returns empty slice, no error |
| 11 | TestParseDuplicateDependencies | Same `g:a` twice → first wins, second skipped |
| 12 | TestParseLocationAccuracy | Verify exact Line, StartIndex, EndIndex values |
| 13 | TestParseNonExistentFile | Returns error |
| 14 | TestParseMixedOperators | Mix of `%` and `%%` in same Seq |
| 15 | TestResolveVersion | Table-driven: exact, variable, missing, empty |
| 16 | TestParseAddSbtPlugin | `addSbtPlugin("g" % "a" % "v")` syntax from `plugins.sbt` |
| 17 | TestParseLazyVal | `lazy val v = "1.0"` → variable extracted and resolved |
| 18 | TestParseDef | `def v = "1.0"` → variable extracted and resolved |
| 19 | TestParseWithExclude | `"g" % "a" % "v" exclude("x", "y")` → parsed, EndIndex excludes modifier |
| 20 | TestParseWithIntransitive | `"g" % "a" % "v" intransitive()` → parsed, EndIndex excludes modifier |
| 21 | TestParseWithCross | `"g" % "a" % "v" cross CrossVersion.full` → parsed, EndIndex excludes modifier |
| 22 | TestParseWithExcludeAll | `"g" % "a" % "v" excludeAll(...)` → parsed, EndIndex excludes modifier |
| 23 | TestParseDependencyOverrides | `dependencyOverrides += "g" % "a" % "v"` → parsed correctly |
| 24 | TestParseWithClassifier | `"g" % "a" % "v" % "test" classifier "tests"` → parsed, classifier ignored |
| 25 | TestExtractVariables | Table-driven: val, lazy val, def, commented out, indented |
| 26 | TestSbtParser_Parse_RealFile | Parse `../../testdata/build.sbt` and validate against expected packages |
| 27 | TestSbtParser_Parse_PluginsFile | Parse `../../testdata/plugins.sbt` and validate plugin dependencies |

---

## Files to Modify (3)

### 4. `pkg/parser/manifest-file-selector.go`

- Add `SbtBuild` to the `Manifest` iota enum (after `GoMod`)
- Add extension-based detection: `if manifestFileExtension == ".sbt" { return SbtBuild }`
  - This matches **all** `.sbt` files (`build.sbt`, `plugins.sbt`, `dependencies.sbt`, etc.)
  - Follows the same pattern used for `.csproj` detection

### 5. `pkg/parser/parser_factory.go`

- Add import: `"github.com/Checkmarx/manifest-parser/internal/parsers/sbt"`
- Add case: `case SbtBuild: return &sbt.SbtParser{}`

### 6. `pkg/parser/manifest-file-selector_test.go`

- Add `TestManifestFileSelector_ExpectSbtBuild` test for `build.sbt`
- Add `TestManifestFileSelector_ExpectSbtPlugins` test for `plugins.sbt`
- Add `TestManifestFileSelector_ExpectSbtCustom` test for `dependencies.sbt`

---

## Implementation Order

1. Create `internal/parsers/sbt/sbt-parser.go` (core parser)
2. Create `internal/testdata/build.sbt` (test fixture)
3. Create `internal/parsers/sbt/sbt-parser_test.go` (tests)
4. Modify `pkg/parser/manifest-file-selector.go` (enum + detection)
5. Modify `pkg/parser/manifest-file-selector_test.go` (selector test)
6. Modify `pkg/parser/parser_factory.go` (factory registration)
7. Run `go test ./...` to verify all tests pass with no regressions

## Verification

1. `go build ./...` — compiles cleanly
2. `go test ./internal/parsers/sbt/ -v` — all SBT parser tests pass
3. `go test ./pkg/parser/ -v` — selector + factory tests pass (including new SBT test)
4. `go test ./... -v` — full suite, no regressions
5. `go test ./... -cover` — check coverage
6. `go run cmd/main.go internal/testdata/build.sbt` — produces correct JSON output
7. `go run cmd/main.go internal/testdata/plugins.sbt` — produces correct JSON output for plugin dependencies

---

## Production-Readiness Hardening (v2)

The following gaps were identified after initial implementation and are addressed in the updated parser:

| # | Gap | Fix | Impact |
|---|-----|-----|--------|
| 1 | `lazy val` not matched | Extend varRegex to `(?:lazy\s+)?(?:val\|def)` | **High** — many real projects use `lazy val` |
| 2 | `def` declarations not matched | Same regex extension | **Medium** — some projects use `def` for versions |
| 3 | Modifiers corrupt EndIndex | `computeLocationIndices` trims `exclude(...)`, `intransitive()`, `withSources()`, `withJavadoc()`, `cross(...)`, `classifier(...)` | **Medium** — common in complex builds |
| 4 | Closing `)` from wrappers in EndIndex | Trim trailing `)` after modifiers | **Medium** — affects `addSbtPlugin(...)` |
| 5 | `log.Printf` in library code | Remove all `log.Printf` calls — library consumers control their own logging | **Medium** — breaks clean library usage |
| 6 | `dependencyOverrides` not tested | Already works (regex is context-free), add explicit test | **Low** — verification only |
| 7 | `classifier` keyword | Already handled by optional scope group in regex, add explicit test | **Low** — verification only |