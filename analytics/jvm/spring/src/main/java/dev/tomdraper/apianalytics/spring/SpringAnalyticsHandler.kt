package dev.tomdraper.apianalytics.spring

import dev.tomdraper.apianalytics.AbstractAnalyticsHandler
import dev.tomdraper.apianalytics.PayloadHandler.AnalyticsPayload
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.MediaType
import org.springframework.web.client.RestTemplate

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
    "Express",//"Spring", // Spoofing Express temporarily until the backend can catch up
    privacyLevel
) {
    private val logger: Logger = LoggerFactory.getLogger(SpringAnalyticsHandler::class.java)

    override fun send(payload: AnalyticsPayload, endpoint: String) {
        logger.debug("Sending payload to analytics API...")
        val body = objectMapper.writeValueAsString(payload)
        val headers = HttpHeaders()
        headers.contentType = MediaType.APPLICATION_JSON
        val request = HttpEntity(body, headers)
        client.postForEntity(endpoint, request, String::class.java)
    }
}