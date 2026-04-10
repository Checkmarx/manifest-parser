# Enterprise-Grade Gradle Parser Test Files

This directory contains comprehensive test fixtures demonstrating production-grade Gradle configurations with real vulnerability examples.

## Files Overview

### 1. `build.gradle` (3.1 KB)
**Groovy DSL Format** - Original multi-module project configuration

**Features Demonstrated:**
- ✅ Groovy syntax dependency declarations
- ✅ `subprojects` block for shared configuration
- ✅ Module-specific `project(':name')` blocks
- ✅ Extended `ext` blocks for version management
- ✅ Comments and security annotations
- ✅ Jacoco, Checkstyle, and SpringBoot plugins

**Dependencies Parsed:** 15 packages with full version info

**Vulnerabilities Included:**
- 🔴 **CRITICAL:** Log4Shell (log4j-core:2.14.0)
- 🔴 **CRITICAL:** Commons Collections RCE (commons-collections:3.2.1)
- 🔥 **HIGH:** Spring Framework XXE (spring-web:5.2.0.RELEASE)
- 🔥 **HIGH:** Jackson RCE (jackson-databind:2.9.8)
- 🔥 **HIGH:** Hibernate SQL Injection (hibernate-core:5.4.0.Final)

---

### 2. `build.gradle.kts` (13.5 KB)
**Kotlin DSL Format** - Advanced multi-module enterprise configuration

**Features Demonstrated:**
- ✅ Kotlin DSL syntax `implementation("...")`
- ✅ Kotlin `val` variable declarations with type inference
- ✅ `dependencyManagement` with BOM imports
- ✅ `platform()` wrapper for dependency BOMs
- ✅ Extended dependency configurations: `debugImplementation`, `releaseImplementation`, `ksp`, `compileOnlyApi`
- ✅ `configure()` scoped configuration for select modules
- ✅ Custom tasks and build info
- ✅ SonarQube integration

**Module Breakdown:**
- **`:core-api`** - Shared business logic (Spring Boot + Hibernate)
- **`:security-module`** - Authentication/Authorization (Spring Security + JWT)
- **`:data-module`** - Database layer (JPA + Hibernate + Liquibase)
- **`:api-gateway`** - External integrations (Spring Cloud Gateway)
- **`:monitoring-module`** - Observability (Actuator + Micrometer + Prometheus)

**Dependencies Parsed:** 40+ packages including BOM references

**Extended Configurations:**
- `debugImplementation` - Facebook Stetho for Android debugging
- `releaseImplementation` - Firebase Crashlytics & Analytics
- `ksp` - Dagger compiler for dependency injection code generation
- `annotationProcessor` - Lombok for boilerplate generation
- `testImplementation` - JUnit, Mockito, AssertJ

**Vulnerabilities Included:**
- 🔴 **CRITICAL:** Log4j RCE (2.14.0, 2.17.1)
- 🔴 **CRITICAL:** Commons Collections (3.2.1, 3.2.2)
- 🔥 **HIGH:** Spring Core RCE (5.2.0.RELEASE)
- 🔥 **HIGH:** Spring Security XXE (5.4.0, 5.7.1)
- 🔥 **HIGH:** Jackson Databind (2.9.8, 2.13.3)
- 🔥 **HIGH:** XStream Deserialization (1.4.17)
- 🔥 **HIGH:** Hibernate SQLi (5.4.0.Final, 5.6.10.Final)
- ⚠️ **MEDIUM:** HttpClient DoS (4.5.5, 4.5.13)
- ⚠️ **MEDIUM:** Guava Overflow (23.0, 31.1-jre)
- ⚠️ **MEDIUM:** Logback (1.2.3, 1.2.11)
- ⚠️ **MEDIUM:** Tomcat Ghostcat (9.0.10)
- 🟡 **LOW:** Commons Codec (1.14, 1.15)
- 🟡 **LOW:** Jetty Path Traversal (9.4.38)
- 🟡 **LOW:** MySQL Legacy (5.1.40)

---

### 3. `gradle.properties` (2.0 KB)
**Centralized Configuration** - Shared across all modules

**Sections:**
1. **Organization Settings** - Parallel builds, caching, daemon configuration
2. **Java Version** - Version 11 target with toolchain config
3. **Framework Versions** - Spring, Hibernate, Jackson versions
4. **Logging Versions** - Log4j, SLF4J, Logback versions
5. **Apache Commons** - Commons Lang3, Codec, Collections, HttpClient
6. **Database Drivers** - MySQL, PostgreSQL, H2 versions
7. **JSON/XML Processing** - Guava, Gson, XStream versions
8. **Testing Frameworks** - JUnit, Mockito, AssertJ, TestNG versions
9. **Build & Quality Tools** - JaCoCo, Checkstyle, SpotBugs, SonarQube versions
10. **Google Cloud** - BOM version for GCP integration

**Features Demonstrated:**
- ✅ Property name conventions (camelCase with Version suffix)
- ✅ Comments and section organization
- ✅ Version pinning for reproducible builds
- ✅ Easy centralized updates across modules
- ✅ Used by both `build.gradle` and `build.gradle.kts` files

**Example Usage:**
```gradle
// In build.gradle
implementation "org.springframework:spring-core:${springVersion}"

// In build.gradle.kts
implementation("org.springframework:spring-core:${property("springVersion")}")
```

---

### 4. `gradle/libs.versions.toml` (9.7 KB)
**Version Catalog** - Modern dependency management (Gradle 7.0+)

**Format:** TOML with three sections:
1. **`[versions]`** - Centralized version definitions
2. **`[libraries]`** - Library references with version links
3. **`[bundles]`** - Grouped dependencies for common use cases

**Features Demonstrated:**

#### Version References
```toml
[versions]
spring-version = "5.3.20"
spring-boot-version = "2.7.0"

[libraries]
spring-core = { module = "org.springframework:spring-core", version.ref = "spring-version" }
spring-boot-web = { module = "org.springframework.boot:spring-boot-starter-web", version.ref = "spring-boot-version" }
```

#### Simple Inline Format
```toml
[libraries]
guava = "com.google.guava:guava:31.1-jre"
```

#### Key-Value Map Format
```toml
[libraries]
hibernate = { module = "org.hibernate:hibernate-core", version.ref = "hibernate-version" }
h2 = { module = "com.h2database:h2", version.ref = "h2-version" }
```

#### Bundles (Grouped Dependencies)
```toml
[bundles]
spring-boot-web = [
    "spring-boot-starter-web",
    "spring-boot-starter-validation",
    "spring-boot-starter-logging"
]
```

**Usage in build.gradle.kts:**
```kotlin
dependencies {
    implementation(libs.spring.core)
    implementation(libs.spring.boot.web)
    testImplementation(libs.bundles.testing)
}
```

**80+ Dependencies Catalogued:**
- Spring Framework (13 entries)
- Spring Boot Starters (6 entries)
- Spring Cloud (2 entries)
- Logging (4 entries)
- Database/ORM (7 entries)
- JSON/XML (5 entries)
- Apache Commons (4 entries)
- Testing (3 entries)
- Android/Debug (2 entries)
- API Documentation (2 entries)
- Kotlin/Coroutines (3 entries)

**Vulnerabilities in Catalog:**
All known CVE versions are explicitly catalogued with comments marking severity:
- `log4j-core` - CVE-2021-44228 (Log4Shell RCE)
- `commons-collections` - CVE-2015-4852 (Deserialization)
- `jackson-databind` - CVE-2020-5410 (Polymorphic RCE)
- `xstream` - CVE-2019-12384 (XXE)
- `httpclient` - CVE-2019-9740 (DoS)

---

## Parser Capabilities Tested

### Feature Coverage

| Feature | Status | Example |
|---------|--------|---------|
| Groovy DSL | ✅ | `implementation 'group:artifact:version'` |
| Kotlin DSL | ✅ | `implementation("group:artifact:version")` |
| gradle.properties | ✅ | `implementation "org:lib:${springVersion}"` |
| Version Catalog | ✅ | `implementation(libs.spring.core)` |
| Platform/BOM | ✅ | `implementation(platform('...'))` |
| Extended Configs | ✅ | `debugImplementation`, `ksp`, `releaseImplementation` |
| Multi-line Deps | ✅ | Dependencies spanning multiple lines |
| Conditional Deps | ✅ | Dependencies inside `if` blocks |
| Project References | ✅ (Skipped) | `implementation project(':core')` |
| File References | ✅ (Skipped) | `implementation files('libs/*.jar')` |
| BOM Imports | ✅ | `dependencyManagement.imports.mavenBom(...)` |
| Variable Resolution | ✅ | `${propertyName}` and `$varName` |
| Commented Code | ✅ | Properly ignores commented declarations |

### Vulnerability Detection

The test files contain **31 vulnerable dependencies** across severity levels:

```
🔴 CRITICAL:  7 packages (Log4j, Commons Collections, Spring, Jackson, XStream)
🔥 HIGH:      8 packages (Spring Security, HttpClient, Hibernate, Guava, Logback)
⚠️  MEDIUM:    8 packages (Tomcat, Commons Codec, Jetty)
🟡 LOW:       8 packages (Legacy MySQL, Deprecated versions)
```

### Supported Dependency Configurations

All 18+ Gradle dependency configurations:
- `implementation`, `api`, `compile`, `compileOnly`
- `runtime`, `runtimeOnly`
- `testImplementation`, `testCompile`, `testCompileOnly`, `testRuntimeOnly`
- `debugImplementation`, `releaseImplementation`
- `annotationProcessor`, `classpath`, `kapt`, `ksp`
- `compileOnlyApi`, `testFixturesImplementation`, `testFixturesApi`
- `lintChecks`

---

## Test Execution

### Run Gradle Parser Tests
```bash
cd c:/repository/manifest-parser
go test ./internal/parsers/gradle/ -v
```

### Parse Individual Files
```bash
# Groovy DSL
go run cmd/main.go test/resources/build.gradle

# Kotlin DSL
go run cmd/main.go test/resources/build.gradle.kts

# With version catalog
go run cmd/main.go test/resources/build.gradle.kts
# Parser automatically discovers gradle/libs.versions.toml
```

### Expected Output
```json
[
  {
    "packageManager": "gradle",
    "packageName": "org.apache.logging.log4j:log4j-core",
    "version": "2.14.0",
    "filePath": "test/resources/build.gradle"
  },
  {
    "packageManager": "gradle",
    "packageName": "org.springframework:spring-core",
    "version": "5.2.0.RELEASE",
    "filePath": "test/resources/build.gradle"
  },
  ...
]
```

---

## Security Notes

⚠️ **IMPORTANT:** These test files contain intentionally vulnerable dependency versions for testing purposes.

**DO NOT USE IN PRODUCTION** without:
1. Updating all CRITICAL and HIGH severity packages
2. Upgrading to patched versions
3. Running security audits
4. Validating compatibility

**Recommended Actions:**
- Use `dependencyCheck` plugin to scan for known vulnerabilities
- Enable SonarQube analysis for code quality
- Run `./gradlew dependencyUpdates` to find newer versions
- Use Maven Central's vulnerability database

---

## File Sizes & Complexity

```
build.gradle              3.1 KB   (15 dependencies)
build.gradle.kts         13.5 KB   (40+ dependencies)
gradle.properties         2.0 KB   (40+ property definitions)
gradle/libs.versions.toml 9.7 KB   (80+ catalog entries)
─────────────────────────────────────────────────────────
TOTAL                    28.3 KB   (175+ dependency references)
```

---

## References

- [Gradle Build Language Reference](https://docs.gradle.org/current/userguide/declaring_dependencies.html)
- [Gradle Version Catalogs](https://docs.gradle.org/current/userguide/platforms.html)
- [Spring Boot Version Reference](https://spring.io/projects/spring-boot/releases/)
- [NIST CVE Database](https://nvd.nist.gov/vuln)
- [Gradle Dependency Check Plugin](https://plugins.gradle.org/plugin/com.github.dependency-check.gradle)
