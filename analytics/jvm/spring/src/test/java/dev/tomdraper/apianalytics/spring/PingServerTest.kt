package dev.tomdraper.apianalytics.spring

import org.junit.jupiter.api.Test
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.boot.test.web.server.LocalServerPort
import org.springframework.web.client.ResourceAccessException
import org.springframework.web.client.RestTemplate
import java.lang.Thread.sleep

/** Warning: this test says that it passed even though it could have failed during shutdown.
 * Please check this test manually until a proper fix can be found. */
@SpringBootTest(
    webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT,
    classes = [DummyEntrypoint::class],
    properties = ["apianalytics.apiKey=f9b678de-ddc7-48eb-91fc-d97e8e52c0a1"]
)
class PingServerTest {

    val logger: Logger = LoggerFactory.getLogger(PingServerTest::class.java)

    /** Primitives like [Int] cant be `lateinit`, use [Integer] instead. */
    @Suppress("PLATFORM_CLASS_MAPPED_TO_KOTLIN")
    @LocalServerPort
    lateinit var port: Integer

    var restTemp: RestTemplate = RestTemplate()

    @Test
    fun spamPing() {
        repeat(50) { index ->
            try {
                val res = restTemp.getForEntity("http://localhost:$port/spring", String::class.java)
                logger.info("Ping $index successful: ${res.body}")
            } catch (rae: ResourceAccessException) {
                logger.error(rae.message)
            }
        }
        sleep(5000)
    }
}