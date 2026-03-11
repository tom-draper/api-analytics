use actix_analytics::Analytics;
use actix_web::{get, web, App, HttpServer, Responder, Result};
use dotenv::dotenv;
use serde::Serialize;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

#[get("/")]
async fn index() -> Result<impl Responder> {
    Ok(web::Json(JsonData {
        message: String::from("Hello World!"),
    }))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv().ok();
    let api_key = std::env::var("API_KEY").expect("API_KEY must be set.");

    println!("Server listening at: http://127.0.0.1:8080");
    HttpServer::new(move || {
        App::new()
            .wrap(Analytics::new(api_key.clone()))
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}

#[cfg(test)]
mod tests {
    use super::*;
    use actix_web::{http::StatusCode, test, App};

    #[actix_web::test]
    async fn test_get_root_passes_through() {
        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()))
                .service(index),
        )
        .await;

        let req = test::TestRequest::get().uri("/").to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_404_passes_through() {
        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()))
                .service(index),
        )
        .await;

        let req = test::TestRequest::get().uri("/missing").to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::NOT_FOUND);
    }

    #[actix_web::test]
    async fn test_post_method_passes_through() {
        use actix_web::{post, HttpResponse};

        #[post("/data")]
        async fn post_data() -> HttpResponse {
            HttpResponse::Created().finish()
        }

        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()))
                .service(post_data),
        )
        .await;

        let req = test::TestRequest::post().uri("/data").to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::CREATED);
    }

    #[actix_web::test]
    async fn test_x_forwarded_for_header() {
        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()))
                .service(index),
        )
        .await;

        let req = test::TestRequest::get()
            .uri("/")
            .insert_header(("x-forwarded-for", "203.0.113.1, 10.0.0.1"))
            .to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_cf_connecting_ip_header() {
        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()))
                .service(index),
        )
        .await;

        let req = test::TestRequest::get()
            .uri("/")
            .insert_header(("cf-connecting-ip", "203.0.113.1"))
            .to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_privacy_level_2() {
        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()).with_privacy_level(2))
                .service(index),
        )
        .await;

        let req = test::TestRequest::get()
            .uri("/")
            .insert_header(("cf-connecting-ip", "203.0.113.1"))
            .to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_custom_path_mapper() {
        use actix_web::{get, HttpResponse};

        #[get("/hello")]
        async fn hello() -> HttpResponse {
            HttpResponse::Ok().finish()
        }

        let app = test::init_service(
            App::new()
                .wrap(
                    Analytics::new("test-key".to_string())
                        .with_path_mapper(|req| format!("/prefix{}", req.path())),
                )
                .service(hello),
        )
        .await;

        let req = test::TestRequest::get().uri("/hello").to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_custom_user_id_mapper() {
        let app = test::init_service(
            App::new()
                .wrap(
                    Analytics::new("test-key".to_string())
                        .with_user_id_mapper(|_req| "user-123".to_string()),
                )
                .service(index),
        )
        .await;

        let req = test::TestRequest::get().uri("/").to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_custom_server_url() {
        let app = test::init_service(
            App::new()
                .wrap(
                    Analytics::new("test-key".to_string())
                        .with_server_url("https://custom.example.com".to_string()),
                )
                .service(index),
        )
        .await;

        let req = test::TestRequest::get().uri("/").to_request();
        let resp = test::call_service(&app, req).await;
        assert_eq!(resp.status(), StatusCode::OK);
    }

    #[actix_web::test]
    async fn test_multiple_requests() {
        let app = test::init_service(
            App::new()
                .wrap(Analytics::new("test-key".to_string()))
                .service(index),
        )
        .await;

        for (method, uri, expected) in [
            ("GET", "/", 200u16),
            ("GET", "/missing", 404),
        ] {
            let req = test::TestRequest::with_uri(uri)
                .method(actix_web::http::Method::from_bytes(method.as_bytes()).unwrap())
                .to_request();
            let resp = test::call_service(&app, req).await;
            assert_eq!(resp.status().as_u16(), expected);
        }
    }
}
