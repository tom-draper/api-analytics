plugins {
    alias(libs.plugins.kotlin)
    //id("org.gradle.playframework") version "0.12" // Use the latest version
    scala
}

group = "dev.tomdraper.apianalytics"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation(project(":common"))
    implementation(libs.jackson.core)
    implementation(libs.jackson.scala)
    implementation("org.scala-lang:scala-library:2.13.12")
    implementation("com.typesafe.play:play_2.13:2.9.0") //2.8.20
    implementation("com.typesafe.play:play-ws_2.13:2.9.0")


    //testImplementation("org.scalatestplus.play:scalatestplus-play_2.13:5.1.0")
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
