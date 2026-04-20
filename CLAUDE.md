# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Go module that parses package manifests from multiple ecosystems (Maven, npm, Python, Go, .NET) and returns each declared dependency along with the **exact line/character range** of its declaration. Consumed by [AST-CLI](https://github.com/Checkmarx/ast-cli) to correlate manifest entries with Checkmarx runtime scans — so the `Locations` field is part of the public contract, not a debugging convenience.

## Commands

```bash
go test ./...                          # run all tests
go test ./internal/parsers/maven/...   # run tests for a single parser
go test -run TestName ./path/...       # run a single test by name
go test ./... -coverprofile cover.out  # CI gate: total coverage must be >= 60%
go build -o manifest-parser ./cmd      # build CLI
go run ./cmd <manifest-file>           # run CLI against a manifest
```

Dependencies are vendored (`vendor/`). Go version is pinned via `go.mod` (1.23 / toolchain 1.24.2).

## Architecture

The module is organized around one interface and a dispatching factory:

- [pkg/parser/parser.go](pkg/parser/parser.go) — `Parser` interface (`Parse(manifestFile string) ([]models.Package, error)`).
- [pkg/parser/parser_factory.go](pkg/parser/parser_factory.go) — `ParsersFactory(manifest string)` is the **only** public entry point. It calls `selectManifestFile` and returns the right concrete parser, or `nil` for unsupported files.
- [pkg/parser/manifest-file-selector.go](pkg/parser/manifest-file-selector.go) — maps filename/extension to a `Manifest` enum. Adding a new ecosystem means editing this file, the factory, and adding a package under `internal/parsers/`.
- [pkg/parser/models/package_model.go](pkg/parser/models/package_model.go) — the `Package` / `Location` structs returned to callers. `Locations` is a slice: Maven returns one entry per line of a multi-line `<dependency>` block; most others return a single entry.

Per-ecosystem parsers live under [internal/parsers/](internal/parsers/):
- `maven/` — parses `pom.xml` with `encoding/xml`, then re-scans the raw text to locate each `<dependency>` block line by line. Resolves `${property}` vars from `<properties>` and falls back to `<dependencyManagement>` for empty/ranged versions. Only **direct** `<dependencies>` are emitted (managed-only deps are intentionally skipped to avoid duplicates — see commit `9e490aa`).
- `npm/` — parses `package.json` plus, if present as a sibling file, `package-lock.json` (v1 and v2/v3 formats). Ranged specifiers (`^`, `~`, `*`, `>`, `<`) trigger a lookup in the lockfile; `isLockVersionGreater` compares part-by-part numerically to decide whether the lockfile version satisfies the spec. Without a lock match, ranged versions resolve to `"latest"`.
- `pypi/` — line-oriented scan of `requirements*.txt` / `packages*.txt`. **Only `package==version` is supported** — `pip freeze`, Poetry, and pip-tools output are explicitly out of scope (see README "Known Limitations"). Comments (`#`) and environment markers (`;`) are stripped.
- `golang/` — uses `golang.org/x/mod/modfile` to parse `go.mod`, then uses the parser's line metadata to compute character offsets.
- `dotnet/` — three separate parsers sharing patterns: `csproj_parser.go` (`.csproj`), `directory_packages_props_parser.go` (central package management), `packages_config_parser.go` (legacy). Versions are read from either a `Version` attribute or a nested `<Version>` element; bracketed ranges become `"latest"`.

### Invariants worth preserving

- **`Location` uses 0-based line numbers** in most parsers (Maven, Go, npm, pypi use `lineNum - 1` or a 0-based counter). Downstream AST-CLI depends on this; don't "fix" it to 1-based without coordinating.
- **Unresolvable or ranged versions resolve to the literal string `"latest"`**, never an empty string. Callers branch on this value.
- **`PackageManager` strings are part of the contract**: `"mvn"`, `"npm"`, `"pypi"`, `"go"`, `"nuget"` (used by all three dotnet parsers). Don't rename them.
- Maven emits one `Location` per **non-comment line** of the `<dependency>` block (open tag, each child, close tag) so AST-CLI can annotate the whole block. Single-line `Locations` for Maven would be a regression.

## Tests & fixtures

Each parser has a `*_test.go` next to it using `testify`. Shared fixtures live in [test/resources/](test/resources/) (e.g. `pom.xml`, `package.json`, `requirements.txt`, `test_go.mod`, `Bootstrap.csproj`, `Gateway.csproj`, `packages.config`, `Directory.Packages.props`). When adding behaviors, add a fixture here rather than embedding large manifests in test source.

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) enforces a **60% total coverage floor** — adding an untested branch to an already-thin package can push the whole repo below the gate.
