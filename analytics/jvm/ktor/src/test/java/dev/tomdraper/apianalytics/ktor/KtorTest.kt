package dev.tomdraper.apianalytics.ktor

import dev.tomdraper.apianalytics.ktor.KtorAnalyticsPlugin.AnalyticsPlugin
import io.ktor.client.call.*
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.jackson.*
import io.ktor.server.application.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import io.ktor.server.testing.*
import org.junit.jupiter.api.Test
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation as ClientContentNegotiation
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation as ServerContentNegotiation

object KtorTest {

    private val logger: Logger = LoggerFactory.getLogger(KtorTest::class.java)

    @Test
    fun spamEndpoint() = testApplication {

        application {
            install(ServerContentNegotiation) { jackson { } }
            install(AnalyticsPlugin) {
                apiKey = "f9b678de-ddc7-48eb-91fc-d97e8e52c0a1"
                timeout = 10000
            }

            routing {
                get("/") {
                    call.respond(HttpStatusCode.OK, mapOf("hello" to "world"))
                }
            }
        }

        val client = createClient {
            install(ClientContentNegotiation) { jackson { } }
        }

        repeat(50) {iteration ->
            logger.info("$iteration")
            val res = client.get("/")
            logger.info("${res.status.value} ${res.status.description} ${res.body<String>()}")
        }
    }
}