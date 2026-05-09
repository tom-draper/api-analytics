plugins {
    alias(libs.plugins.kotlin)
    alias(libs.plugins.play)
    scala
}

group = "dev.tomdraper.apianalytics"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation(project(":common"))
    implementation(libs.jackson.scala)
    implementation(libs.scala.stdlib)
    implementation(libs.play.core)
    implementation(libs.play.webservice)

    testImplementation("org.scalactic:scalactic_3:3.2.19")
    testImplementation("org.scalatest:scalatest_3:3.2.19")
    testImplementation("org.scalatestplus.play:scalatestplus-play_2.13:5.1.0")
}
/*
play {
    platform {
        playVersion.set("2.8.20")
        scalaVersion.set("2.13")
        javaVersion.set(JavaVersion.VERSION_17)
    }
}
*/
