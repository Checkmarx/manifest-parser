name := "demo"
scalaVersion := "2.13.12"

val akkaVersion = "2.8.5"

libraryDependencies ++= Seq(
  "org.scala-lang" % "scala-library" % "2.13.12",
  "com.typesafe.akka" %% "akka-actor" % akkaVersion,
  "org.scalatest" %% "scalatest" % "3.2.18" % Test,
  "ch.qos.logback" % "logback-classic" % "1.4.14"
)
