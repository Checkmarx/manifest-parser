# manifest-parser

A Go module for parsing package manifest files from multiple package managers. It extracts package names, versions, and their exact source locations (line and character offsets) from dependency declarations.

This module is consumed by the [AST-CLI](https://github.com/Checkmarx/ast-cli) to identify declared dependencies and power Checkmarx runtime scans.

## Supported Manifests

| Ecosystem  | File(s)                                                    |
|------------|------------------------------------------------------------|
| Maven      | `pom.xml`                                                  |
| npm        | `package.json`                                             |
| Python     | `requirements*.txt`, `packages*.txt`                       |
| Go         | `go.mod`                                                   |
| .NET       | `*.csproj`, `Directory.Packages.props`, `packages.config`  |

## Installation

```bash
go get github.com/Checkmarx/manifest-parser
```

## Usage

The entry point is the `ParsersFactory`, which selects the correct parser based on the manifest file name/extension.

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    "github.com/Checkmarx/manifest-parser/pkg/parser"
)

func main() {
    manifestFile := "path/to/pom.xml"

    p := parser.ParsersFactory(manifestFile)
    if p == nil {
        log.Fatalf("Unsupported manifest type: %s", manifestFile)
    }

    packages, err := p.Parse(manifestFile)
    if err != nil {
        log.Fatalf("Error parsing manifest: %v", err)
    }

    out, _ := json.MarshalIndent(packages, "", "  ")
    fmt.Println(string(out))
}
```

### Package Model

Each parser returns a slice of `models.Package`:

```go
type Package struct {
    PackageManager string
    PackageName    string
    Version        string
    FilePath       string
    Locations      []Location
}

type Location struct {
    Line       int
    StartIndex int
    EndIndex   int
}
```


`Locations` points to the exact position of the dependency declaration in the source manifest, which downstream tools use for inline annotations and remediation.

## CLI

A small CLI is provided under [cmd/main.go](cmd/main.go) for local testing:

```bash
go run ./cmd <manifest-file>
```

Example:

```bash
go run ./cmd test/fixtures/pom.xml
```

## Project Layout

```
cmd/                 # CLI entry point
pkg/parser/          # Public API: Parser interface, factory, models
internal/parsers/    # Per-ecosystem parser implementations
  ├── dotnet/
  ├── golang/
  ├── maven/
  ├── npm/
  └── pypi/
test/                # Integration tests and fixtures
```

## Integration with AST-CLI

The [AST-CLI](https://github.com/Checkmarx/ast-cli) imports this module to discover declared dependencies from a scanned repository, feeding them into Checkmarx runtime scanning to correlate manifest declarations with runtime package usage.

## Known Limitations

The following limitations apply when this parser is used as part of the Checkmarx One Developer Assist realtime OSS scanner (see the [official docs](https://docs.checkmarx.com/en/34965-405960-checkmarx-one-developer-assist.html)):

- **Direct dependencies only** — vulnerabilities are identified only in packages declared directly in the manifest. Transitive dependencies are not resolved or scanned.
- **Version specifiers are not evaluated** — package managers commonly allow range/wildcard specifiers (e.g., `^`, `~`, `*`, etc.). The scanner does not resolve these; when encountered, it falls back to analyzing the *latest* version of the package.
- **Python `requirements.txt` format** — only traditional, manually authored files using the `package==version` format are supported. Auto-generated files (e.g., produced by `pip freeze`, `pip-tools`, `Poetry`) are not supported.
- **Scope vs. full SCA** — the realtime OSS scanner is intentionally lighter than the full Checkmarx One SCA scanner and is therefore less comprehensive.

## Development

Run the test suite:

```bash
go test ./...
```

Build the CLI:

```bash
go build -o manifest-parser ./cmd
```

## License

See repository for license details.
