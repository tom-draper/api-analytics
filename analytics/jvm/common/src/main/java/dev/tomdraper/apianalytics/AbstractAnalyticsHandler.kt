package dev.tomdraper.apianalytics

import com.fasterxml.jackson.databind.ObjectMapper

abstract class AbstractAnalyticsHandler<R, C>(
    open val apiKey: String?,
    open val client: C,
    open val loggingTimeout: Long,
    open val serverUrl: String,
    private val frameworkName: String,
    open val privacyLevel: Int,
) {
    //
    private val requests: MutableList<PayloadHandler.RequestData> = emptyList<PayloadHandler.RequestData>().toMutableList()
    //
    private var lastPosted: Long = System.currentTimeMillis()
    /** Universal [ObjectMapper] for implementations to use instead of creating their own. */
    protected val objectMapper = ObjectMapper()

    //
    abstract fun send(payload: PayloadHandler.AnalyticsPayload, endpoint: String): R

    //
    fun logRequest(requestData: PayloadHandler.RequestData): R? {
        this.requests += requestData

        return if ((System.currentTimeMillis() - this.lastPosted) > this.loggingTimeout) forceSend()
        else null
    }

    /** Force-sends a request to the API, ignoring any usage limits.
     * Usually only called during shutdown or panic, but also used in [logRequest].
     * This function should be used *very* carefully. */
    fun forceSend(): R? {
        if (this.apiKey == null) return null

        val endpoint = getServerEndpoint()
        val payload = PayloadHandler.AnalyticsPayload(
            apiKey!!,
            requests,
            frameworkName,
            privacyLevel
        )
        println(objectMapper.writeValueAsString(payload))

        return send(payload, endpoint)
    }

    //
    private fun getServerEndpoint(): String =
        if (serverUrl.endsWith("/")) "${serverUrl}api/log-request"
        else "$serverUrl/api/log-request"
}
