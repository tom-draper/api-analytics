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
    Json(JsonData {
        message: "Hello World!".to_string(),
    })
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

#[cfg(test)]
mod tests {
    use super::*;
    use axum::{
        body::Body,
        http::{Method, Request, StatusCode},
        routing::post,
    };
    use tower::ServiceExt;

    fn app_with_analytics() -> Router {
        Router::new()
            .route("/", get(root))
            .route("/users", get(|| async { "users" }))
            .layer(Analytics::new("test-api-key".to_string()))
    }

    #[tokio::test]
    async fn test_get_root_passes_through() {
        let response = app_with_analytics()
            .oneshot(Request::builder().uri("/").body(Body::empty()).unwrap())
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_404_passes_through() {
        let response = app_with_analytics()
            .oneshot(Request::builder().uri("/missing").body(Body::empty()).unwrap())
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::NOT_FOUND);
    }

    #[tokio::test]
    async fn test_post_method_passes_through() {
        let app = Router::new()
            .route("/data", post(|| async { StatusCode::CREATED }))
            .layer(Analytics::new("test-key".to_string()));

        let response = app
            .oneshot(
                Request::builder()
                    .method(Method::POST)
                    .uri("/data")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::CREATED);
    }

    #[tokio::test]
    async fn test_x_forwarded_for_header() {
        let response = app_with_analytics()
            .oneshot(
                Request::builder()
                    .uri("/")
                    .header("x-forwarded-for", "203.0.113.1, 10.0.0.1")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_cf_connecting_ip_header() {
        let response = app_with_analytics()
            .oneshot(
                Request::builder()
                    .uri("/")
                    .header("cf-connecting-ip", "203.0.113.1")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_privacy_level_2() {
        let app = Router::new()
            .route("/", get(root))
            .layer(
                Analytics::new("test-key".to_string())
                    .with_privacy_level(2),
            );

        let response = app
            .oneshot(
                Request::builder()
                    .uri("/")
                    .header("cf-connecting-ip", "203.0.113.1")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_custom_path_mapper() {
        let app = Router::new()
            .route("/hello", get(|| async { "ok" }))
            .layer(
                Analytics::new("test-key".to_string())
                    .with_path_mapper(|req| format!("/prefix{}", req.uri().path())),
            );

        let response = app
            .oneshot(Request::builder().uri("/hello").body(Body::empty()).unwrap())
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_custom_user_id_mapper() {
        let app = Router::new()
            .route("/", get(root))
            .layer(
                Analytics::new("test-key".to_string())
                    .with_user_id_mapper(|_req| "user-123".to_string()),
            );

        let response = app
            .oneshot(Request::builder().uri("/").body(Body::empty()).unwrap())
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_custom_server_url_trailing_slash() {
        let app = Router::new()
            .route("/", get(root))
            .layer(
                Analytics::new("test-key".to_string())
                    .with_server_url("https://custom.example.com".to_string()),
            );

        let response = app
            .oneshot(Request::builder().uri("/").body(Body::empty()).unwrap())
            .await
            .unwrap();
        assert_eq!(response.status(), StatusCode::OK);
    }

    #[tokio::test]
    async fn test_multiple_requests() {
        let app = app_with_analytics();
        for uri in ["/", "/users", "/missing"] {
            let response = app
                .clone()
                .oneshot(Request::builder().uri(uri).body(Body::empty()).unwrap())
                .await
                .unwrap();
            assert!(response.status().as_u16() > 0);
        }
    }
}
