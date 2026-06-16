# iOS Package Managers Implemented

We've implemented **3 iOS/macOS package managers** for the manifest-parser. Here's everything about them:

---

## Overview Table

| Package Manager | Language | File Types | Status | Complexity |
|---|---|---|---|---|
| **Swift PM** (Swift Package Manager) | Swift | `Package.swift`, `Package.resolved` | ✅ Complete | Medium |
| **CocoaPods** | Ruby/YAML/JSON | `Podfile`, `Podfile.lock`, `.podspec`, `.podspec.json` | ✅ Complete | High |
| **Carthage** | Plaintext | `Cartfile`, `Cartfile.private`, `Cartfile.resolved` | ✅ Complete | Low |

---

## 1️⃣ Swift PM (Swift Package Manager)

### What is Swift PM?

Apple's official package manager for Swift. It's built into Xcode and is the modern standard for Swift projects.

**Use case:** Swift apps, libraries, server-side Swift projects

### File Types We Support

#### A. `Package.swift` (Manifest File)

This is where dependencies are declared.

**Real Example:**
```swift
import PackageDescription

let package = Package(
    name: "MyApp",
    dependencies: [
        // Minimum version
        .package(url: "https://github.com/apple/swift-nio.git", from: "2.0.0"),
        
        // Exact version
        .package(url: "https://github.com/Alamofire/Alamofire.git", exact: "5.2.0"),
        
        // Version range
        .package(url: "https://github.com/SnapKit/SnapKit.git", .upToNextMajor(from: "5.0.0")),
        
        // Branch reference
        .package(url: "https://github.com/apple/swift-collections.git", branch: "main"),
    ],
    targets: [
        .target(name: "MyApp", dependencies: ["Alamofire", "SnapKit"]),
    ]
)
```

**What our parser extracts:**
```json
{
  "PackageName": "Alamofire",
  "Version": "5.2.0",
  "PackageManager": "swift",
  "FilePath": "Package.swift",
  "Locations": [{"Line": 10, "StartIndex": 24, "EndIndex": 67}]
}
```

**Version Specifiers We Handle:**
- `from: "1.0.0"` → `>=` version
- `exact: "1.0.0"` → Pinned version
- `.upToNextMajor(from: "1.0.0")` → `1.x.x` range
- `.upToNextMinor(from: "1.0.0")` → `1.x.x` range
- `branch: "main"` → Resolves to `latest`
- `revision: "abc123"` → Resolves to `latest`

#### B. `Package.resolved` (Lock File)

This is generated when you run `swift build`. It pins exact resolved versions.

**Real Example:**
```json
{
  "pins": [
    {
      "package": "Alamofire",
      "repositoryURL": "https://github.com/Alamofire/Alamofire.git",
      "state": {
        "branch": null,
        "revision": "f82c23a8a7ef8dc1a765d6098294df1c4572cb43",
        "version": "5.2.0"
      }
    }
  ]
}
```

**What our parser does:**
- Extracts the pinned version from the lock file
- Uses it to provide accurate vulnerability reporting
- Without the lock file, we'd have to say "version unknown"

### Files Included in Implementation

```
internal/parsers/swiftpm/
├── swiftpm_parser.go           ← Main dispatcher
├── package_swift.go            ← Parses Package.swift
├── package_resolved.go         ← Parses Package.resolved
├── package_swift_test.go       ← Tests for Package.swift
└── package_resolved_test.go    ← Tests for Package.resolved
```

---

## 2️⃣ CocoaPods

### What is CocoaPods?

The most widely used dependency manager for iOS/macOS. Uses Ruby DSL and supports a centralized pod repository.

**Use case:** Almost every iOS/macOS app from 2010-2023

### File Types We Support

#### A. `Podfile` (Manifest File)

Declares dependencies in Ruby DSL.

**Real Example:**
```ruby
platform :ios, '13.0'

target 'MyApp' do
  pod 'Alamofire', '~> 5.0'
  pod 'SwiftyJSON', '5.0.0'
  pod 'RealmSwift'
  
  target 'MyAppTests' do
    inherit! :search_paths
    pod 'Quick'
    pod 'Nimble'
  end
end
```

**What our parser extracts:**
```json
{
  "PackageName": "Alamofire",
  "Version": "latest",  // ~> 5.0 is a range, so "latest"
  "PackageManager": "cocoapods",
  "FilePath": "Podfile",
  "Locations": [{"Line": 5, "StartIndex": 7, "EndIndex": 32}]
}
```

**Version Specifiers We Handle:**
- `pod 'Name', '1.0.0'` → Exact version
- `pod 'Name', '~> 1.0'` → Compatible release range → `latest`
- `pod 'Name', '> 1.0'` → Minimum version → `latest`
- `pod 'Name'` → No version specified → `latest`
- `pod 'Name', :git => 'url', :branch => 'main'` → Git reference → `latest`

#### B. `Podfile.lock` (Lock File)

Generated when you run `pod install`. Pins exact resolved versions.

**Real Example:**
```yaml
PODS:
  - Alamofire (5.2.0)
  - SwiftyJSON (5.0.0)
  - RealmSwift (10.5.0)

DEPENDENCIES:
  - Alamofire (~> 5.0)
  - SwiftyJSON (= 5.0.0)
  - RealmSwift

SPEC REPOS:
  trunk:
    - Alamofire
    - SwiftyJSON
    - RealmSwift

SPEC CHECKSUMS:
  Alamofire: abc123...
  SwiftyJSON: def456...
  RealmSwift: ghi789...

PODFILE CHECKSUM: jkl012...
```

**What our parser does:**
- Reads both `Podfile` AND `Podfile.lock`
- If there's a lock file, uses the pinned version
- If no lock file, uses the version spec from Podfile

#### C. `.podspec` (Pod Specification File)

Describes a CocoaPod (library definition in Ruby).

**Real Example:**
```ruby
Pod::Spec.new do |s|
  s.name         = "MyLib"
  s.version      = "1.0.0"
  
  s.dependencies do |dep|
    dep.add_dependency 'Alamofire', '~> 5.0'
    dep.add_dependency 'SwiftyJSON', '5.0.0'
  end
end
```

**What our parser extracts:**
```json
{
  "PackageName": "Alamofire",
  "Version": "latest",
  "PackageManager": "cocoapods",
  "FilePath": "MyLib.podspec",
  "Locations": [...]
}
```

#### D. `.podspec.json` (Pod Specification in JSON)

Same as `.podspec` but in JSON format.

**Real Example:**
```json
{
  "name": "MyLib",
  "version": "1.0.0",
  "dependencies": {
    "Alamofire": ["~> 5.0"],
    "SwiftyJSON": ["5.0.0"]
  }
}
```

**What our parser does:** Same as `.podspec`, just different format

### Files Included in Implementation

```
internal/parsers/cocoapods/
├── cocoapods_parser.go        ← Main dispatcher
├── podfile.go                 ← Parses Podfile (Ruby DSL)
├── podfile_lock.go            ← Parses Podfile.lock (YAML)
├── podspec.go                 ← Parses .podspec (Ruby DSL)
├── podspec_json.go            ← Parses .podspec.json (JSON)
└── cocoapods_parser_test.go   ← Unit tests
```

### Why CocoaPods is Complex

1. **Ruby DSL parsing** – Not as structured as JSON/TOML
2. **Nested targets** – Podfile can have nested dependencies (app, tests, frameworks)
3. **Multiple file types** – Need parsers for Podfile + Podfile.lock + .podspec + .podspec.json
4. **Transitive dependencies** – One pod can depend on many others
5. **Version syntax** – Supports semantic versioning ranges (`~>`, `>=`, `>`, etc.)

---

## 3️⃣ Carthage

### What is Carthage?

A decentralized dependency manager for iOS/macOS. Simpler than CocoaPods, uses build artifacts instead of source.

**Use case:** Teams that want less magic, prefer binary frameworks

### File Types We Support

#### A. `Cartfile` (Manifest File)

Declares dependencies using simple syntax.

**Real Example:**
```
github "Alamofire/Alamofire" "5.2.0"
github "ReactiveCocoa/ReactiveSwift" == 6.7.0
github "onevcat/Kingfisher" ~> 7.0
git "https://example.com/internal/MyFramework.git" "1.4.2"
binary "https://my.domain.com/release/MyFramework.json" ~> 2.3
```

**What our parser extracts:**
```json
{
  "PackageName": "Alamofire",
  "Version": "5.2.0",
  "PackageManager": "carthage",
  "FilePath": "Cartfile",
  "Locations": [{"Line": 0, "StartIndex": 0, "EndIndex": 42}]
}
```

**Three Origin Types:**
1. **github** – GitHub repos: `github "owner/repo" "version"`
2. **git** – Any Git URL: `git "https://..." "version"`
3. **binary** – Prebuilt frameworks: `binary "https://..." "version"`

**Version Specifiers:**
- Quoted exact: `"5.2.0"` → Exact version
- Equality: `== 6.7.0` → Exact version
- Range: `~> 7.0` → Latest compatible → `latest`
- Greater: `>= 1.0` → Latest matching → `latest`
- No spec: Just `github "owner/repo"` → `latest`

#### B. `Cartfile.private` (Private Dependencies)

Same as Cartfile but for internal/test dependencies not exposed to consumers.

**Real Example:**
```
github "Quick/Nimble" "9.2.0"
github "pointfreeco/swift-snapshot-testing"
git "https://example.com/internal/TestUtils.git" branch "main"
```

#### C. `Cartfile.resolved` (Lock File)

Generated when you run `carthage update`. Pins exact versions.

**Real Example:**
```
github "Alamofire/Alamofire" "5.2.0"
github "ReactiveCocoa/ReactiveSwift" "6.7.0"
github "onevcat/Kingfisher" "7.6.2"
git "https://example.com/internal/MyFramework.git" "1.4.2"
binary "https://my.domain.com/release/MyFramework.json" "2.3.1"
```

**What our parser does:**
- Reads all three files
- `Cartfile` provides dependency declarations
- `Cartfile.private` adds private dependencies
- `Cartfile.resolved` provides pinned versions

### Files Included in Implementation

```
internal/parsers/carthage/
├── carthage_parser.go      ← Main parser (handles all 3 file types)
└── carthage_parser_test.go ← Unit tests
```

### Why Carthage is Simple

1. **Plain text format** – One line per dependency
2. **Consistent structure** – `origin "source" "version"`
3. **Single file type** – One parser handles all variants
4. **No nesting** – No nested dependencies or targets

---

## 📊 Comparison: SwiftPM vs CocoaPods vs Carthage

| Aspect | SwiftPM | CocoaPods | Carthage |
|---|---|---|---|
| **Origin** | Apple (official) | Community | Community |
| **When** | Modern (2017+) | Legacy (2011+) | Mid-range (2014+) |
| **Language** | Swift DSL | Ruby DSL | Plain text |
| **Adoption** | Growing rapidly | Widespread (legacy) | Declining |
| **Lock File** | Package.resolved | Podfile.lock | Cartfile.resolved |
| **Multi-file** | 2 types | 4 types | 3 types |
| **Parsing Complexity** | Medium | High | Low |
| **Version Specifiers** | 6 types | 4 types | 4 types |

---

## Test Fixtures

We have test data for all three:

```
test/resources/
├── Package.swift              ← SwiftPM declaration
├── Package.resolved           ← SwiftPM lock file
├── Podfile                    ← CocoaPods declaration
├── Podfile.lock               ← CocoaPods lock file
├── TestPod.podspec            ← CocoaPods spec file
├── TestPod.podspec.json       ← CocoaPods JSON spec
├── Cartfile                   ← Carthage declaration
├── Cartfile.private           ← Carthage private deps
└── Cartfile.resolved          ← Carthage lock file
```

---

## Usage in Demo

### Showing iOS Parsers in a Demo:

**Script:**
> "We also support iOS and macOS dependency managers. Here we have three: Swift PM (Apple's official), CocoaPods (the legacy standard), and Carthage (decentralized alternative).
>
> Each has different file formats, but our parser handles them all and provides real-time vulnerability scanning with exact line tracking."

**What to show:**
1. Open `Package.swift` - Show SwiftPM dependencies
2. Open `Podfile` - Show CocoaPods dependencies
3. Open `Cartfile` - Show Carthage dependencies
4. Highlight vulnerable versions in each
5. Point out how version resolution works for each

---

## Code Files Structure

### SwiftPM Parser Files:

- **swiftpm_parser.go** – Main entry point, dispatches to Package.swift or Package.resolved parser
- **package_swift.go** – Parses Swift DSL in Package.swift file
- **package_resolved.go** – Parses JSON lock file

### CocoaPods Parser Files:

- **cocoapods_parser.go** – Main dispatcher
- **podfile.go** – Parses Ruby DSL Podfile syntax
- **podfile_lock.go** – Parses YAML Podfile.lock
- **podspec.go** – Parses Ruby DSL .podspec files
- **podspec_json.go** – Parses JSON .podspec files

### Carthage Parser Files:

- **carthage_parser.go** – Single parser handles all 3 file types (Cartfile, Cartfile.private, Cartfile.resolved)

---

## In Summary

✅ **Swift PM**: Modern, growing, 2 file types
✅ **CocoaPods**: Widely used, complex, 4 file types  
✅ **Carthage**: Simple, declining, 3 file types

All three are fully implemented with:
- ✅ Dependency extraction
- ✅ Version resolution
- ✅ Location tracking (line & character)
- ✅ Lock file support
- ✅ Comprehensive unit tests
- ✅ Real test fixtures

