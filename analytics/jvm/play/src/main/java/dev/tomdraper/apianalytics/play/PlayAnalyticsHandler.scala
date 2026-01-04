package dev.tomdraper.apianalytics.play

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

  override def send(payload: PayloadHandler.AnalyticsPayload, endpoint: String): Future[WSResponse] =
    httpClient.url(endpoint)
      .withHttpHeaders("Content-Type" -> "application/json")
      .post(getObjectMapper.writeValueAsString(payload))
}