# Package Manager Enum Alignment Report

**Status:** ⚠️ **MISALIGNMENT DETECTED** - Enum values differ between manifest-parser and Dustico API

---

## Comparison Table

| Package Manager | Dustico API (Doc) | Manifest-Parser Code | Status | Notes |
|---|---|---|---|---|
| Ruby | `rubygems` | ❌ NOT IMPLEMENTED | ❌ MISSING | Not yet in manifest-parser |
| Maven | `mvn` | ✅ `"mvn"` | ✅ MATCH | Aligned |
| Node/JavaScript | `npm` | ✅ `"npm"` | ✅ MATCH | Aligned |
| Python | `pypi` | ✅ `"pypi"` | ✅ MATCH | Aligned |
| Dart/Flutter | `pub` | ❌ NOT IMPLEMENTED | ❌ MISSING | Not yet in manifest-parser |
| Go | `go` | ✅ `"go"` | ✅ MATCH | Aligned |
| PHP/Composer | `packagist` | ❌ NOT IMPLEMENTED | ❌ MISSING | Not yet in manifest-parser |
| .NET/C# | `nuget` | ✅ `"nuget"` | ✅ MATCH | Aligned |
| Swift/iOS | `swift` | ✅ `"swift"` | ✅ MATCH | Aligned |
| **NEW** - Gradle | - | ✅ `"gradle"` | ⚠️ ADDITION | Not in Dustico list, we added it |
| **NEW** - SBT | - | ✅ `"sbt"` | ⚠️ ADDITION | Not in Dustico list, we added it |
| **NEW** - CocoaPods | - | ✅ `"cocoapods"` | ⚠️ ADDITION | Not in Dustico list, we added it |
| **NEW** - Carthage | - | ✅ `"carthage"` | ⚠️ ADDITION | Not in Dustico list, we added it |

---

## Issues Identified

### ✅ **FIXED: Swift Package Manager Enum Mismatch**

**Resolution:**
- **Dustico API expects:** `swift`
- **Manifest-Parser now sends:** `swift` ✅

**Location:** `internal/parsers/swiftpm/swiftpm_parser.go:15`
```go
const packageManagerName = "swift"
```

**Changes Applied:**
- ✅ Updated `swiftpm_parser.go` - changed const from `"swiftpm"` to `"swift"`
- ✅ Updated `package_swift_test.go` - test assertions now expect `"swift"`
- ✅ Updated `package_resolved_test.go` - test assertions now expect `"swift"`
- ✅ Updated `ast-cli` - mapping table changed from `"swiftpm": "Ios"` to `"swift": "Ios"`

**Status:** RESOLVED - All enum values now align with Dustico API specification

---

### 🟡 **MISSING IMPLEMENTATIONS**

These package managers are in the Dustico API documentation but NOT YET implemented in manifest-parser:

1. **Ruby (rubygems)** - Not in roadmap yet
2. **PHP/Composer (packagist)** - Not in roadmap yet  
3. **Dart/Flutter (pub)** - Not in roadmap yet

**Action Required:**
- Confirm with Dustico team whether these are needed
- Add to manifest-parser roadmap if required

---

### 🟠 **NEW ADDITIONS NOT IN DUSTICO LIST**

Manifest-parser supports additional package managers NOT listed in Dustico documentation:

1. **Gradle** → sends `"gradle"` to API
2. **SBT** → sends `"sbt"` to API
3. **CocoaPods** → sends `"cocoapods"` to API
4. **Carthage** → sends `"carthage"` to API

**Action Required:**
- Ask Dustico team if they support these new package types
- If not supported, either:
  - Stop sending them to Dustico API, OR
  - Request Dustico to add support for these package types

---

## Questions for Dustico Team

Based on this alignment analysis, we should ask the Dustico team:

1. **Is the package type list complete?** Any new types added after the last documentation update?

2. **Should we use `swift` or `swiftpm` for Swift Package Manager?**
   - Documentation says `swift`
   - Our manifest-parser currently sends `swiftpm`
   - Which is correct?

3. **Do you support the following new package types?**
   - Gradle (`gradle`)
   - SBT (`sbt`)
   - CocoaPods (`cocoapods`)
   - Carthage (`carthage`)

4. **Should we implement the missing package managers?**
   - Ruby/RubyGems (`rubygems`)
   - PHP/Composer (`packagist`)
   - Dart/Flutter (`pub`)

---

## Manifest-Parser Current Implementation Status

### ✅ Fully Implemented (12 total)
1. Maven (`mvn`)
2. npm (`npm`)
3. Python (`pypi`)
4. Go (`go`)
5. .NET (`nuget`)
6. Gradle (`gradle`) - NEW
7. SBT (`sbt`) - NEW
8. Swift PM (`swift`) - ✅ ALIGNED
9. CocoaPods (`cocoapods`) - NEW
10. Carthage (`carthage`) - NEW

### ❌ Not Yet Implemented (3 in Dustico list)
1. Ruby (`rubygems`)
2. PHP (`packagist`)
3. Dart (`pub`)

---

## Recommended Fix Priority

### **Priority 1: IMMEDIATE** ✅ COMPLETED
- ✅ Changed `"swiftpm"` to `"swift"` in manifest-parser
- ✅ Updated ast-cli Real-Time Scanner mapping to send correct enum value
- ✅ Updated all test assertions and documentation files

### **Priority 2: CLARIFICATION**
- Ask Dustico team if they support Gradle, SBT, CocoaPods, Carthage
- Get confirmation on whether `swift` or `swiftpm` is correct
- Get confirmation that the Dustico doc is current

### **Priority 3: IMPLEMENTATION (if needed)**
- Add Ruby, PHP, Dart parsers to manifest-parser if Dustico requests them
- Update Real-Time Scanner to handle any new package types

---

## Code Change Required

**File:** `internal/parsers/swiftpm/swiftpm_parser.go`

**Current:**
```go
const packageManagerName = "swiftpm"
```

**Should be:**
```go
const packageManagerName = "swift"
```

**Also verify:** All tests and Real-Time Scanner code that references this value.

