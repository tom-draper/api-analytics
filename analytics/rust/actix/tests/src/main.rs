use actix_web::{get, web, Responder, Result};
use serde::Serialize;
use dotenv::dotenv;
mod analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

#[get("/")]
async fn index() -> Result<impl Responder> {
    let json_data = JsonData {
        message: String::from("Hello World!"),
    };
    Ok(web::Json(json_data))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();

    use actix_web::{App, HttpServer};
    
    HttpServer::new(|| {
        let api_key = std::env::var("API_KEY").expect("API_KEY must be set.");

        App::new()
            .wrap(analytics::Analytics::new(api_key))
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
