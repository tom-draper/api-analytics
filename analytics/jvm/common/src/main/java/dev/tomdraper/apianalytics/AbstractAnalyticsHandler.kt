package dev.tomdraper.apianalytics

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

    //
    abstract fun send(payload: PayloadHandler.AnalyticsPayload, endpoint: String): R

    //
    fun logRequest(requestData: PayloadHandler.RequestData): R? {
        if (this.apiKey == null) return null

        this.requests += requestData
        if ((System.currentTimeMillis() - this.lastPosted) > this.loggingTimeout) {
            val endpoint = getServerEndpoint()

            return send(
                PayloadHandler.AnalyticsPayload(
                    apiKey!!,
                    requests,
                    frameworkName,
                    privacyLevel
                ),
                endpoint
            )
        }
        return null
    }
    //
    private fun getServerEndpoint(): String =
        if (serverUrl.endsWith("/")) "${serverUrl}api/log-request"
        else "$serverUrl/api/log-request"
}
