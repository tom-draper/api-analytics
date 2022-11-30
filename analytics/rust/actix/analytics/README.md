# Actix Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
cargo add actix-analytics
```

```rust
use actix_web::{get, web, Responder, Result};
use serde::Serialize;
use actix_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

#[get("/")]
async fn index() -> Result<impl Responder> {
    let json_data = JsonData {
        message: "Hello World!".to_string(),
    };
    Ok(web::Json(json_data))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    use actix_web::{App, HttpServer};

    HttpServer::new(|| {
        App::new()
            .wrap(Analytics::new(<api_key>))
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.
