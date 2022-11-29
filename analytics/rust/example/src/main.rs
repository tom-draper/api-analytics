use actix_web::{get, web, Responder, Result};
use serde::Serialize;
mod analytics;

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
            .wrap(analytics::Analytics::new("hello".to_string()))
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
