plugins {
    alias(libs.plugins.kotlin)
}

group = "org.example"
version = "unspecified"

repositories {
    mavenCentral()
}

dependencies {
    implementation(libs.jackson.core)
    implementation(libs.jackson.kotlin)
}
