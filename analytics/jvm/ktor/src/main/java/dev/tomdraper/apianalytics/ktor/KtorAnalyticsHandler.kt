package dev.tomdraper.apianalytics.ktor

import dev.tomdraper.apianalytics.AbstractAnalyticsHandler
import dev.tomdraper.apianalytics.PayloadHandler
import io.ktor.client.*
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import kotlinx.coroutines.runBlocking

class KtorAnalyticsHandler(
    override val apiKey: String?,
    override val client: HttpClient,
    override val loggingTimeout: Long,
    override val serverUrl: String,
    override val privacyLevel: Int
) : AbstractAnalyticsHandler<HttpResponse, HttpClient>(
    apiKey,
    client,
    loggingTimeout,
    serverUrl,
    "Ktor",
    privacyLevel
) {
    override fun send(payload: PayloadHandler.AnalyticsPayload, endpoint: String): HttpResponse = runBlocking {
        return@runBlocking client.post(endpoint) {
            contentType(ContentType.Application.Json)
            setBody(payload)
        }
    }
}