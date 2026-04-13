// Project settings
name := "vulnerable-test-project"
version := "1.0.0"
scalaVersion := "2.13.12"

val jacksonVersion = "2.13.0"
lazy val log4jVersion = "2.14.0"
def strutsVersion = "2.5.20"

// Single dependency with % — CVE-2021-44228 (Log4Shell)
libraryDependencies += "org.apache.logging.log4j" % "log4j-core" % log4jVersion

// Single dependency with %% — safe dependency
libraryDependencies += "org.typelevel" %% "cats-core" % "2.9.0"

// Seq block with mixed operators and vulnerable packages
libraryDependencies ++= Seq(
  "com.fasterxml.jackson.core" % "jackson-databind" % jacksonVersion,
  "org.apache.struts" % "struts2-core" % strutsVersion,
  "commons-collections" % "commons-collections" % "3.2.1",
  "org.yaml" % "snakeyaml" % "1.26",
  "io.netty" %% "netty-codec-http" % "4.1.68.Final" % "test"
)

/*
  This is a block comment — dependencies here should NOT be parsed
  "org.example" % "should-not-parse" % "1.0.0"
*/

// Scala.js dependency with %%%
libraryDependencies += "org.scala-js" %%% "scalajs-dom" % "2.4.0"

// Dependency with exclude modifier
libraryDependencies += "org.apache.hadoop" % "hadoop-common" % "3.3.4" exclude("org.slf4j", "slf4j-log4j12")

// Dependency override
dependencyOverrides += "com.google.guava" % "guava" % "32.1.2-jre"
