package dev.tomdraper.apianalytics.spring

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@Suppress("unused")
@SpringBootApplication
internal class DummyEntrypoint {
    fun main(args: Array<String>) {
        runApplication<DummyEntrypoint>(*args)
    }
}