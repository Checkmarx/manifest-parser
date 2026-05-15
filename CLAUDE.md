# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go module that parses package manifests from multiple ecosystems (Maven, npm, Python, Go, .NET, Gradle) and returns each declared dependency along with the **exact line/character range** of its declaration. Consumed by [AST-CLI](https://github.com/Checkmarx/ast-cli) to correlate manifest entries with Checkmarx runtime scans — so the `Locations` field is part of the public contract, not a debugging convenience.

**Status:** Active / maintained. Part of the Checkmarx One SCA pipeline.

## Technology Stack

| Component | Details |
|-----------|---------|
| Language | Go 1.23.0 / toolchain go1.24.2 |
| Test framework | `github.com/stretchr/testify v1.8.4` |
| Go module parsing | `golang.org/x/mod v0.24.0` |
| XML parsing | stdlib `encoding/xml` |
| JSON output | stdlib `encoding/json` |
| Database | None |
| Web framework | None |
| Dependencies | Vendored (`vendor/`) — no `go mod download` required |

## Repository Structure

```
cmd/                            CLI entry point (main.go)
pkg/parser/                     Public API
  ├── parser.go                 Parser interface
  ├── parser_factory.go         ParsersFactory — sole public entry point
  ├── manifest-file-selector.go Filename → Manifest enum mapping
  └── models/
      └── package_model.go      Package / Location structs
internal/parsers/               Per-ecosystem implementations (not importable by callers)
  ├── dotnet/
  ├── golang/
  ├── gradle/
  ├── maven/
  ├── npm/
  └── pypi/
test/resources/                 Shared fixture files for all parser tests
vendor/                         Vendored dependencies
.github/workflows/ci.yml        CI pipeline (test + 60% coverage gate)
```

## Development Setup

**Prerequisites:** Go ≥ 1.23, git. No other tools required — dependencies are vendored.

```bash
git clone https://github.com/Checkmarx/manifest-parser.git
cd manifest-parser

go test ./...                          # run all tests
go test ./internal/parsers/gradle/...  # run tests for a single parser
go test -run TestName ./path/...       # run a single test by name
go test ./... -coverprofile cover.out  # CI gate: total coverage must be >= 60%
go tool cover -html cover.out          # view coverage report in browser
go build -o manifest-parser ./cmd      # build CLI
go run ./cmd test/resources/pom.xml    # run CLI against a fixture
```

**Sample output** from `go run ./cmd test/resources/pom.xml`:

```json
[
  {
    "PackageManager": "mvn",
    "PackageName": "junit:junit",
    "Version": "4.13.2",
    "FilePath": "test/resources/pom.xml",
    "Locations": [{ "Line": 14, "StartIndex": 4, "EndIndex": 20 }]
  }
]
```

## API / Interfaces

The public API lives entirely under `pkg/`:

**`Parser` interface** ([pkg/parser/parser.go](pkg/parser/parser.go)):
```go
type Parser interface {
    Parse(manifestFile string) ([]models.Package, error)
}
```

**`ParsersFactory`** ([pkg/parser/parser_factory.go](pkg/parser/parser_factory.go)) — the **only** public entry point. Returns a concrete `Parser` for the given filename, or `nil` for unsupported files.

**`Package` / `Location` structs** ([pkg/parser/models/package_model.go](pkg/parser/models/package_model.go)):
```go
type Package struct {
    PackageManager string
    PackageName    string
    Version        string
    FilePath       string
    Locations      []Location
}

type Location struct {
    Line       int  // 0-based (all parsers)
    StartIndex int  // 0-based byte offset from start of line
    EndIndex   int  // 0-based byte offset from start of line
}
```

Adding a new ecosystem: edit `manifest-file-selector.go`, add a case in `parser_factory.go`, and add a package under `internal/parsers/`.

## Architecture

The module is organized around one interface and a dispatching factory:

- [pkg/parser/manifest-file-selector.go](pkg/parser/manifest-file-selector.go) — maps filename/extension to a `Manifest` enum.
- [pkg/parser/parser_factory.go](pkg/parser/parser_factory.go) — dispatches to the right concrete parser.
- [pkg/parser/models/package_model.go](pkg/parser/models/package_model.go) — `Locations` is a slice: Maven returns one entry per line of a multi-line `<dependency>` block; most others return a single entry.

Per-ecosystem parsers live under [internal/parsers/](internal/parsers/):

- `gradle/` — parses `build.gradle` / `build.gradle.kts` (Groovy + Kotlin DSL) and `gradle/libs.versions.toml` version catalogs. Resolves variables from `gradle.properties` and `ext {}` blocks. `PackageManager` = `"gradle"`.
- `maven/` — parses `pom.xml` with `encoding/xml`, then re-scans the raw text to locate each `<dependency>` block line by line. Resolves `${property}` vars from `<properties>` and falls back to `<dependencyManagement>` for empty/ranged versions. Only **direct** `<dependencies>` are emitted (managed-only deps are intentionally skipped to avoid duplicates — see PR #15). `PackageManager` = `"mvn"`.
- `npm/` — parses `package.json` plus, if present as a sibling file, `package-lock.json` (v1 and v2/v3 formats). Ranged specifiers (`^`, `~`, `*`, `>`, `<`) trigger a lookup in the lockfile. Without a lock match, ranged versions resolve to `"latest"`. `PackageManager` = `"npm"`.
- `pypi/` — line-oriented scan of `requirements*.txt` / `packages*.txt`. **Only `package==version` is supported.** Comments (`#`) and environment markers (`;`) are stripped. `PackageManager` = `"pypi"`.
- `golang/` — uses `golang.org/x/mod/modfile` to parse `go.mod`, then uses the parser's line metadata to compute character offsets. `PackageManager` = `"go"`.
- `dotnet/` — three parsers: `csproj_parser.go` (`.csproj`), `directory_packages_props_parser.go` (central package management), `packages_config_parser.go` (legacy). Bracketed version ranges become `"latest"`. `PackageManager` = `"nuget"` for all three.
- `sbt/` — parses any `.sbt` file (`build.sbt`, `project/plugins.sbt`, `version.sbt`, etc.) using line-oriented scanning. Supports Scala build dependencies in `"group" % "name" % "version"` format. `PackageManager` = `"sbt"`.

## Project Rules (Invariants)

- **`Location.Line` MUST be 0-based for ALL parsers.** When iterating `for i, line := range lines`, emit `Line: i` — never `i + 1`. Editors display 1-based line numbers; downstream consumers add `+1` for display. If parser output matches the editor's line number, it's off-by-one.
- **`Location.StartIndex` / `EndIndex` are 0-based byte offsets** from the start of the line. They are byte offsets, not rune/character offsets — relevant for non-ASCII manifests.
- **Unresolvable or ranged versions resolve to the literal string `"latest"`**, never an empty string. Callers branch on this value.
- **`PackageManager` strings are part of the contract**: `"gradle"`, `"mvn"`, `"npm"`, `"pypi"`, `"go"`, `"nuget"`, `"sbt"`. Don't rename them.
- Maven emits one `Location` per **non-comment line** of the `<dependency>` block (open tag, each child element, close tag). Single-line `Locations` for Maven would be a regression.
- Do not add `ParsersFactory` overloads or alternative entry points without coordinating with AST-CLI.
- **Do not modify or rename existing `PackageManager` strings**. AST-CLI and Checkmarx One SCA branch on these values — a silent rename breaks downstream parsing with no compile-time error. If a rename is genuinely required, stop and confirm with the user.

## Testing Strategy

Each parser has a `*_test.go` co-located with it using `testify`. Shared fixtures live in [test/resources/](test/resources/):

```
test/resources/
├── build.gradle              Groovy DSL
├── build.gradle.kts          Kotlin DSL
├── gradle/libs.versions.toml Version catalog (80+ entries)
├── gradle.properties         Centralized Gradle properties
├── pom.xml                   Maven
├── package.json              npm
├── test_go.mod               Go modules
├── Bootstrap.csproj          .NET csproj
├── Gateway.csproj            .NET csproj (variant)
├── Directory.Packages.props  .NET centralized packages
├── packages.config           .NET legacy NuGet
└── requirements.txt          Python pip
```

When adding behaviours, add a fixture here rather than embedding large manifests in test source.

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) enforces a **60% total coverage floor** — adding an untested branch to an already-thin package can push the whole repo below the gate. View coverage locally with `go tool cover -html cover.out`.

Expected pattern for a new parser: fixture file under `test/resources/` + `<ecosystem>_parser_test.go` co-located with the parser, using `testify` assertions on `PackageName`, `Version`, `PackageManager`, and `Locations`.

## Known Issues / Limitations

- **pypi**: Only `package==version` syntax is supported. `pip freeze`, Poetry, and pip-tools output are explicitly out of scope.
- **npm**: Ranged version specifiers (`^`, `~`, `*`, `>`, `<`) without a matching `package-lock.json` entry resolve to `"latest"` rather than the actual installed version.
- **Maven**: Managed-only deps (present in `<dependencyManagement>` but not in `<dependencies>`) are not emitted, to avoid duplicating entries already declared in a BOM consumer.
- **dotnet**: Bracketed version ranges (e.g., `[1.0,2.0)`) become `"latest"`.
- **All parsers**: Direct dependencies only — transitive dependencies are not resolved or scanned.

## External Integrations

- **AST-CLI** ([Checkmarx/ast-cli](https://github.com/Checkmarx/ast-cli)) — primary consumer. Imports this module as a Go library. The fields `Locations`, `PackageManager`, `PackageName`, and `Version` on the `Package` struct are load-bearing: AST-CLI uses them to annotate scan results and drive remediation UI.
- **Checkmarx One SCA** — downstream scan engine that receives the parsed dependency list.

## Deployment

N/A — this is a Go library consumed via `go get github.com/Checkmarx/manifest-parser`. It is not deployed as a service. The CLI (`cmd/`) is a local testing convenience, not a production artifact.

## Performance Considerations

- Maven re-scans the raw XML bytes after `encoding/xml` parsing (two passes). Large `pom.xml` files are loaded fully into memory; there is no streaming.
- Gradle version catalog parsing reads `libs.versions.toml` once, separately from the build file. Large catalogs (80+ entries) are fine; pathologically large files are not size-bounded.
- No caching between calls to `ParsersFactory` — each invocation allocates fresh parser state.

## Security & Access

- Parsers consume **untrusted manifest files** (user-supplied input):
  - `encoding/xml` does **not** resolve external entities or DTDs by default — XXE is not a risk with the standard library decoder.
  - There is no file-size limit enforced before reading. Callers in adversarial environments should validate file size before calling `Parse`.
  - Path traversal: `ParsersFactory` accepts an arbitrary file path; callers are responsible for sanitising paths before passing them in.
- No credentials, secrets, or network calls inside any parser.

## Logging

- The **library** (`pkg/`, `internal/`) returns `error` values and does not log. Callers should not expect any log output from the library.
- The **CLI** (`cmd/main.go`) uses `log.Fatalf` on parse/marshal errors and exits non-zero. Normal output is JSON printed to stdout.

## Coding Standards

- `gofmt` and `go vet` clean — CI will fail otherwise.
- Exported identifiers live in `pkg/`; internal logic lives in `internal/`. Do not add exported symbols to `internal/`.
- Parser packages follow the naming layout: `internal/parsers/<ecosystem>/<ecosystem>_parser.go` + `<ecosystem>_parser_test.go`.
- No global state in parsers — each concrete parser type is a stateless zero-value struct.

## Debugging Steps

1. **Run one parser against a fixture:**
   ```bash
   go run ./cmd test/resources/pom.xml
   go run ./cmd test/resources/build.gradle
   ```

2. **Verbose test output to see which test case fails:**
   ```bash
   go test -v ./internal/parsers/maven/...
   ```

3. **Location off-by-one:** Parser violated 0-based contract. Grep for `i + 1` patterns near `Line:` / `LineNum:` assignments — emit `Line: i`, not `i + 1`.

4. **Version resolves to `"latest"` unexpectedly:** check whether the version string matches a range specifier (`^`, `~`, `[`, `*`) or whether a lock file / properties file is present in the same directory as the fixture.

5. **New ecosystem not dispatched:** verify `selectManifestFile` in `manifest-file-selector.go` handles the new filename/extension and that the factory `switch` has a corresponding case.
