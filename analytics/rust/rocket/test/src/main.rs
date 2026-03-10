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
    Json(JsonData {
        message: "Hello World".to_string(),
    })
}

#[get("/users")]
fn users() -> Json<JsonData> {
    Json(JsonData {
        message: "Users".to_string(),
    })
}

#[launch]
fn rocket() -> _ {
    dotenv().ok();
    let api_key = std::env::var("API_KEY").expect("API_KEY must be set.");
    rocket::build()
        .mount("/", routes![root, users])
        .attach(Analytics::new(api_key))
}

#[cfg(test)]
mod tests {
    use super::*;
    use rocket::{http::Status, local::asynchronous::Client};
    use rocket_analytics::Analytics;

    async fn test_client() -> Client {
        let rocket = rocket::build()
            .mount("/", routes![root, users])
            .attach(Analytics::new("test-api-key".to_string()));
        Client::tracked(rocket).await.unwrap()
    }

    #[rocket::async_test]
    async fn test_get_root_passes_through() {
        let client = test_client().await;
        let response = client.get("/").dispatch().await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_404_passes_through() {
        let client = test_client().await;
        let response = client.get("/missing").dispatch().await;
        assert_eq!(response.status(), Status::NotFound);
    }

    #[rocket::async_test]
    async fn test_get_users_passes_through() {
        let client = test_client().await;
        let response = client.get("/users").dispatch().await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_x_forwarded_for_header() {
        let client = test_client().await;
        let response = client
            .get("/")
            .header(rocket::http::Header::new("x-forwarded-for", "203.0.113.1, 10.0.0.1"))
            .dispatch()
            .await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_cf_connecting_ip_header() {
        let client = test_client().await;
        let response = client
            .get("/")
            .header(rocket::http::Header::new("cf-connecting-ip", "203.0.113.1"))
            .dispatch()
            .await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_privacy_level_2() {
        let rocket = rocket::build()
            .mount("/", routes![root])
            .attach(Analytics::new("test-key".to_string()).with_privacy_level(2));
        let client = Client::tracked(rocket).await.unwrap();

        let response = client
            .get("/")
            .header(rocket::http::Header::new("cf-connecting-ip", "203.0.113.1"))
            .dispatch()
            .await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_custom_path_mapper() {
        let rocket = rocket::build()
            .mount("/", routes![root])
            .attach(
                Analytics::new("test-key".to_string())
                    .with_path_mapper(|req| format!("/prefix{}", req.uri().path())),
            );
        let client = Client::tracked(rocket).await.unwrap();

        let response = client.get("/").dispatch().await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_custom_user_id_mapper() {
        let rocket = rocket::build()
            .mount("/", routes![root])
            .attach(
                Analytics::new("test-key".to_string())
                    .with_user_id_mapper(|_req| "user-123".to_string()),
            );
        let client = Client::tracked(rocket).await.unwrap();

        let response = client.get("/").dispatch().await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_custom_server_url() {
        let rocket = rocket::build()
            .mount("/", routes![root])
            .attach(
                Analytics::new("test-key".to_string())
                    .with_server_url("https://custom.example.com".to_string()),
            );
        let client = Client::tracked(rocket).await.unwrap();

        let response = client.get("/").dispatch().await;
        assert_eq!(response.status(), Status::Ok);
    }

    #[rocket::async_test]
    async fn test_multiple_requests() {
        let client = test_client().await;
        let cases = [("/", Status::Ok), ("/users", Status::Ok), ("/missing", Status::NotFound)];
        for (uri, expected_status) in cases {
            let response = client.get(uri).dispatch().await;
            assert_eq!(response.status(), expected_status);
        }
    }
}
