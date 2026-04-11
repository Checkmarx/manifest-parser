# Manifest Parser

A production-grade Go library for parsing dependency manifests across multiple package managers. Extracts package dependencies from build files and dependency declarations in a standardized format for security scanning, SBOM generation, and dependency analysis.

## 🎯 Purpose

This parser extracts software dependencies from project manifest files and provides:
- **Standardized Package Output** - Consistent JSON format across all package managers
- **Version Tracking** - Precise version information for vulnerability scanning
- **Location Tracking** - File path and line numbers for each dependency
- **Security Scanning** - Integration with SCA (Software Composition Analysis) tools
- **SBOM Generation** - Software Bill of Materials (cyclonedx, spdx) support

## 📦 Supported Package Managers

| Manager | Format | Status | Features |
|---------|--------|--------|----------|
| **Gradle** | `build.gradle`, `build.gradle.kts`, `libs.versions.toml` | ✅ Production | Latest DSL + catalogs + direct TOML parsing |
| **Maven** | `pom.xml` | ✅ Production | Properties, BOMs, ranges |
| **npm/Node.js** | `package.json` | ✅ Production | Dependencies, dev, peer, optional |
| **Go** | `go.mod` | ✅ Production | Direct imports, indirect |
| **.NET** | `.csproj`, `Directory.Packages.props`, `packages.config` | ✅ Production | Multi-format support |
| **Python** | `requirements.txt` | ✅ Production | Pip format with ranges |

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/Checkmarx/manifest-parser
```

### Usage

```go
package main

import (
    "fmt"
    "github.com/Checkmarx/manifest-parser/pkg/parser"
)

func main() {
    // Create parser for manifest file
    p := parser.ParsersFactory("path/to/package.json")
    if p == nil {
        fmt.Println("Unsupported manifest type")
        return
    }

    // Parse dependencies
    packages, err := p.Parse("path/to/package.json")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // Process results
    for _, pkg := range packages {
        fmt.Printf("%s:%s@%s\n", pkg.PackageManager, pkg.PackageName, pkg.Version)
    }
}
```

### Command Line

```bash
# Parse any supported manifest
go run cmd/main.go path/to/manifest

# Examples
go run cmd/main.go project/pom.xml
go run cmd/main.go project/package.json
go run cmd/main.go project/build.gradle
go run cmd/main.go project/go.mod
```

---

## 📋 Detailed Parser Documentation

### 1. Gradle Parser

**Files:** `build.gradle`, `build.gradle.kts`, `gradle/libs.versions.toml`

#### Features

✅ **Groovy DSL** - Traditional Android/Java Gradle syntax
✅ **Kotlin DSL** - Modern type-safe Gradle syntax
✅ **gradle.properties** - Centralized property management
✅ **Version Catalog** - `gradle/libs.versions.toml` (Gradle 7.0+)
✅ **BOM/Platform** - Dependency Bill of Materials imports
✅ **Multi-Module** - Subproject and module-specific configurations
✅ **19 Configurations** - implementation, api, testImplementation, debugImplementation, ksp, etc.

#### Dependency Declaration Support

```gradle
// String notation
implementation 'org.springframework:spring-core:5.3.20'

// Kotlin DSL
implementation("org.springframework:spring-core:5.3.20")

// Map notation
implementation group: 'org.springframework', name: 'spring-core', version: '5.3.20'

// Platform/BOM
implementation platform('org.springframework.boot:spring-boot-dependencies:2.7.0')

// Version Catalog
implementation(libs.spring.core)
```

#### Variable Resolution

```gradle
// gradle.properties
springVersion=5.3.20

// build.gradle
implementation "org.springframework:spring-core:${springVersion}"

// ext blocks
ext {
    log4jVersion = '2.17.1'
}
dependencies {
    implementation "org.apache.logging.log4j:log4j-core:$log4jVersion"
}
```

#### Supported Configurations

| Type | Purpose |
|------|---------|
| `implementation` | Runtime + compile dependencies |
| `api` | Public API (exported to consumers) |
| `compileOnly` | Compile-time only (e.g., annotations) |
| `runtimeOnly` | Runtime-only (excluded from compile) |
| `testImplementation` | Test-only dependencies |
| `debugImplementation` | Debug build variant |
| `releaseImplementation` | Release build variant |
| `annotationProcessor` | Annotation code generation |
| `ksp` / `kapt` | Kotlin/Java code generation |
| `classpath` | Buildscript dependencies |
| Plus 9 more variants for testing, fixtures, lint checks |

#### Example: Multi-Module Project

```kotlin
// build.gradle.kts
subprojects {
    apply(plugin = "java")
    
    dependencies {
        implementation("org.springframework.boot:spring-boot-starter-web")
    }
}

project(":api-module") {
    dependencies {
        implementation(project(":core"))
        implementation("org.springframework.security:spring-security-core:5.7.1")
    }
}
```

#### Version Catalog Support

**Direct Parsing:** You can now parse `libs.versions.toml` directly!

```bash
# Parse version catalog directly
go run cmd/main.go gradle/libs.versions.toml
```

**Catalog Format:**

```toml
# gradle/libs.versions.toml
[versions]
spring-version = "5.3.20"

[libraries]
spring-core = { module = "org.springframework:spring-core", version.ref = "spring-version" }

[bundles]
spring = ["spring-core", "spring-context"]
```

**Automatic Discovery:** When parsing `build.gradle` or `build.gradle.kts`, the parser automatically discovers and parses `gradle/libs.versions.toml` in the same directory.

#### Parser Capabilities

**Build File Parsing:**
- ✅ Parses Groovy and Kotlin DSL
- ✅ Resolves variables from gradle.properties
- ✅ Discovers and parses version catalogs
- ✅ Unwraps platform()/enforcedPlatform() BOMs
- ✅ Walks up directory tree for parent properties
- ✅ Filters out project references (multi-module)
- ✅ Skips file references (local JARs)
- ✅ Handles multi-line declarations
- ✅ Parses conditional if blocks

**Version Catalog Parsing:**
- ✅ Direct parsing of `libs.versions.toml` files
- ✅ Extracts all 80+ library definitions
- ✅ Resolves version references
- ✅ Supports all catalog formats (simple, module, key-value)
- ✅ Works standalone or auto-discovered by build files

**General:**
- ❌ Does not evaluate dynamic Gradle code

#### Test Resources

```
test/resources/
├── build.gradle              - Groovy DSL with subprojects
├── build.gradle.kts          - Kotlin DSL with 5 modules
├── gradle.properties         - Centralized properties
└── gradle/libs.versions.toml - 80+ catalog entries
```

**Test Coverage:** 16 passing tests including platform dependencies, version catalogs, extended configurations, parent property inheritance

---

### 2. Maven Parser

**File:** `pom.xml`

#### Features

✅ **Dependency Management** - BOM imports and managed versions
✅ **Multi-Module** - Parent/child POM relationships
✅ **Properties** - Variable substitution with `${property}`
✅ **Version Ranges** - `[1.0,2.0)` notation handling
✅ **Scopes** - compile, runtime, test, provided, optional, system
✅ **Location Tracking** - Exact line numbers in POM files

#### Dependency Declaration Support

```xml
<!-- Basic dependency -->
<dependency>
    <groupId>org.springframework</groupId>
    <artifactId>spring-core</artifactId>
    <version>5.3.20</version>
</dependency>

<!-- With scope -->
<dependency>
    <groupId>junit</groupId>
    <artifactId>junit</artifactId>
    <version>4.13.2</version>
    <scope>test</scope>
</dependency>

<!-- Using property -->
<dependency>
    <groupId>org.springframework</groupId>
    <artifactId>spring-core</artifactId>
    <version>${spring.version}</version>
</dependency>

<!-- Version range -->
<dependency>
    <groupId>com.example</groupId>
    <artifactId>library</artifactId>
    <version>[1.0,2.0)</version>
</dependency>

<!-- BOM import -->
<dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-dependencies</artifactId>
            <version>2.7.0</version>
            <type>pom</type>
            <scope>import</scope>
        </dependency>
    </dependencies>
</dependencyManagement>
```

#### Property Resolution

```xml
<properties>
    <spring.version>5.3.20</spring.version>
</properties>

<!-- Resolved to 5.3.20 -->
<version>${spring.version}</version>
```

#### Dependency Scopes

| Scope | Purpose |
|-------|---------|
| `compile` | Runtime + compile (default) |
| `test` | Test-only dependencies |
| `runtime` | Runtime-only |
| `provided` | Compile-only, provided at runtime |
| `optional` | Included optionally |
| `system` | Local filesystem JAR |

#### Parser Capabilities

- ✅ Parses POM XML structure
- ✅ Resolves properties and version ranges
- ✅ Handles BOM imports and managed dependencies
- ✅ Tracks multi-line elements
- ✅ Extracts scope information
- ✅ Locates exact line numbers
- ✅ Supports parent POM references

#### Example: Multi-Module Project

```xml
<!-- parent-pom.xml -->
<groupId>com.example</groupId>
<artifactId>parent</artifactId>
<version>1.0.0</version>
<packaging>pom</packaging>

<modules>
    <module>core</module>
    <module>api</module>
</modules>

<!-- child-pom: core/pom.xml -->
<parent>
    <groupId>com.example</groupId>
    <artifactId>parent</artifactId>
    <version>1.0.0</version>
</parent>

<artifactId>core</artifactId>

<dependencies>
    <dependency>
        <groupId>org.springframework</groupId>
        <artifactId>spring-core</artifactId>
        <version>${spring.version}</version>
    </dependency>
</dependencies>
```

---

### 3. NPM/Node.js Parser

**File:** `package.json`

#### Features

✅ **Dependency Types** - dependencies, devDependencies, peerDependencies, optionalDependencies
✅ **Version Resolution** - Resolves ranges using package-lock.json
✅ **Exact Versions** - Extracts actual installed versions from lock files
✅ **Range Handling** - `^1.0.0`, `~1.0.0`, `*`, ranges

#### Dependency Declaration Support

```json
{
  "dependencies": {
    "express": "4.18.2",
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "jest": "~29.0.0",
    "webpack": "*"
  },
  "peerDependencies": {
    "react": "^18.0.0"
  },
  "optionalDependencies": {
    "fsevents": "2.3.2"
  }
}
```

#### Version Specifiers

| Format | Meaning |
|--------|---------|
| `1.2.3` | Exact version |
| `^1.2.3` | Compatible with 1.2.3 (up to 2.0.0) |
| `~1.2.3` | Approximately 1.2.3 (up to 1.3.0) |
| `>=1.2.3` | Greater than or equal |
| `1.2.x` | Patch-level ranges |
| `*` | Any version |

#### Dependency Types

| Type | Purpose |
|------|---------|
| `dependencies` | Production dependencies |
| `devDependencies` | Development-only (testing, bundling) |
| `peerDependencies` | Consumer-provided dependencies |
| `optionalDependencies` | Optional packages |

#### Parser Capabilities

- ✅ Parses package.json JSON
- ✅ Resolves version ranges using package-lock.json
- ✅ Extracts all 4 dependency types
- ✅ Handles multiple version specifiers
- ✅ Provides exact installed versions

#### Example: Large Project

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "dependencies": {
    "react": "18.2.0",
    "react-dom": "18.2.0",
    "axios": "^1.4.0"
  },
  "devDependencies": {
    "@babel/core": "^7.22.0",
    "webpack": "^5.88.0",
    "jest": "~29.0.0"
  }
}
```

---

### 4. Go Modules Parser

**File:** `go.mod`

#### Features

✅ **Module Dependencies** - Direct and indirect imports
✅ **Version Pinning** - Exact semver versions
✅ **Replace Directives** - Local and remote replacements
✅ **Exclude Directives** - Version exclusions
✅ **Go Version** - Minimum Go version requirement

#### Dependency Declaration Support

```go
module github.com/example/project

go 1.19

require (
    github.com/gorilla/mux v1.8.0
    github.com/google/uuid v1.3.0
)

require (
    github.com/stretchr/testify v1.8.4 // indirect
)

replace (
    github.com/old/module => github.com/new/module v1.2.3
    github.com/local/module => ./local/path
)

exclude (
    github.com/bad/module v1.0.0
)
```

#### Dependency Status

| Type | Purpose |
|------|---------|
| `require` | Direct dependencies |
| `require (indirect)` | Transitive dependencies |
| `replace` | Local/remote replacements |
| `exclude` | Excluded versions |

#### Parser Capabilities

- ✅ Parses go.mod file format
- ✅ Extracts direct and indirect imports
- ✅ Handles replace and exclude directives
- ✅ Tracks minimum Go version
- ✅ Provides exact line numbers

#### Example: Complex Project

```go
module github.com/checkmarx/scanner

go 1.20

require (
    github.com/spf13/cobra v1.7.0
    github.com/sirupsen/logrus v1.9.3
)

require (
    github.com/inconshreveable/log15 v2.3.2 // indirect
    golang.org/x/sys v0.10.0 // indirect
)

replace github.com/local/package => ../local/package

exclude golang.org/x/text v0.3.0
```

---

### 5. .NET / C# Parser

**Files:** `.csproj`, `Directory.Packages.props`, `packages.config`

#### Features

✅ **Project References** - `.csproj` PackageReference elements
✅ **Centralized Management** - `Directory.Packages.props` for monorepos
✅ **Legacy Format** - `packages.config` (NuGet v2)
✅ **Target Frameworks** - Framework-specific dependencies
✅ **Metadata** - Version, Include, Exclude attributes

#### Dependency Declaration Support

##### `.csproj` Format (Modern)

```xml
<ItemGroup>
    <PackageReference Include="Microsoft.AspNetCore.App" Version="2.2.0" />
    <PackageReference Include="Newtonsoft.Json" Version="13.0.2" />
    <PackageReference Include="System.Net.Http" Version="4.3.4" Condition="'$(TargetFramework)' == 'net472'" />
</ItemGroup>
```

##### `Directory.Packages.props` (Centralized)

```xml
<PropertyGroup>
    <ManagePackageVersionsCentrally>true</ManagePackageVersionsCentrally>
</PropertyGroup>

<ItemGroup>
    <PackageVersion Include="Microsoft.AspNetCore.App" Version="2.2.0" />
    <PackageVersion Include="Newtonsoft.Json" Version="13.0.2" />
</ItemGroup>
```

##### `packages.config` (Legacy NuGet)

```xml
<?xml version="1.0" encoding="utf-8"?>
<packages>
    <package id="Newtonsoft.Json" version="13.0.2" targetFramework="net48" />
    <package id="Microsoft.AspNetCore" version="2.2.0" targetFramework="net472" />
</packages>
```

#### Package Metadata

| Attribute | Purpose |
|-----------|---------|
| `Include` / `id` | Package name |
| `Version` | Semantic version |
| `TargetFramework` | Framework specificity |
| `Condition` | Conditional inclusion |
| `Exclude` | Excluded frameworks |

#### Parser Capabilities

- ✅ Parses `.csproj` XML structure
- ✅ Extracts `Directory.Packages.props` central versions
- ✅ Handles legacy `packages.config` format
- ✅ Respects framework-specific conditions
- ✅ Tracks line numbers and locations

#### Example: Multi-Framework Project

```xml
<!-- .csproj -->
<Project Sdk="Microsoft.NET.Sdk">
    <PropertyGroup>
        <TargetFrameworks>net6.0;net8.0;net472</TargetFrameworks>
    </PropertyGroup>

    <ItemGroup>
        <PackageReference Include="Serilog" Version="2.12.0" />
        <PackageReference Include="System.Net.Http" Version="4.3.4" Condition="'$(TargetFramework)' == 'net472'" />
    </ItemGroup>
</Project>
```

---

### 6. Python / Pip Parser

**File:** `requirements.txt`

#### Features

✅ **Pip Format** - Standard Python dependency format
✅ **Version Specifiers** - `==`, `>=`, `<=`, `~=`, ranges
✅ **Comments & Empty Lines** - Properly ignored
✅ **Environment Markers** - OS/Python version conditions
✅ **Git References** - VCS dependencies

#### Dependency Declaration Support

```txt
# Production dependencies
Django==4.2.0
djangorestframework>=3.14.0,<4.0
requests~=2.31.0

# Dev dependencies
pytest>=7.0.0
black==23.0.0

# Git references
git+https://github.com/example/repo.git@main#egg=mypackage

# With environment markers
pywin32>=300; sys_platform == 'win32'
```

#### Version Specifiers

| Specifier | Meaning |
|-----------|---------|
| `==1.4.2` | Exact version |
| `>=1.4.2` | Greater than or equal |
| `<=1.4.2` | Less than or equal |
| `!=1.4.2` | Not equal |
| `~=1.4.2` | Compatible release (1.4.x) |
| `*` | Any version |

#### Environment Markers

```txt
# Platform-specific
pywin32>=300; sys_platform == 'win32'

# Python version specific
dataclasses; python_version < '3.7'

# Complex conditions
numpy>=1.20; python_version >= '3.8' and sys_platform != 'win32'
```

#### Parser Capabilities

- ✅ Parses pip requirements format
- ✅ Extracts package names and versions
- ✅ Handles version specifier ranges
- ✅ Recognizes environment markers
- ✅ Ignores comments and blank lines

#### Example: Complete Project

```txt
# Python 3.8+
Python>=3.8

# Web Framework
Flask==2.3.0
Flask-SQLAlchemy>=3.0.0,<4.0

# Database
psycopg2-binary~=2.9.0
SQLAlchemy>=2.0.0

# Testing
pytest>=7.0.0
pytest-cov>=4.0.0

# Development
black==23.0.0
flake8>=6.0.0

# OS-specific
pywin32>=300; sys_platform == 'win32'
```

---

## 📊 Output Format

All parsers return a standardized `Package` structure:

```go
type Package struct {
    PackageManager string      // "gradle", "maven", "npm", "go", "dotnet", "pip"
    PackageName    string      // "group:name" or "name"
    Version        string      // "1.2.3"
    FilePath       string      // Path to manifest file
    Locations      []Location  // Line numbers
}

type Location struct {
    Line       int            // Line number (1-indexed)
    StartIndex int            // Character offset
    EndIndex   int            // Character offset
}
```

### JSON Output Example

```json
[
  {
    "packageManager": "gradle",
    "packageName": "org.springframework:spring-core",
    "version": "5.3.20",
    "filePath": "build.gradle",
    "locations": [
      {
        "line": 42,
        "startIndex": 0,
        "endIndex": 0
      }
    ]
  },
  {
    "packageManager": "maven",
    "packageName": "com.google.guava:guava",
    "version": "31.1-jre",
    "filePath": "pom.xml",
    "locations": [
      {
        "line": 127,
        "startIndex": 0,
        "endIndex": 0
      }
    ]
  }
]
```

---

## 🔒 Security & Vulnerability Detection

This parser is designed to support security scanning and SCA (Software Composition Analysis) tools:

### Integration with Vulnerability Databases

```
Dependency Extraction → Vulnerability Database → Risk Assessment
                            (NVD CVE)
                         (GitHub Advisory)
                         (Snyk Database)
                         (Sonatype OSS)
```

### Example: Detecting Log4j RCE

```gradle
dependencies {
    implementation 'org.apache.logging.log4j:log4j-core:2.14.0'  // CVE-2021-44228
}
```

Parser extracts → `org.apache.logging.log4j:log4j-core:2.14.0`
↓
Vulnerability checker matches → CVE-2021-44228 (CRITICAL - Log4Shell RCE)

---

## 🏗️ Architecture

```
Parser Interface (parser.go)
    ↓
Manifest Detection (manifest-file-selector.go)
    ↓
Parser Factory (parser_factory.go)
    ↓
Language-Specific Parsers
    ├─ Gradle Parser (gradle/gradle_parser.go, gradle/version_catalog.go)
    ├─ Maven Parser (maven/maven-pom-parser.go)
    ├─ npm Parser (npm/package_json_parser.go)
    ├─ Go Parser (golang/go-mod-parser.go)
    ├─ .NET Parsers (dotnet/csproj_parser.go, etc.)
    └─ Python Parser (pypi/pypi-parser.go)
    ↓
Standardized Package Output (models/package_model.go)
```

---

## 🧪 Testing

Run tests for all parsers:

```bash
# Run all tests
go test ./...

# Run specific parser tests
go test ./internal/parsers/gradle/ -v
go test ./internal/parsers/maven/ -v
go test ./internal/parsers/npm/ -v

# With coverage
go test ./... -cover
```

### Test Resources

```
test/resources/
├── build.gradle              (Gradle DSL)
├── build.gradle.kts          (Kotlin DSL)
├── pom.xml                   (Maven)
├── package.json              (npm)
├── test_go.mod               (Go Modules)
├── Bootstrap.csproj          (.NET Framework)
├── Directory.Packages.props   (.NET Centralized)
├── packages.config           (.NET Legacy)
└── requirements.txt          (Python)
```

---

## 📚 Documentation

- [Gradle Parser Details](test/resources/GRADLE_TEST_FILES_README.md) - Comprehensive Gradle documentation with 31 vulnerable dependencies for testing
- [Maven Documentation](https://maven.apache.org/pom.html)
- [npm Documentation](https://docs.npmjs.com/cli/v10/configuring-npm/package-json)
- [Go Modules Documentation](https://go.dev/ref/mod)
- [NuGet Documentation](https://learn.microsoft.com/en-us/nuget/)
- [Pip Documentation](https://pip.pypa.io/)

---

## 🤝 Contributing

Contributions welcome! Focus areas:

- [ ] Add Ruby Bundler support (Gemfile)
- [ ] Add PHP Composer support (composer.json)
- [ ] Add Rust Cargo support (Cargo.toml)
- [ ] Improve version range resolution
- [ ] Add more vulnerability test cases
- [ ] Performance optimizations

---

## ⚖️ License

This project is part of the Checkmarx AST (Application Security Testing) suite.

---

## 🚀 Features Summary

| Feature | Gradle | Maven | npm | Go | .NET | Python |
|---------|--------|-------|-----|----|----|--------|
| Multi-file format | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Property resolution | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Version ranges | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ |
| BOM imports | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Multi-module | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ |
| Line numbers | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Comments/ignored | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Scope separation | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ |

---

## 📝 Version History

- **v3.0.0** - Added Gradle version catalog support, enhanced property resolution
- **v2.5.0** - Added .NET Directory.Packages.props support
- **v2.0.0** - Initial multi-parser support

---

## 📧 Contact & Support

For issues, questions, or feature requests:
- GitHub Issues: [manifest-parser/issues](https://github.com/Checkmarx/manifest-parser/issues)
- Security: [security@checkmarx.com](mailto:security@checkmarx.com)

---

**Made with ❤️ for secure software supply chain management**
