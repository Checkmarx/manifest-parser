/*
 * Enterprise-Grade Multi-Module Gradle Build Configuration
 *
 * This build.gradle.kts demonstrates:
 * - Kotlin DSL dependency declarations
 * - Variable resolution from gradle.properties
 * - Platform/BOM dependencies
 * - Version catalog references (with libs.versions.toml)
 * - Extended dependency configurations
 * - Production-ready vulnerability examples
 */

import java.time.Instant

plugins {
    kotlin("jvm") version "1.6.21" apply false
    id("org.springframework.boot") version "2.7.0" apply false
    id("io.spring.dependency-management") version "1.0.11.RELEASE"
    id("org.sonarqube") version "3.4.0.2513" apply false
    id("jacoco")
    id("checkstyle")
}

group = "com.enterprise.platform"
version = "3.1.0"

repositories {
    mavenCentral()
    google()
    maven(url = "https://plugins.gradle.org/m2/")
}

/**
 * Configure all subprojects with common settings
 */
subprojects {
    apply(plugin = "java")
    apply(plugin = "jacoco")
    apply(plugin = "checkstyle")

    java {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
        toolchain {
            languageVersion.set(JavaLanguageVersion.of(11))
        }
    }

    repositories {
        mavenCentral()
        google()
    }

    dependencyManagement {
        imports {
            mavenBom("org.springframework.cloud:spring-cloud-dependencies:${property("springCloudVersion")}")
            mavenBom("com.google.cloud:libraries-bom:${property("googleCloudBomVersion")}")
        }
    }

    dependencies {
        // ===============================================================
        // 🔴 CRITICAL VULNERABILITIES - MUST BE REMEDIATED
        // ===============================================================
        // CVE-2021-44228 (Log4j RCE) - Apache Log4j 2.14.0
        // DO NOT USE IN PRODUCTION
        implementation("org.apache.logging.log4j:log4j-core:2.14.0")

        // CVE-2015-4852 (Deserialization RCE) - Commons Collections 3.2.1
        // Gadget chain exploitable with certain frameworks
        implementation("commons-collections:commons-collections:3.2.1")

        // ===============================================================
        // 🔥 HIGH VULNERABILITIES - SHOULD UPGRADE
        // ===============================================================
        // CVE-2019-2725 (RCE) - Spring Framework 5.2.0
        // Improper validation in Spring Core
        implementation("org.springframework:spring-core:5.2.0.RELEASE")

        // CVE-2020-5410 (Arbitrary File Write) - Jackson Databind 2.9.8
        // Multiple polymorphic deserialization gadgets
        implementation("com.fasterxml.jackson.core:jackson-databind:2.9.8")

        // CVE-2019-12384 (Deserialization RCE) - XStream 1.4.17
        // Unsafe unmarshalling of XML data
        implementation("com.thoughtworks.xstream:xstream:1.4.17")

        // CVE-2019-2725 (SQL Injection) - Hibernate 5.4.0
        // HQL injection via eager initialization of associations
        implementation("org.hibernate:hibernate-core:5.4.0.Final")

        // ===============================================================
        // ⚠️ MEDIUM VULNERABILITIES - PLAN UPGRADES
        // ===============================================================
        // CVE-2021-21341 (XXE) - org.springframework.security 5.4.0
        // XML External Entity vulnerability in XML parsing
        implementation("org.springframework.security:spring-security-core:5.4.0")

        // CVE-2019-9740 (DoS) - Apache HttpClient 4.5.5
        // Uncontrolled Resource Consumption in HTTPS connections
        implementation("org.apache.httpcomponents:httpclient:4.5.5")

        // CVE-2018-14335 (Missing bounds check) - Guava 23.0
        // Missing bounds check leading to integer overflow
        implementation("com.google.guava:guava:23.0")

        // CVE-2019-1010022 (Buffer Overflow) - Logback 1.2.3
        // Improper input validation in configuration parsing
        implementation("ch.qos.logback:logback-classic:1.2.3")

        // ===============================================================
        // 🟡 LOW VULNERABILITIES - MONITOR
        // ===============================================================
        // CVE-2020-1938 (AJP Ghostcat) - Tomcat Embed 9.0.10
        // Arbitrary file read/write via AJP protocol
        implementation("org.apache.tomcat.embed:tomcat-embed-core:9.0.10")

        // CVE-2020-13956 (DoS) - Apache Commons Codec 1.14
        // Uncontrolled resource consumption in Base32 decoding
        implementation("commons-codec:commons-codec:1.14")

        // CVE-2020-17527 (Path Traversal) - Jetty 9.4.38
        // URI path traversal via encoded characters
        implementation("org.eclipse.jetty:jetty-server:9.4.38.v20210224")

        // ===============================================================
        // DATABASE DRIVERS
        // ===============================================================
        // Production-grade: PostgreSQL (Recommended over MySQL for security)
        implementation("org.postgresql:postgresql:${property("postgresqlVersion")}")

        // Legacy MySQL (deprecated in favor of PostgreSQL)
        implementation("mysql:mysql-connector-java:5.1.40")

        // In-memory testing database
        testImplementation("com.h2database:h2:${property("h2Version")}")

        // ===============================================================
        // TESTING FRAMEWORKS
        // ===============================================================
        testImplementation("junit:junit:${property("junitVersion")}")
        testImplementation("org.mockito:mockito-core:${property("mockitoVersion")}")
        testImplementation("org.assertj:assertj-core:${property("assertjVersion")}")
        testImplementation("org.testng:testng:${property("testngVersion")}")

        // ===============================================================
        // QUALITY & OBSERVABILITY
        // ===============================================================
        implementation("org.slf4j:slf4j-api:${property("slf4jVersion")}")

        // Annotation processing
        annotationProcessor("org.projectlombok:lombok:1.18.24")
        testAnnotationProcessor("org.projectlombok:lombok:1.18.24")
    }

    // Configure Checkstyle
    checkstyle {
        toolVersion = "10.2"
        configFile = file("${rootProject.projectDir}/checkstyle.xml")
    }

    // Configure JaCoCo
    jacoco {
        toolVersion = "0.8.8"
    }

    tasks.jacocoTestReport {
        reports {
            xml.required.set(true)
            html.required.set(true)
            csv.required.set(false)
        }
    }

    tasks.test {
        useJUnitPlatform()
        finalizedBy(tasks.jacocoTestReport)
    }
}

/**
 * Core API Module
 * Contains shared business logic and data access layer
 */
project(":core-api") {
    apply(plugin = "org.springframework.boot")
    apply(plugin = "kotlin")

    dependencies {
        // Spring Framework Core
        implementation("org.springframework.boot:spring-boot-starter-web")
        implementation("org.springframework.boot:spring-boot-starter-data-jpa")
        implementation("org.springframework.boot:spring-boot-starter-validation")

        // Spring Security (vulnerable version)
        implementation("org.springframework.security:spring-security-core:${property("springSecurityVersion")}")

        // Kotlin Support
        implementation(kotlin("stdlib-jdk11"))
        implementation(kotlin("reflect"))
    }
}

/**
 * Security Module
 * Contains authentication and authorization logic
 */
project(":security-module") {
    apply(plugin = "org.springframework.boot")

    dependencies {
        implementation(project(":core-api"))

        // Spring Security stack
        implementation("org.springframework.security:spring-security-core:${property("springSecurityVersion")}")
        implementation("org.springframework.security:spring-security-crypto:${property("springSecurityVersion")}")
        implementation("org.springframework.security:spring-security-web:${property("springSecurityVersion")}")

        // JWT/OAuth2
        implementation("io.jsonwebtoken:jjwt:0.11.5")

        // LDAP Integration
        implementation("org.springframework.security:spring-security-ldap:${property("springSecurityVersion")}")
    }
}

/**
 * Data Module
 * Database access and persistence layer
 */
project(":data-module") {
    apply(plugin = "org.springframework.boot")

    dependencies {
        implementation(project(":core-api"))

        // Spring Data
        implementation("org.springframework.boot:spring-boot-starter-data-jpa")
        implementation("org.springframework.boot:spring-boot-starter-data-rest")

        // Hibernate (vulnerable version)
        implementation("org.hibernate:hibernate-core:${property("hibernateVersion")}")
        implementation("org.hibernate:hibernate-validator:${property("hibernateVersion")}")

        // Connection pooling
        implementation("org.apache.commons:commons-dbcp2:2.9.0")

        // Liquibase for schema versioning
        implementation("org.liquibase:liquibase-core:4.9.1")
    }
}

/**
 * API Gateway Module
 * REST API and external integrations
 */
project(":api-gateway") {
    apply(plugin = "org.springframework.boot")

    dependencies {
        implementation(project(":core-api"))
        implementation(project(":security-module"))

        // Spring Cloud Gateway
        implementation("org.springframework.cloud:spring-cloud-starter-gateway")
        implementation("org.springframework.cloud:spring-cloud-starter-consul-discovery")

        // API Documentation
        implementation("org.springdoc:springdoc-openapi-ui:1.6.9")

        // HTTP Client (vulnerable version)
        implementation("org.apache.httpcomponents:httpclient:${property("commonsHttpClientVersion")}")
    }
}

/**
 * Monitoring Module
 * Metrics, logging, and health checks
 */
project(":monitoring-module") {
    apply(plugin = "org.springframework.boot")

    dependencies {
        // Spring Boot Actuator
        implementation("org.springframework.boot:spring-boot-starter-actuator")

        // Micrometer metrics
        implementation("io.micrometer:micrometer-registry-prometheus:1.9.1")

        // Logging (Log4j vulnerable version + fallback)
        implementation("org.apache.logging.log4j:log4j-api:${property("log4jVersion")}")
        implementation("org.apache.logging.log4j:log4j-core:${property("log4jCoreVersion")}")
        implementation("org.slf4j:slf4j-log4j12:${property("slf4jVersion")}")

        // Structured logging
        implementation("net.logstash.logback:logstash-logback-encoder:7.2")
    }
}

/**
 * Advanced Configurations using Platform/BOM
 */
configure(subprojects.filter { it.name in listOf("api-gateway", "data-module") }) {
    dependencies {
        // Google Cloud Platform integration
        implementation(platform("com.google.cloud:libraries-bom:${property("googleCloudBomVersion")}"))
        implementation("com.google.cloud:google-cloud-storage")
        implementation("com.google.cloud:google-cloud-pubsub")
    }
}

/**
 * Extended Dependency Configurations for Android modules (if applicable)
 */
configure(subprojects.filter { it.name.contains("android") }) {
    dependencies {
        debugImplementation("com.facebook.stetho:stetho:1.6.0")
        debugImplementation("com.facebook.stetho:stetho-okhttp3:1.6.0")

        releaseImplementation("com.google.firebase:firebase-crashlytics:18.0.0")
        releaseImplementation("com.google.firebase:firebase-analytics:21.1.1")

        // Code generation for Android
        ksp("com.google.dagger:dagger-compiler:2.42")
    }
}

/**
 * Root Project Tasks
 */
tasks {
    val buildInfo = register("buildInfo") {
        doLast {
            println("""
                ╔════════════════════════════════════════════════════════════════════╗
                ║              ENTERPRISE BUILD CONFIGURATION                        ║
                ║                                                                    ║
                ║  Project: ${project.group}                      ║
                ║  Version: ${project.version}                                       ║
                ║  Java: ${java.sourceCompatibility}                                               ║
                ║  Built: ${Instant.now()}     ║
                ║                                                                    ║
                ║  ⚠️  SECURITY NOTICE:                                              ║
                ║  This build contains known vulnerabilities for testing purposes   ║
                ║  DO NOT USE IN PRODUCTION without remediation                     ║
                ║                                                                    ║
                ╚════════════════════════════════════════════════════════════════════╝
            """.trimIndent())
        }
    }

    build {
        dependsOn(buildInfo)
    }
}

// Configure SonarQube analysis
sonarqube {
    properties {
        property("sonar.projectKey", "enterprise-platform")
        property("sonar.projectName", "Enterprise Platform")
        property("sonar.sources", "src/main")
        property("sonar.tests", "src/test")
        property("sonar.coverage.jacoco.xmlReportPaths", "**/target/site/jacoco/jacoco.xml")
    }
}
