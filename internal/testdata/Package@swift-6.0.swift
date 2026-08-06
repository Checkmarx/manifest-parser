// swift-tools-version:6.0
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "VulnerableSwiftApp",
    platforms: [
        .macOS(.v10_15),
        .iOS(.v13),
    ],
    dependencies: [
        // VULNERABLE: Alamofire 4.9.1 has known security issues (CVE-2022-xxxx)
        .package(url: "https://github.com/Alamofire/Alamofire.git", exact: "4.9.1"),

        // VULNERABLE: Cryptoswift has CVE vulnerabilities in older versions
        .package(url: "https://github.com/krzyzanowskim/CryptoSwift.git", .upToNextMajor(from: "1.0.0")),

        // VULNERABLE: Starscream 3.1.1 has security vulnerabilities
        .package(url: "https://github.com/daltoniam/Starscream.git", exact: "3.1.1"),

        // VULNERABLE: Kingfisher 5.15.8 has known issues
        .package(url: "https://github.com/onevcat/Kingfisher.git", .upToNextMinor(from: "5.15.8")),

        // VULNERABLE: SwiftyJSON 4.3.0 has outdated dependencies
        .package(url: "https://github.com/SwiftyJSON/SwiftyJSON.git", exact: "4.3.0"),

        // VULNERABLE: RxSwift 5.0.0 has known vulnerabilities
        .package(url: "https://github.com/ReactiveX/RxSwift.git", .upToNextMajor(from: "5.0.0")),

        // VULNERABLE: Vapor 3.3.0 (older version with CVEs)
        .package(url: "https://github.com/vapor/vapor.git", exact: "3.3.0"),

        // VULNERABLE: SQLite.swift older version
        .package(url: "https://github.com/stephencelis/SQLite.swift.git", .upToNextMajor(from: "0.11.0")),

        // VULNERABLE: DateTools has security issues in older releases
        .package(url: "https://github.com/MatthewYork/DateTools.git", exact: "2.0.0"),

        // VULNERABLE: ObjectMapper 3.5.0
        .package(url: "https://github.com/tristanhimmelman/ObjectMapper.git", exact: "3.5.0"),
    ],
    targets: [
        .target(
            name: "VulnerableSwiftApp",
            dependencies: [
                .product(name: "Alamofire", package: "Alamofire"),
                .product(name: "CryptoSwift", package: "CryptoSwift"),
                .product(name: "Starscream", package: "Starscream"),
                .product(name: "Kingfisher", package: "Kingfisher"),
                .product(name: "SwiftyJSON", package: "SwiftyJSON"),
                .product(name: "RxSwift", package: "RxSwift"),
                .product(name: "Vapor", package: "Vapor"),
                .product(name: "SQLite", package: "SQLite.swift"),
                "DateTools",
                .product(name: "ObjectMapper", package: "ObjectMapper"),
            ],
            path: "Sources"
        ),
        .testTarget(
            name: "VulnerableSwiftAppTests",
            dependencies: [
                .target(name: "VulnerableSwiftApp"),
                .product(name: "RxSwift", package: "RxSwift"),
            ],
            path: "Tests"
        ),
    ]
)
