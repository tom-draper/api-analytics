package dev.tomdraper.apianalytics.play

import dev.tomdraper.apianalytics.PayloadHandler.{DEFAULT_ENDPOINT, RequestData, createdAt}
import play.api.ConfigLoader.stringLoader
import play.api.Configuration
import play.api.libs.ws.WSClient
import play.api.mvc._

import javax.inject.Inject
import scala.annotation.unused
import scala.concurrent.ExecutionContext

@unused
class PlayAnalyticsFilter @Inject()(
  wsClient: WSClient,
  config: Configuration
)(implicit val ec: ExecutionContext) extends EssentialFilter {

  private val apiKey: String = config.get[String]("apianalytics.api-key")
  private val timeout: Long = config.getOptional[Long]("apianalytics.timeout").getOrElse(60000L)
  private val privacyLevel: Int = config.getOptional[Int]("apianalytics.privacy-level").getOrElse(0)
  private val serverUrl: String = config.getOptional[String]("apianalytics.server-url").getOrElse(DEFAULT_ENDPOINT)

  private val apiHandler = new PlayAnalyticsHandler(
    apiKey,
    wsClient,
    timeout,
    serverUrl,
    privacyLevel
  )

  override def apply(next: EssentialAction): EssentialAction = EssentialAction { request =>
    val startTime = System.currentTimeMillis()

    next(request).map { result =>
      val endTime = System.currentTimeMillis()
      val elapsedTime = endTime - startTime

      apiHandler.logRequest(new RequestData(
        request.host,
        request.connection.remoteAddress.getHostAddress,
        request.headers.get("user-Agent").orNull,
        request.uri,
        result.header.status,
        request.method,
        elapsedTime,
        null,
        createdAt()
      ))

      result
    }
  }
}
