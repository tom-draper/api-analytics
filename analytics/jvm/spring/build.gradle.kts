plugins {
    alias(libs.plugins.kotlin)
    alias(libs.plugins.spring.kotlin)
    alias(libs.plugins.spring.boot)
    id("io.spring.dependency-management")
    application
}

group = "dev.tomdraper.apianalytics"
version = "1.0.0"

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(17)
    }
}

application {
    mainClass.set("dev.tomdraper.apianalytics.spring.DummyEntrypoint")
}

repositories {
    mavenCentral()
}

dependencies {
    implementation(project(":common"))
    implementation("org.springframework.boot:spring-boot-starter-web")
    implementation("org.jetbrains.kotlin:kotlin-reflect")
    compileOnly(libs.lombok)
    annotationProcessor(libs.lombok)

    testImplementation("org.springframework.boot:spring-boot-starter-test")
    testImplementation(libs.logging.logback)
    testImplementation(libs.jackson.core)
    testImplementation(libs.junit.kotlin)
    testImplementation(libs.junit.api)
    testRuntimeOnly(libs.junit.engine)
}
kotlin {
    compilerOptions {
        freeCompilerArgs.addAll("-Xjsr305=strict")
    }
}

tasks.withType<Test> {
    useJUnitPlatform()
}

tasks.withType<Copy> {
    duplicatesStrategy = DuplicatesStrategy.EXCLUDE
}
