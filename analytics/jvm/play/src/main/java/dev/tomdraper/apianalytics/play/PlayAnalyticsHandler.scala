package dev.tomdraper.apianalytics.play

import com.fasterxml.jackson.databind.ObjectMapper
import dev.tomdraper.apianalytics.{AbstractAnalyticsHandler, PayloadHandler}
import play.api.libs.ws.{WSClient, WSResponse}

import scala.annotation.unused
import scala.concurrent.Future

@unused
class PlayAnalyticsHandler(
  apiKey: String,
  httpClient: WSClient,
  timeout: Long,
  serverUrl: String,
  privacyLevel: Int
) extends AbstractAnalyticsHandler[Future[WSResponse], WSClient](
  apiKey,
  httpClient,
  timeout,
  serverUrl,
  "Play",
  privacyLevel
) {

  private val objectMapper = new ObjectMapper()

  override def send(payload: PayloadHandler.AnalyticsPayload, endpoint: String): Future[WSResponse] =
    httpClient.url(endpoint)
      .withHttpHeaders("Content-Type" -> "application/json")
      .post(objectMapper.writeValueAsString(payload))
}