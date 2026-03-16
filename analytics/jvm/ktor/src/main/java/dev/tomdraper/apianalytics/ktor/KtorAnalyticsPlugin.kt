package dev.tomdraper.apianalytics.ktor

import dev.tomdraper.apianalytics.PayloadHandler
import dev.tomdraper.apianalytics.UniversalAnalyticsConfig
import io.ktor.client.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.statement.*
import io.ktor.serialization.jackson.*
import io.ktor.server.application.*
import io.ktor.server.application.hooks.*
import io.ktor.server.plugins.*
import io.ktor.server.request.*
import io.ktor.util.*
import io.ktor.util.logging.*

@Suppress("unused")
object KtorAnalyticsPlugin {
    private val LOG: Logger = KtorSimpleLogger("AnalyticsPlugin")

    val AnalyticsPlugin = createApplicationPlugin(
        name = "AnalyticsPlugin",
        createConfiguration = ::KtorAnalyticsConfig
    ) {
        val handler = KtorAnalyticsHandler(
            this.pluginConfig.apiKey,
            this.pluginConfig.httpClient,
            this.pluginConfig.timeout,
            this.pluginConfig.serverUrl,
            this.pluginConfig.privacyLevel,
        )
        val start = AttributeKey<Long>("AnalyticsPluginStartTime")

        onCallReceive { call, _ ->
            call.attributes.put(start, System.currentTimeMillis())
        }
        onCallRespond { call, _ ->
            var responseTime: Long = -1
            try {
                responseTime = System.currentTimeMillis() - call.attributes[start]
            } catch (resTimeErr: IllegalStateException) {
                LOG.error("Issue grabbing response time!")
            }

            // Short name so string templating in the debug log isn't a complete mess.
            val hr = handler.logRequest(
                PayloadHandler.RequestData(
                    call.request.host(),
                    call.request.origin.remoteAddress,
                    call.request.userAgent(),
                    call.request.path(),
                    call.response.status()?.value,
                    call.request.httpMethod.value,
                    responseTime,
                    null,
                    PayloadHandler.createdAt()
                )
            )

            if (hr != null)
                LOG.debug("""Request to ApiAnalytics (specifically ${hr.request.url.host}) 
                    |ran with status ${hr.status.value} / ${hr.status.description}""".trimMargin())
        }
        on(MonitoringEvent(ApplicationStopping)) {app ->
            println(handler.forceSend())
            app.monitor.unsubscribe(ApplicationStopping) { /* no-op, release resources */ }
        }
    }

    class KtorAnalyticsConfig : UniversalAnalyticsConfig() {
        /** ApiAnalytics by default uses its own [HttpClient] and Jackson for serde.
         * Overriding this with your preferred Ktor HttpClient is recommended, but completely optional. */
        var httpClient: HttpClient = HttpClient { install(ContentNegotiation) { jackson {  } } }
    }
}