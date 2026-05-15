// SBT plugins
addSbtPlugin("com.eed3si9n" % "sbt-assembly" % "2.1.0")
addSbtPlugin("org.scalameta" % "sbt-scalafmt" % "2.5.2")
addSbtPlugin("com.github.sbt" % "sbt-native-packager" % "1.9.16")

// Vulnerable dependencies for testing (intentional - to verify IDE decorations)
addSbtPlugin("org.apache.log4j" % "log4j-core" % "2.14.1")
addSbtPlugin("org.apache.commons" % "commons-compress" % "1.20")
addSbtPlugin("commons-io" % "commons-io" % "2.4")
