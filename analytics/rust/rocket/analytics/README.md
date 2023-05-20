# Rocket Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to https://apianalytics.dev/generate to generate your unique API key with a single click. This key is used to monitor your API server and should be stored privately. It's also required in order to view your API analytics dashboard and data.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there is minimal impact on the performance of your API.

```bash
cargo add rocket-analytics
```

```rust
#[macro_use]
extern crate rocket;
use rocket::serde::json::Json;
use rocket_analytics::Analytics;
use serde::Serialize;

#[derive(Serialize)]
pub struct JsonData {
    message: String,
}

#[get("/")]
fn root() -> Json<JsonData> {
    let data = JsonData {
        message: "Hello World".to_string(),
    };
    Json(data)
}

#[launch]
fn rocket() -> _ {
    rocket::build()
        .mount("/", routes![root])
        .attach(Analytics::new(<API-KEY>))
}
```

### 3. View your analytics

Your API will now log and store incoming request data on all valid routes. Your logged data can be viewed using two methods:

1. Through visualizations and statistics on our dashboard
2. Accessed directly via our data API

You can use the same API key across multiple APIs, but all your data will appear in the same dashboard. We recommend generating a new API key for each additional API server you want analytics for.

#### Dashboard

Head to https://apianalytics.dev/dashboard and paste in your API key to access your dashboard.

Demo: https://apianalytics.dev/dashboard/demo

![Dashboard](https://user-images.githubusercontent.com/41476809/211800529-a84a0aa3-70c9-47d4-aa0d-7f9bbd3bc9b5.png)

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

Your data can be filtered by providing URL parameters in your request.

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

## Monitoring (coming soon)

Opt-in active API monitoring is coming soon. Our servers will regularly ping your API endpoints to monitor uptime and response time. Optional email alerts to notify you when your endpoints are down can be subscribed to.

![Monitoring](https://user-images.githubusercontent.com/41476809/208298759-f937b668-2d86-43a2-b615-6b7f0b2bc20c.png)

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

Data collected is only ever used to populate your analytics dashboard. All data stored is anonymous, with the API key the only link between you and your logged request data. Should you lose your API key, you will have no method to access your API analytics.

### Data Deletion

At any time, you can delete all stored data associated with your API key by going to https://apianalytics.dev/delete and entering your API key.

API keys and their associated API request data are scheduled to be deleted after 6 months of inactivity.

## Contributions

Contributions, issues and feature requests are welcome.

- Fork it (https://github.com/tom-draper/api-analytics)
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

