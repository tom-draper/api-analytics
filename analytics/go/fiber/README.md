# Fiber Analytics

A free and lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to [apianalytics.dev/generate](https://apianalytics.dev/generate) to generate your unique API key with a single click. This key is used to monitor your specific API and should be stored privately. It's also required in order to access your API analytics dashboard and data.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there is minimal impact on the performance of your API.

[![Fiber](https://img.shields.io/badge/go.mod-Fiber-blue)](https://github.com/tom-draper/api-analytics/tree/main/analytics/go/fiber)

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/fiber
```

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"
)

func root(c *fiber.Ctx) error {
    data := map[string]string{
        "message": "Hello, World!",
    }
    return c.JSON(data)
}

func main() {
	app := fiber.New()

	app.Use(analytics.Analytics(<API-KEY>)) // Add middleware

	app.Get("/", root)
	app.Listen(":8080")
}
```

### 3. View your analytics

Your API will now log and store incoming request data on all routes. Logged data can be viewed using two methods:

1. Through visualizations and statistics on the dashboard
2. Accessed directly via the data API

You can use the same API key across multiple APIs, but all requests will appear in the same dashboard. It's recommended to generate a new API key for each of your API servers.

#### Dashboard

Head to [apianalytics.dev/dashboard](https://apianalytics.dev/dashboard) and paste in your API key to access your dashboard.

Demo: [apianalytics.dev/dashboard/demo](https://apianalytics.dev/dashboard/demo)

![dashboard](https://user-images.githubusercontent.com/41476809/272061832-74ba4146-f4b3-4c05-b759-3946f4deb9de.png)

#### Data API

Raw logged request data can be fetched from the data API. Simply send a GET request to `https://apianalytics-server.com/api/data` with your API key set as `X-AUTH-TOKEN` in the headers.

##### Python

```py
import requests

headers = {
    "X-AUTH-TOKEN": <API-KEY>
}

response = requests.get("https://apianalytics-server.com/api/data", headers=headers)
print(response.json())
```

##### Node.js

```js
fetch("https://apianalytics-server.com/api/data", {
    headers: { "X-AUTH-TOKEN": <API-KEY> },
})
    .then((response) => {
        return response.json();
    })
    .then((data) => {
        console.log(data);
    });
```

##### cURL

```bash
curl --header "X-AUTH-TOKEN: <API-KEY>" https://apianalytics-server.com/api/data
```

##### Parameters

You can filter your data by providing URL parameters in your request.

- `page` - the page number, with a max page size of 50,000 (defaults to 1)
- `date` - the day the requests occurred on (`YYYY-MM-DD`)
- `dateFrom` - a lower bound of a date range the requests occurred in (`YYYY-MM-DD`)
- `dateTo` - a upper bound of a date range the requests occurred in (`YYYY-MM-DD`)
- `hostname` - the hostname of your service
- `ipAddress` - the IP address of the client
- `status` - the status code of the response
- `location` - a two-character location code of the client
- `user_id` - a custom user identifier (only relevant if a `GetUserID` mapper function has been set within config)

Example:

```bash
curl --header "X-AUTH-TOKEN: <API-KEY>" https://apianalytics-server.com/api/data?page=3&dateFrom=2022-01-01&hostname=apianalytics.dev&status=200&user_id=b56cbd92-1168-4d7b-8d94-0418da207908
```

## Customisation

Custom mapping functions can be assigned to override the default behaviour and define how values are extracted from each incoming request to better suit your specific API.

```go
package main

import (
	analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	config := analytics.NewConfig()
	config.GetIPAddress = func(c *fiber.Ctx) string {
		return c.IP()
	}
	config.GetUserAgent = func(c *fiber.Ctx) string {
		return string(c.Request().Header.UserAgent())
	}
	app.Use(analytics.AnalyticsWithConfig(<API-KEY>, config)) // Add middleware
}
```

## Client ID and Privacy

By default, API Analytics logs and stores the client IP address of all incoming requests made to your API and infers a location (country) from each IP address if possible. The IP address is used as a form of client identification in the dashboard to estimate the number of users accessing your service.

This behaviour can be controlled through a privacy level defined in the configuration of the API middleware. There are three privacy levels to choose from 0 (default) to a maximum of 2. A privacy level of 1 will disable IP address storing, and a value of 2 will also disable location inference.

Privacy Levels:

- `0` - The client IP address is used to infer a location and then stored for user identification. (default)
- `1` - The client IP address is used to infer a location and then discarded.
- `2` - The client IP address is never accessed and location is never inferred.

```go
config := analytics.NewConfig()
config.PrivacyLevel = 2 // Disable IP storing and location inference
```

With any of these privacy levels, you have the option to define a custom user ID as a function of a request by providing a mapper function in the API middleware configuration. For example, your service may require an API key held in the `X-AUTH-TOKEN` header field which is used to identify a user of your service. In the dashboard, this custom user ID will identify the user in conjunction with the IP address or as an alternative depending on the privacy level set.

```py
config := analytics.NewConfig()
config.GetUserID = func(c *fiber.Ctx) string {
	return c.Get("X-AUTH-TOKEN")
}
```

## Data and Security

All data is stored securely, and in compliance with The EU General Data Protection Regulation (GDPR).

For any given request to your API, data recorded is strictly limited to:

- Request method (GET, POST, PUT, etc.)
- Endpoint requested
- User agent
- Client IP address (optional)
- Timestamp of the request
- Response status code
- Response time
- Hostname of API
- API framework in use (FastAPI, Flask, Express, etc.)

Data collected is only ever used to populate your analytics dashboard, and never shared with a third-party. All stored data is pseudo-anonymous, with the API key the only link between you and your logged request data. Should you lose your API key, you will have no method to access your API analytics.

View our full <a href="https://www.apianalytics.dev/privacy-policy">privacy policy</a> and <a href="https://www.apianalytics.dev/faq">frequently asked questions</a> on our website.

### Data Deletion

At any time, you can delete all stored data associated with your API key by going to [apianalytics.dev/delete](https://apianalytics.dev/delete) and entering your API key.

API keys and their associated logged request data are scheduled to be deleted after 6 months of inactivity, or 3 months have elapsed without logging a request.

## Monitoring

Active API monitoring can be set up by heading to [apianalytics.dev/monitoring](https://apianalytics.dev/monitoring) to enter your API key. Our servers will regularly ping chosen endpoints to monitor uptime and response time. 

![Monitoring](https://user-images.githubusercontent.com/41476809/208298759-f937b668-2d86-43a2-b615-6b7f0b2bc20c.png)

## Limitations

In order to keep the service free, up to 1.5 million requests can be stored against an API key. This is enforced as a rolling limit; old requests will be replaced by new requests. If your API would rapidly exceed this limit, we recommend you try other solutions or check out [self-hosting](./server/self-hosting/README.md).

## Self-Hosting

The project can be self-hosted by following the [guide](./server/self-hosting/README.md).

Please note: Self-hosting is still undergoing testing, development and further improvements to make it as easy as possible to deploy. It is currently recommended that you avoid self-hosting for production use.

## Contributions

Contributions, issues and feature requests are welcome.

- Fork it (https://github.com/tom-draper/api-analytics)
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

---

If you find value in my work consider supporting me.

Buy Me a Coffee: https://www.buymeacoffee.com/tomdraper<br>
PayPal: https://www.paypal.com/paypalme/tomdraper

