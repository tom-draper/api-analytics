use axum::{
    routing::get,
    Json, Router,
};
use serde::Serialize;
use std::net::SocketAddr;
use tokio;
use dotenv::dotenv;
mod analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

// basic handler that responds with a static string
async fn root() -> Json<JsonData> {
    let json_data = JsonData {
        message: "Hello World!".to_string(),
    };
    Json(json_data)
}

#[tokio::main]
async fn main() {
    dotenv().ok();
    let api_key = std::env::var("API_KEY").expect("API_KEY must be set.");

    let app = Router::new()
        .route("/", get(root))
        .layer(analytics::Analytics::new(api_key));

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}