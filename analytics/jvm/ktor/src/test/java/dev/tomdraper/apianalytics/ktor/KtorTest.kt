package dev.tomdraper.apianalytics.ktor

import dev.tomdraper.apianalytics.ktor.KtorAnalyticsPlugin.AnalyticsPlugin
import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.engine.okhttp.*
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.jackson.*
import io.ktor.server.application.*
import io.ktor.server.engine.*
import io.ktor.server.netty.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation as ClientContentNegotiation
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation as ServerContentNegotiation

object KtorTest {

    @Test
    fun spamEndpoint() {
        repeat(50) {iteration ->
            logger.info("$iteration")
            runBlocking {
                val res = testClient.get("localhost:80")
                logger.info("${res.status.value} ${res.status.description} ${res.body<String>()}")
            }
        }
    }

    private val logger: Logger = LoggerFactory.getLogger(KtorTest::class.java)

    private val testClient: HttpClient = HttpClient(engineFactory = OkHttp) {
        install(ClientContentNegotiation) { jackson { } }
    }

    @JvmStatic
    @BeforeAll
    fun startHttpServer() {
        logger.info("Starting server..")
        embeddedServer(factory = Netty) {
            logger.info("Installing plugins...")
            install(ServerContentNegotiation) { jackson { } }
            install(AnalyticsPlugin) {
                apiKey = "f9b678de-ddc7-48eb-91fc-d97e8e52c0a1"
                timeout = 10000
                httpClient = testClient
            }

            routing {
                get("/") {
                    call.respond(HttpStatusCode.OK, mapOf("hello" to "world"))
                }
            }
        }.start(wait = true)
    }
}