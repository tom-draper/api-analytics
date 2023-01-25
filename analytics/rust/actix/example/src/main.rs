use actix_web::{get, web, Responder, Result, App, HttpServer};
use serde::Serialize;
use dotenv::dotenv;
use actix_analytics::Analytics;

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

    println!("Server listening at: http://127.0.0.1:8080");
    HttpServer::new(|| {
        let api_key = std::env::var("API_KEY").expect("API_KEY must be set.");

        App::new().wrap(Analytics::new(api_key)).service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}

