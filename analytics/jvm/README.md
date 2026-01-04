# API Analytics

A free and lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to [apianalytics.dev/generate](https://apianalytics.dev/generate) to generate your unique API key with a single click. This key is used to monitor your specific API and should be stored privately. It's also required in order to access your API analytics dashboard and data.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there is minimal impact on the performance of your API.

#### Ktor

First, import it using Gradle;

```kotlin
repositories {
    mavenCentral() // This package and all it's dependencies are in Maven Central
}

dependencies {
    implementation("dev.tomdraper.apianalytics:ktor:1.0.0")
}
```

Lastly, install it. The middleware usually creates it's own `HttpClient`, but overwriting it is highly recommended in case you have specific settings you use that the default doesnt have.\
Though it should be noted that the example doesnt do anything, as the default one has the exact same settings.

```kotlin
// Imports simplified due to space constraints
import io.ktor.client.*
import io.ktor.server.*
import io.ktor.http.*
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation as ClientContentNegotiation
import io.ktor.server.plugins.contentnegotiation.ContentNegotiation as ServerContentNegotiation

private val myHttpClient: HttpClient = HttpClient(engineFactory = OkHttp) {
    install(ClientContentNegotiation) { jackson { } }
}

embeddedServer(factory = Netty) {
    install(ServerContentNegotiation) { jackson { } }
    install(AnalyticsPlugin) {
        apiKey = "[YOUR API KEY HERE]"
        httpClient = myHttpClient
    }
}
```

#### Spring

First is import it;

From Maven:

```xml
<dependency>
    <groupId>dev.tomdraper.apianalytics</groupId>
    <artifactId>spring</artifactId>
    <version>1.0.0</version>
</dependency>
```

Or from Gradle:

```kotlin
// While this is the kotlin specific format, it will still work (at least in modern Gradle)
implementation("dev.tomdraper.apianalytics:spring:1.0.0")
```

Next configure the plugin, either in;

`application.properties`:
```properties
apianalytics.apiKey = [YOUR API KEY HERE]
# Spring-only configs:
apianalytics.sendUserId = false # If a numerical user-id is found in the header name below, it is sent as well
apianalytics.userHeader = "" # The request header checked for numerical IDs
```

or `application.yml`:
```yml
apianalytics:
    apiKey = "[YOUR API KEY HERE]"
    sendUserId = false
    userHeader = ""
```

Then you're done, as the `@Component` annotation on the middleware will register it automatically.

#### PlayFramework

First import it in SBT:

```sbt
libraryDependencies += "dev.tomdraper.apianalytics" % "play" % "1.0.0"
```

Next, set up the `application.conf` file:

```conf
apianalytics {
    apiKey = [YOUR API KEY HERE]
}
```

Then add it to your project:

```scala
import dev.tomdraper.apianalytics.play.PlayAnalyticsFilter
import play.api.http.HttpFilters
import javax.inject.Inject

class MyFilters @Inject()(playAnalyticsFilter: PlayAnalyticsFilter) extends HttpFilters {
  def filters = Seq(playAnalyticsFilter)
}
```
