#[macro_use]
extern crate rocket;
use dotenv::dotenv;
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
    dotenv().ok();
    let api_key = std::env::var("API_KEY").expect("API_KEY must be set.");
    rocket::build()
        .mount("/", routes![root])
        .attach(Analytics::new(api_key))
}
