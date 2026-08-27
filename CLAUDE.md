# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go module that parses package manifests from multiple ecosystems (Maven, npm, Python, Go, .NET, Gradle, SBT, Dart/Flutter, Swift, CocoaPods, Carthage, PHP/Composer, RubyGems, Bower) and returns each declared dependency along with the **exact line/character range** of its declaration. Consumed by [AST-CLI](https://github.com/Checkmarx/ast-cli) to correlate manifest entries with Checkmarx runtime scans — so the `Locations` field is part of the public contract, not a debugging convenience.

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
  ├── bower/
  ├── carthage/
  ├── cocoapods/
  ├── composer/
  ├── dart/
  ├── dotnet/
  ├── golang/
  ├── gradle/
  ├── maven/
  ├── npm/
  ├── poetry/
  ├── pypi/
  ├── rubygems/
  ├── sbt/
  ├── setuptools/
  └── swiftpm/
internal/testdata/              Parser-specific test fixtures (sbt, Python, .NET, SwiftPM, CocoaPods)
test/resources/                 Shared fixture files for all parser tests
vendor/                         Vendored dependencies
.github/workflows/ci.yml        CI pipeline (test + 60% coverage gate)
```

## Development Setup

**Prerequisites:** Go ≥ 1.23, git. No other tools required — dependencies are vendored.

```bash
git clone https://github.com/Checkmarx/manifest-parser.git
cd manifest-parser

go test ./...                             # run all tests
go test ./internal/parsers/gradle/...     # run tests for a single parser
go test -run TestName ./path/...          # run a single test by name
go test ./... -coverprofile cover.out     # CI gate: total coverage must be >= 60%
go tool cover -html cover.out             # view coverage report in browser
go build -o manifest-parser ./cmd         # build CLI
go run ./cmd test/resources/pom.xml       # run CLI against a fixture
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
- `npm/` — parses `package.json` plus, if present as a sibling file, `package-lock.json` (v1 and v2/v3 formats) or `yarn.lock` (Yarn 1 and Yarn 2+ formats) as a fallback when no `package-lock.json` exists. Ranged specifiers (`^`, `~`, `*`, `>`, `<`) trigger a lookup in the lockfile. Without a lock match, ranged versions resolve to `"latest"`. `PackageManager` = `"npm"`.
- `poetry/` — parses `pyproject.toml`: Poetry's own `[tool.poetry.dependencies]` / `[tool.poetry.dev-dependencies]` / `[tool.poetry.group.*.dependencies]` (key-value format), plus the PEP 621 `[project]` / `[project.optional-dependencies]` and PEP 735 `[dependency-groups]` array formats. **uv projects have no dedicated parser** — uv uses standard PEP 621 `pyproject.toml`, so uv dependencies flow through this same parser. Supports exact versions, ranges (`^`, `~`, `>=`, `<=`), wildcards (`*`, `1.2.*`, `==1.*`), PEP 440 arbitrary equality (`===`), bare unversioned names (`"requests"`), PEP 508 extras (`"requests[security]"` reports `PackageName` as `"requests"`, not `"requests[security]"`), inline tables, optional packages, dependency groups, PEP 735 `{include-group = "..."}` references (skipped, not emitted as a package — the referenced group's own array is parsed independently), and markers — including marker-only requirements with no version specifier at all (`"tomli; python_version < '3.11'"`), where the marker's own comparison operators are correctly excluded from the package-version search. Resolves exact versions from a sibling `poetry.lock` and/or `uv.lock`; if both are present, `poetry.lock` wins per-package and `uv.lock` fills in anything it doesn't cover. Lock lookups are PEP 503 name-normalized (case-insensitive, `-`/`_`/`.` equivalent) and also fire for bare unversioned dependencies (the internal `"latest"` placeholder is treated as unresolved, not as a literal concrete version). Ranged/wildcard/unversioned dependencies without a lock match resolve to `"latest"`. `PackageManager` = `"pypi"` (Poetry and uv packages are both PyPI packages).
- `pypi/` — line-oriented scan of `requirements*.txt`, `requirement*.txt`, `constraints*.txt`, and `packages*.txt`. Supports six Python dependency formats: pip, pip-freeze, pip-compile, pip-tools, uv export, and Poetry export. Features: line continuations (`\`), `--hash=` stripping, pip CLI option skipping (`-i`, `-r`, `-c`, `-e`, etc.), VCS requirements (`git+`, `hg+`, `svn+`, `bzr+` with `#egg=`), URL requirements (PEP 508 `pkg @ URL`), `===` arbitrary equality, and environment markers (`;`). `PackageManager` = `"pypi"`.
- `setuptools/` — two parsers for Python packaging manifests: `setup_cfg_parser.go` (`setup.cfg` INI format) and `setup_py_parser.go` (`setup.py` script). Both support `install_requires`, `setup_requires`, `tests_require`, and `extras_require`. Duplicate packages across sections are stored as separate entries with distinct line numbers. `PackageManager` = `"pypi"`.
- `golang/` — uses `golang.org/x/mod/modfile` to parse `go.mod`, then uses the parser's line metadata to compute character offsets. `PackageManager` = `"go"`.
- `dotnet/` — three parsers: `csproj_parser.go` (`.csproj`), `directory_packages_props_parser.go` (central package management), `packages_config_parser.go` (legacy). Bracketed version ranges become `"latest"`. `PackageManager` = `"nuget"` for all three.
- `dart/` — parses `pubspec.yaml` (Dart/Flutter manifest) plus, if present as a sibling file, `pubspec.lock` (YAML lock). Mines the `dependencies`, `dev_dependencies`, and `dependency_overrides` sections via line-oriented scanning. The dependency-entry indent is detected per section (from the first entry), so any consistent indent width — not just 2 spaces — is parsed. Supports pub's constraint operators (`^`, `>=`, `<=`, `>`, `<`, compound ranges, exact, `any`) and nested source maps (`git`, `path`, `hosted`, `sdk`); the npm-style `~` and `*` are handled defensively though they are not valid pub syntax. SDK-sourced deps (`flutter`, `flutter_test`, anything with an `sdk:` key) are skipped since they are not pub.dev packages. Duplicate declarations across sections each emit a separate entry with their own location. Ranged/source-map versions resolve from the lock, falling back to `"latest"`. `pubspec.lock` is a resolver-only sibling file, like every other lock file in this module — it is **not** reachable as a standalone manifest through `ParsersFactory` (`selectManifestFile("pubspec.lock")` returns `-1`; see `TestManifestFileSelector_PubspecLockNotStandalone`), even though `DartParser.Parse` has an internal case for it. `PackageManager` = `"pub"`.
- `sbt/` — parses any `.sbt` file (`build.sbt`, `plugins.sbt`, `dependencies.sbt`, etc.) using line-oriented scanning. Supports `val`/`lazy val`/`def` variable declarations, all SBT dependency operators (`%`, `%%`, `%%%`), `Seq(...)` blocks, `addSbtPlugin(...)` syntax, dependency modifiers (`exclude`, `excludeAll`, `intransitive`, `withSources`, `classifier`, `cross`), block and inline comments, scope annotations, and duplicate detection. `PackageManager` = `"sbt"`.
- `swiftpm/` — dispatches on filename: `Package.swift` (Swift DSL, line-oriented scan that accumulates multi-line `.package(...)` calls until parens balance) or `Package.resolved` (JSON lock file, one entry per pin). Package name is taken from `url:`/`id:` args (or `identity`/`package` in the lock); purely `path:`-based local packages are skipped since they have no registry identity. Version comes from `.exact`/`from`/`.upToNextMajor`/`.upToNextMinor`, overridden by a sibling `Package.resolved` when present; unresolvable specs become `"latest"`. `Package.resolved` is a resolver-only sibling file, like every other lock file in this module — it is **not** reachable as a standalone manifest through `ParsersFactory` (`selectManifestFile("Package.resolved")` returns `-1`; see `TestManifestFileSelector_PackageResolvedNotStandalone`), even though `SwiftPmParser.Parse` has an internal case for it. `PackageManager` = `"swift"`.
- `cocoapods/` — dispatches on filename across four sub-parsers: `podfile.go` (`Podfile`, Ruby DSL `pod 'Name', 'version'` lines), `podfile_lock.go` (`Podfile.lock`, pseudo-YAML, mines only the 2-space-indented `PODS:` section to avoid capturing transitive sub-dependency lines), `podspec.go` (`*.podspec`, Ruby DSL `<receiver>.dependency 'Name', 'version'`), and `podspec_json.go` (`*.podspec.json`, trunk-published JSON spec). `target 'Name' do ... end` blocks are not tracked — each `pod` line is emitted independently. Git/branch/tag/commit/path references and range operators (`~>`, `>=`, etc.) resolve to `"latest"`; a sibling `Podfile.lock` overrides with the exact resolved version. `PackageManager` = `"cocoapods"`.
- `carthage/` — parses `Cartfile` / `Cartfile.private` (production and private/test deps) plus `Cartfile.resolved` as a sibling lock file. Supports `github "org/repo"`, `git "url"`, and `binary "url"` origin lines; package name is the `org/repo` slug for GitHub origins or the last path segment for git/binary origins. Only exact (`== "1.2.3"`) or bare quoted semver-like specs resolve to a concrete version — everything else (including no spec) resolves to `"latest"` unless overridden by the lock file. `PackageManager` = `"carthage"`.
- `composer/` — parses `composer.json`'s `require` and `require-dev` maps, plus a sibling `composer.lock` for resolving ranged versions (`^`, `~`, `*`, comparison operators, `||`, hyphen ranges). Virtual/platform packages (`php`, `php-*`, `ext-*`, `lib-*`, `composer-*`, `hhvm*`) are skipped — real Packagist packages always contain a `/` in their name, so that check is used to never misclassify a real `vendor/package`. `PackageManager` = `"packagist"`.
- `rubygems/` — line-oriented regex scan of `Gemfile` for `gem 'name', 'version'` declarations, with version resolution from a sibling `Gemfile.lock`'s `GEM`/`specs:` section (4-space-indented `name (version)` entries only — `GIT`/`PATH` sections are not mined). Ranged specifiers (`~>`, `>=`, `<=`, `>`, `<`, `!=`, `*`) without a lock match resolve to `"latest"`. `PackageManager` = `"rubygems"`.
- `bower/` — parses `bower.json`'s `dependencies` and `devDependencies` maps using the same brace-depth-tracking JSON positional scan as npm. Bower has no lock file, so any range/wildcard (`^`, `~`, `*`, `x`-ranges), URL, git, local-path, or GitHub-shorthand (`user/repo`) specifier resolves to `"latest"`. **`PackageManager` = `"npm"`** — intentionally reused rather than a new `"bower"` string, since Bower packages are JS-ecosystem packages; do not treat this as a bug.

## Project Rules (Invariants)

- **`Location.Line` MUST be 0-based for ALL parsers.** When iterating `for i, line := range lines`, emit `Line: i` — never `i + 1`. Editors display 1-based line numbers; downstream consumers add `+1` for display. If parser output matches the editor's line number, it's off-by-one.
- **`Location.StartIndex` / `EndIndex` are 0-based byte offsets** from the start of the line. They are byte offsets, not rune/character offsets — relevant for non-ASCII manifests.
- **Unresolvable or ranged versions resolve to the literal string `"latest"`**, never an empty string. Callers branch on this value.
- **`PackageManager` strings are part of the contract**: `"gradle"`, `"mvn"`, `"npm"`, `"pypi"`, `"go"`, `"nuget"`, `"sbt"`, `"pub"`, `"swift"`, `"cocoapods"`, `"carthage"`, `"packagist"`, `"rubygems"`. Don't rename them.
- Maven emits one `Location` per **non-comment line** of the `<dependency>` block (open tag, each child element, close tag). Single-line `Locations` for Maven would be a regression.
- Do not add `ParsersFactory` overloads or alternative entry points without coordinating with AST-CLI.
- **Do not modify or rename existing `PackageManager` strings**. AST-CLI and Checkmarx One SCA branch on these values — a silent rename breaks downstream parsing with no compile-time error. If a rename is genuinely required, stop and confirm with the user.
- All Python parsers (`pypi/`, `poetry/`, `setuptools/`) return `PackageManager` = `"pypi"` because all Python packages ultimately live on PyPI regardless of the tool that declared them. Do not introduce separate strings like `"poetry"` or `"setuptools"`.
- `bower/` deliberately returns `PackageManager` = `"npm"`, not `"bower"` — this is intentional (Bower packages are JS-ecosystem packages), not an oversight. Don't "fix" it without confirming with the user first.

## Testing Strategy

Each parser has a `*_test.go` co-located with it using `testify`. Fixtures are split across two locations:

**Shared fixtures** in [test/resources/](test/resources/):
```
test/resources/
├── build.gradle              Groovy DSL
├── build.gradle.kts          Kotlin DSL
├── gradle/libs.versions.toml Version catalog (80+ entries)
├── gradle.properties         Centralized Gradle properties
├── pom.xml                   Maven
├── package.json              npm
├── yarn-v1/                  npm + Yarn 1 lockfile pair
├── yarn-v2/                  npm + Yarn 2+ lockfile pair
├── test_go.mod               Go modules
├── Bootstrap.csproj          .NET csproj
├── Gateway.csproj            .NET csproj (variant)
├── Directory.Packages.props  .NET centralized packages
├── packages.config           .NET legacy NuGet
├── requirements.txt          Python pip (basic format)
├── pubspec.yaml              Dart/Flutter pub manifest
├── pubspec.lock              Dart/Flutter pub lock file
├── Package.swift             SwiftPM manifest
├── Package.resolved          SwiftPM lock file
├── Podfile                   CocoaPods app manifest
├── Podfile.lock              CocoaPods lock file
├── Cartfile                  Carthage production deps
├── Cartfile.private          Carthage private/test deps
├── Cartfile.resolved         Carthage lock file
├── composer.json             Composer (PHP) manifest
├── composer.lock             Composer lock file
├── Gemfile                   RubyGems manifest
├── Gemfile.lock              RubyGems lock file
└── bower.json                Bower manifest
```

**Parser-specific fixtures** in [internal/testdata/](internal/testdata/):
```
internal/testdata/
├── build.sbt                            SBT build file (Log4Shell, Struts2, etc.)
├── plugins.sbt                           SBT plugin dependencies
├── pyproject.toml                        Poetry project configuration (requests, flask, pytest, numpy, pandas)
├── setup.cfg                             Setuptools INI format (requests, flask, six, pytest, black)
├── setup.py                              Setuptools Python script (same deps as setup.cfg)
├── ast-visual-studio-extension.csproj    .NET multi-package csproj
├── Package@swift-6.0.swift               SwiftPM multi-toolchain manifest variant
├── Package.resolved.v1.json              SwiftPM v1 lock file format
├── TestPod.podspec                       CocoaPods Ruby DSL pod author spec
└── TestPod.podspec.json                  CocoaPods JSON (trunk-published) pod author spec
```

**PyPI-format fixtures** in [internal/parsers/pypi/testdata/](internal/parsers/pypi/testdata/):
```
internal/parsers/pypi/testdata/
├── requirements-pip-freeze.txt    pip freeze output (exact pinned versions)
├── requirements-pip-compile.txt   pip-compile output with via comments
└── requirements-uv-export.txt     uv export with --hash options and line continuations
```

When adding behaviours, add a fixture here rather than embedding large manifests in test source.

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) enforces a **60% total coverage floor** — adding an untested branch to an already-thin package can push the whole repo below the gate. View coverage locally with `go tool cover -html cover.out`.

Expected pattern for a new parser: fixture file under `test/resources/` or `internal/testdata/` + `<ecosystem>_parser_test.go` co-located with the parser, using `testify` assertions on `PackageName`, `Version`, `PackageManager`, and `Locations`.

## Known Issues / Limitations

- **pypi**: VCS requirements (`git+`, `hg+`, `svn+`, `bzr+`) require an `#egg=<name>` fragment to extract the package name; VCS URLs without `#egg=` are skipped. URL requirements must use PEP 508 `pkg @ URL` syntax with the package name before `@`.
- **poetry**: Multi-line dependency tables spanning more than one line (e.g., `{git = "...", rev = "..."}` across lines) are not fully parsed — the dependency is skipped. Single-line inline tables are supported.
- **uv**: `uv.lock` is deliberately a version-resolver only (see the `poetry/` bullet above) — it is never scanned as a standalone manifest, matching every other lock file in this module (`package-lock.json`, `poetry.lock`, `Podfile.lock`, etc.) and Checkmarx SCA's own manifest-vs-lock model. This was a deliberate design decision, not an oversight: making one lock file standalone while all others remain resolver-only would be an inconsistency with no principled stopping point. The legacy pre-PEP-735 `[tool.uv] dev-dependencies = [...]` format (superseded by `[dependency-groups]`) is not parsed — a documented gap, not a bug. **Known unfixed gap:** uv workspaces keep one `uv.lock` at the workspace root while member packages each have their own `pyproject.toml` in a subdirectory (e.g. `packages/foo/pyproject.toml`); `loadLockVersions` only looks in the *same* directory as the manifest being parsed, so a workspace member's ranged/unversioned dependencies resolve to `"latest"` even though the root `uv.lock` has the exact pin. Fixing this requires walking up parent directories, which needs a bounded/safe search strategy — not yet implemented.
- **npm**: Ranged version specifiers (`^`, `~`, `*`, `>`, `<`) without a matching `package-lock.json` entry resolve to `"latest"` rather than the actual installed version.
- **Maven**: Managed-only deps (present in `<dependencyManagement>` but not in `<dependencies>`) are not emitted, to avoid duplicating entries already declared in a BOM consumer.
- **dotnet**: Bracketed version ranges (e.g., `[1.0,2.0)`) become `"latest"`.
- **dart**: SDK-sourced dependencies (`sdk: flutter`, etc.) are intentionally skipped — they are not pub.dev packages. Ranged constraints and source-map deps (`git`/`path`/`hosted`) without a matching `pubspec.lock` entry resolve to `"latest"`. Flow/inline-map dependency declarations (`dependencies: {http: ^0.13.5}`) are not parsed — block style only.
- **sbt**: Version variables using object member access (e.g., `Versions.log4j`) are not resolved — only simple `val`/`lazy val` string assignments are captured.
- **swiftpm**: Purely `path:`-based local package dependencies (no `url:`/`id:`) are skipped — they have no registry identity to report. `Package@swift-X.Y.swift` (Swift-version-tooled manifests) is parsed with the same logic as `Package.swift`.
- **cocoapods**: `Podfile` `target 'Name' do ... end` blocks are not tracked — every `pod` line is emitted independently of which target it's nested in. Any `:git`/`:branch`/`:tag`/`:commit`/`:path` reference or range operator (`~>`, `>=`, etc.) resolves to `"latest"` unless a sibling `Podfile.lock` has an exact match.
- **carthage**: Only bare quoted semver-like specs or `== "x.y.z"` resolve to a concrete version directly from the manifest; everything else (no spec, branch, commit) resolves to `"latest"` unless overridden by `Cartfile.resolved`.
- **composer**: Virtual/platform packages (`php`, `php-*`, `ext-*`, `lib-*`, `composer-*`, `hhvm*`) are intentionally skipped since they aren't real Packagist packages.
- **rubygems**: Version resolution only reads the `GEM`/`specs:` section of `Gemfile.lock` — gems sourced from `GIT:` or `PATH:` sections in the lock are not resolved and fall back to `"latest"`.
- **bower**: No lock file format exists for Bower, so any range, wildcard, URL, git, local-path, or GitHub-shorthand (`user/repo`) specifier resolves to `"latest"`.
- **All parsers**: Direct dependencies only — transitive dependencies are not resolved or scanned.

## External Integrations

- **AST-CLI** ([Checkmarx/ast-cli](https://github.com/Checkmarx/ast-cli)) — primary consumer. Imports this module as a Go library. The fields `Locations`, `PackageManager`, `PackageName`, and `Version` on the `Package` struct are load-bearing: AST-CLI uses them to annotate scan results and drive remediation UI. Note: AST-CLI maps `"gradle"` and `"sbt"` to `"mvn"` when sending to the Checkmarx scanner API, since both build tools use Maven Central as their registry.
- **Checkmarx One SCA** — downstream scan engine that receives the parsed dependency list.

## Deployment

N/A — this is a Go library consumed via `go get github.com/Checkmarx/manifest-parser`. It is not deployed as a service. The CLI (`cmd/`) is a local testing convenience, not a production artifact.

## Performance Considerations

- Maven re-scans the raw XML bytes after `encoding/xml` parsing (two passes). Large `pom.xml` files are loaded fully into memory; there is no streaming.
- Gradle version catalog parsing reads `libs.versions.toml` once, separately from the build file. Large catalogs (80+ entries) are fine; pathologically large files are not size-bounded.
- pypi parser preprocesses all lines first to join continuations before parsing — the full file is held in memory.
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
- Exception: `setuptools/` parsers use `log.Printf` for debug/warning output during development. This should be treated as temporary and removed before production release.

## Coding Standards

- `gofmt` and `go vet` clean — CI will fail otherwise.
- Exported identifiers live in `pkg/`; internal logic lives in `internal/`. Do not add exported symbols to `internal/`.
- Parser packages follow the naming layout: `internal/parsers/<ecosystem>/<ecosystem>_parser.go` + `<ecosystem>_parser_test.go`.
- No global state in parsers — each concrete parser type is a stateless zero-value struct.
- When splitting file content into lines, always strip `\r` to handle CRLF files on Windows: `strings.TrimRight(line, "\r")`. Failing to do so causes `len(line)` to return one extra byte, producing off-by-one `EndIndex` values.

## Debugging Steps

1. **Run one parser against a fixture:**
   ```bash
   go run ./cmd test/resources/pom.xml
   go run ./cmd test/resources/build.gradle
   go run ./cmd internal/testdata/build.sbt
   go run ./cmd internal/testdata/pyproject.toml
   go run ./cmd internal/testdata/setup.cfg
   go run ./cmd test/resources/Podfile
   go run ./cmd test/resources/Gemfile
   go run ./cmd test/resources/composer.json
   ```

2. **Verbose test output to see which test case fails:**
   ```bash
   go test -v ./internal/parsers/maven/...
   go test -v ./internal/parsers/sbt/...
   go test -v ./internal/parsers/poetry/...
   ```

3. **Location off-by-one:** Parser violated 0-based contract. Grep for `i + 1` patterns near `Line:` / `LineNum:` assignments — emit `Line: i`, not `i + 1`.

4. **EndIndex off-by-one on Windows:** File has CRLF line endings and the parser uses `len(line)` without stripping `\r`. Fix: add `strings.TrimRight(line, "\r")` after `strings.Split(content, "\n")`.

5. **Version resolves to `"latest"` unexpectedly:** check whether the version string matches a range specifier (`^`, `~`, `[`, `*`) or whether a lock file / properties file is present in the same directory as the fixture.

6. **New ecosystem not dispatched:** verify `selectManifestFile` in `manifest-file-selector.go` handles the new filename/extension and that the factory `switch` has a corresponding case.
