package dev.tomdraper.apianalytics.spring

import org.springframework.web.bind.annotation.GetMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@Suppress("unused")
@RestController
@RequestMapping("/spring")
internal class DummyRestController {

    @GetMapping
    fun helloWorld(): Map<String, String> = mapOf("hello" to "world")

}