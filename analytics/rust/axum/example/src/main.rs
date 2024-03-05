use axum::{routing::get, Json, Router};
use axum_analytics::Analytics;
use dotenv::dotenv;
use serde::Serialize;
use std::net::SocketAddr;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

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
        .layer(Analytics::new(api_key));

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    println!("Server listening at: http://127.0.0.1:8080");
    axum::serve(listener, app).await.unwrap();
}
