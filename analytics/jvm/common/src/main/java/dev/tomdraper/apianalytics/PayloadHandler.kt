package dev.tomdraper.apianalytics

import com.fasterxml.jackson.annotation.JsonProperty
import java.text.SimpleDateFormat
import java.util.*

/** Universal payload handling, containing classes to be serialized and sent to the API. */
object PayloadHandler {

    /** Default server for the api endpoint(s). */
    const val DEFAULT_ENDPOINT = "https://www.apianalytics-server.com/"

    /** Universal helper function for calculating the right "created at" time.
     *
     * Getting [java.time.LocalDate] to parse as an ISO date was unnecessarily difficult,
     * so I wrote new logic that's platform agnostic and works with the API correctly. */
    @JvmStatic
    fun createdAt(): String {
        val tz = TimeZone.getTimeZone("UTC")
        val df = SimpleDateFormat("yyyy-MM-dd'T'HH:mm'Z'")
        df.timeZone = tz
        return df.format(Date())
    }

    data class RequestData(
        //
        val hostname: String,
        //
        @JsonProperty("ip_address") val ipAddress: String,
        //
        @JsonProperty("user_agent") val userAgent: String?,
        //
        val path: String,
        //
        @JsonProperty("status") val statusCode: Int?,
        //
        val method: String,
        //
        @JsonProperty("response_time") val responseTime: Long,
        //
        @JsonProperty("user_id") val userId: Int?,
        //
        @JsonProperty("created_at") val createdAt: String?
    )

    data class AnalyticsPayload(
        //
        @JsonProperty("api_key") val apiKey: String,
        //
        val requests: List<RequestData>,
        //
        val framework: String,
        //
        @JsonProperty("privacy_level") val privacyLevel: Int,
    )
}