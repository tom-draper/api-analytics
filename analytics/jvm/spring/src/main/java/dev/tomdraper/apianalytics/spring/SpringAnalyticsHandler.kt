package dev.tomdraper.apianalytics.spring

import com.fasterxml.jackson.databind.ObjectMapper
import dev.tomdraper.apianalytics.AbstractAnalyticsHandler
import dev.tomdraper.apianalytics.PayloadHandler.AnalyticsPayload
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.MediaType
import org.springframework.web.client.RestTemplate
import java.util.concurrent.Executors

class SpringAnalyticsHandler(
    override val apiKey: String?,
    override val loggingTimeout: Long,
    override val serverUrl: String,
    override val privacyLevel: Int
) : AbstractAnalyticsHandler<Unit, RestTemplate>(
    apiKey,
    RestTemplate(),
    loggingTimeout,
    serverUrl,
    "Spring",
    privacyLevel
) {
    //
    private val objectMapper = ObjectMapper()
    //
    private val executorService = Executors.newSingleThreadExecutor()

    override fun send(payload: AnalyticsPayload, endpoint: String) {
        val body = objectMapper.writeValueAsString(payload)
        executorService.submit {
            val headers = HttpHeaders()
            headers.contentType = MediaType.APPLICATION_JSON
            val request = HttpEntity(body, headers)
            client.postForEntity(endpoint, request, String::class.java)
        }
    }
}