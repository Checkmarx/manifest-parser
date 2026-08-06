// swift-tools-version:5.5
//
// Test fixture covering the .package(...) declaration forms supported by the
// SwiftPM parser. Intentionally pins old / vulnerable versions of well-known
// packages so vulnerability scanners have something to flag.

import PackageDescription

let package = Package(
    name: "ManifestParserTestApp",
    platforms: [
        .iOS(.v13),
        .macOS(.v10_15),
    ],
    dependencies: [
        // from: — minimum version (>= 2.0.0)
        .package(url: "https://github.com/apple/swift-nio.git", from: "2.0.0"),

        // .upToNextMajor — semver-major bound
        .package(url: "https://github.com/apple/swift-log.git", .upToNextMajor(from: "1.4.0")),

        // .upToNextMinor — semver-minor bound
        .package(url: "https://github.com/apple/swift-crypto.git", .upToNextMinor(from: "2.0.0")),

        // exact: — pinned single version
        .package(url: "https://github.com/Alamofire/Alamofire.git", exact: "5.2.0"),

        // Multi-line declaration, .upToNextMajor inside
        .package(
            url: "https://github.com/SnapKit/SnapKit.git",
            .upToNextMajor(from: "5.0.0")
        ),

        // Legacy named form (name: + url: + from:)
        .package(name: "Kingfisher", url: "https://github.com/onevcat/Kingfisher.git", from: "7.0.0"),

        // branch: — non-semver, resolves to "latest"
        .package(url: "https://github.com/apple/swift-collections.git", branch: "main"),

        // revision: — non-semver, resolves to "latest"
        .package(url: "https://github.com/apple/swift-syntax.git", revision: "5e9b6f0d"),

        // Registry-identifier form
        .package(id: "apple.swift-algorithms", from: "1.0.0"),

        // Local path — must be SKIPPED by the parser
        .package(path: "../LocalDep"),
    ],
    targets: [
        .target(name: "ManifestParserTestApp", dependencies: []),
    ]
)
