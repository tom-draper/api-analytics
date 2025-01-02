plugins {
    alias(libs.plugins.kotlin)
    alias(libs.plugins.ktor)
}

group = "dev.tomdraper.apianalytics"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    implementation(project(":common"))
    implementation(libs.ktor.server.core)
    implementation(libs.ktor.client.core)
    implementation(libs.ktor.client.content)
    implementation(libs.ktor.serialization.jackson)

    testImplementation(libs.ktor.server.content)
    testImplementation(libs.ktor.server.netty)
    testImplementation(libs.ktor.client.okhttp)
    testImplementation(libs.logging.logback)
    testImplementation(libs.junit.kotlin)
    testImplementation(libs.junit.api)
    testRuntimeOnly(libs.junit.engine)
}

application {
    mainClass.set("dev.tomdraper.apianalytics.ktor.DummyKt")
}

/*tasks.named<Test>("test") {
    useJUnitPlatform()
}*/
