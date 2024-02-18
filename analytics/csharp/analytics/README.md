# ASP.NET Core Analytics

A free and lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to https://apianalytics.dev/generate to generate your unique API key with a single click. This key is used to monitor your API server and should be stored privately. It's also required in order to view your API analytics dashboard and data.

### 2. Add middleware to your API

```sh
dotnet add package APIAnalytics.AspNetCore
```

```cs
using analytics;
using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

app.UseAnalytics(<API-KEY>); // Add middleware

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();
```

### 3. View your analytics

Your API will now log and store incoming request data on all valid routes. Your logged data can be viewed using two methods:

1. Through visualizations and statistics on our dashboard
2. Accessed directly via our data API

You can use the same API key across multiple APIs, but all your data will appear in the same dashboard. We recommend generating a new API key for each additional API server you want analytics for.

#### Dashboard

Head to https://apianalytics.dev/dashboard and paste in your API key to access your dashboard.

Demo: https://apianalytics.dev/dashboard/demo

![dashboard](https://user-images.githubusercontent.com/41476809/272061832-74ba4146-f4b3-4c05-b759-3946f4deb9de.png)

#### Data API

Logged data for all requests can be accessed via our REST API. Simply send a GET request to `https://apianalytics-server.com/api/data` with your API key set as `X-AUTH-TOKEN` in headers.

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

- `date` - specifies a particular day the requests occurred on (`YYYY-MM-DD`)
- `dateFrom` - specifies the lower bound of a date range the requests occurred in (`YYYY-MM-DD`)
- `dateTo` - specifies the upper bound of a date range the requests occurred in (`YYYY-MM-DD`)
- `ipAddress` - an IP address string of the client
- `status` - an integer status code of the response
- `location` - a two-character location code of the client

Example:

```bash
curl --header "X-AUTH-TOKEN: <API-KEY>" https://apianalytics-server.com/api/data?dateFrom=2022-01-01&dateTo=2022-06-01&status=200
```

## Client ID and Privacy

By default, API Analytics logs and stores the client IP address of all incoming requests made to your API and infers a country location from the IP address if possible. This IP address is used as a form of client identification in the dashboard to estimate the number of users accessing your service.

This behaviour can be controlled through a privacy level defined in the configuration of the API middleware. There are three privacy levels to choose from 0 (default) to a maximum of 2. A privacy level of 1 will disable IP address storing, and a value of 2 will also disable location inference.

1. Privacy level = 0 (default) - The client IP address is used to infer a location (country) and then stored for user identification of any given request.
2. Privacy level = 1 - The client IP address is used to infer a location (country) and then discarded.
3. Privacy level = 2 - The client IP address is never accessed and location is never inferred.

```cs
using analytics;
using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

var config = Config()
config.PrivacyLevel = 2

app.UseAnalytics(<API-KEY>, config); // Add middleware

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();
```

With any of these privacy levels, there is still the option to define a custom user ID as a function of a request by providing a mapper function in the API middleware configuration. For example, your service may require an API key sent in the `X-AUTH-TOKEN` header field that can be used to identify a user.

```cs
using analytics;
using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

var config = Config()
config.GetUserID = (context) => {
    if (context.user.Identity.IsAuthenticated)
        return context.user.Identity.Name;
    return "";
};

app.UseAnalytics(<API-KEY>, config); // Add middleware

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();
```

## Data and Security

All data is stored securely in compliance with The EU General Data Protection Regulation (GDPR).

For any given request to your API, data recorded is limited to:

- Path requested by client
- Client IP address
- Client operating system
- Client browser
- Request method (GET, POST, PUT, etc.)
- Time of request
- Status code
- Response time
- API hostname
- API framework (FastAPI, Flask, Express etc.)

Data collected is only ever used to populate your analytics dashboard. All stored data is pseudo-anonymous, with the API key the only link between you and your logged request data. Should you lose your API key, you will have no method to access your API analytics.

### Data Deletion

At any time you can delete all stored data associated with your API key by going to https://apianalytics.dev/delete and entering your API key.

API keys and their associated logged request data are scheduled to be deleted after 6 months of inactivity.

## Monitoring

Active API monitoring can be set up by heading to https://apianalytics.dev/monitoring to enter your API key. Our servers will regularly ping chosen API endpoints to monitor uptime and response time.
<!-- Optional email alerts when your endpoints are down can be subscribed to. -->

![Monitoring](https://user-images.githubusercontent.com/41476809/208298759-f937b668-2d86-43a2-b615-6b7f0b2bc20c.png)

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
