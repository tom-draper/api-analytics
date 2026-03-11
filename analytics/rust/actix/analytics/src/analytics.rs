use actix_rt::spawn;
use actix_web::{
    dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform},
    http::header::{HeaderValue, HOST, USER_AGENT},
    Error,
};
use chrono::Utc;
use futures::future::LocalBoxFuture;
use reqwest::Client;
use serde::Serialize;
use std::sync::{Arc, Mutex};
use std::{
    future::{ready, Ready},
    time::{Duration, Instant},
};

#[derive(Debug, Clone, Serialize)]
struct RequestData {
    hostname: String,
    ip_address: Option<String>,
    path: String,
    user_agent: String,
    method: String,
    response_time: u32,
    status: u16,
    user_id: Option<String>,
    created_at: String,
}

type StringMapper = dyn for<'a> Fn(&ServiceRequest) -> String + Send + Sync;

#[derive(Clone)]
struct Config {
    privacy_level: i32,
    server_url: String,
    get_hostname: Arc<StringMapper>,
    get_ip_address: Arc<StringMapper>,
    get_path: Arc<StringMapper>,
    get_user_agent: Arc<StringMapper>,
    get_user_id: Arc<StringMapper>,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            privacy_level: 0,
            server_url: String::from("https://www.apianalytics-server.com/"),
            get_hostname: Arc::new(get_hostname),
            get_ip_address: Arc::new(get_ip_address),
            get_path: Arc::new(get_path),
            get_user_agent: Arc::new(get_user_agent),
            get_user_id: Arc::new(get_user_id),
        }
    }
}

pub trait HeaderValueExt {
    fn to_string(&self) -> String;
}

impl HeaderValueExt for HeaderValue {
    fn to_string(&self) -> String {
        self.to_str().unwrap_or_default().to_string()
    }
}

fn get_hostname(req: &ServiceRequest) -> String {
    req.headers()
        .get(HOST)
        .map(|x| x.to_string())
        .unwrap_or_default()
}

fn get_ip_address(req: &ServiceRequest) -> String {
    let headers = req.headers();

    if let Some(val) = headers
        .get("cf-connecting-ip")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.trim().to_owned())
        .filter(|s| !s.is_empty())
    {
        return val;
    }

    if let Some(val) = headers
        .get("x-forwarded-for")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.split(',').next())
        .map(|s| s.trim().to_owned())
        .filter(|s| !s.is_empty())
    {
        return val;
    }

    if let Some(val) = headers
        .get("x-real-ip")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.trim().to_owned())
        .filter(|s| !s.is_empty())
    {
        return val;
    }

    req.peer_addr()
        .map(|addr| addr.ip().to_string())
        .unwrap_or_default()
}

fn get_path(req: &ServiceRequest) -> String {
    req.path().to_string()
}

fn get_user_agent(req: &ServiceRequest) -> String {
    req.headers()
        .get(USER_AGENT)
        .map(|x| x.to_string())
        .unwrap_or_default()
}

fn get_user_id(_req: &ServiceRequest) -> String {
    String::new()
}

struct RequestBuffer {
    requests: Vec<RequestData>,
    last_posted: Instant,
}

impl RequestBuffer {
    fn new() -> Self {
        Self {
            requests: Vec::new(),
            last_posted: Instant::now(),
        }
    }
}

pub struct Analytics {
    api_key: String,
    config: Config,
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        Self {
            api_key,
            config: Config::default(),
        }
    }

    pub fn with_privacy_level(mut self, privacy_level: i32) -> Self {
        self.config.privacy_level = privacy_level;
        self
    }

    pub fn with_server_url(mut self, server_url: String) -> Self {
        self.config.server_url = if server_url.ends_with('/') {
            server_url
        } else {
            server_url + "/"
        };
        self
    }

    pub fn with_hostname_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_hostname = Arc::new(mapper);
        self
    }

    pub fn with_ip_address_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_ip_address = Arc::new(mapper);
        self
    }

    pub fn with_path_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_path = Arc::new(mapper);
        self
    }

    pub fn with_user_agent_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_user_agent = Arc::new(mapper);
        self
    }

    pub fn with_user_id_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_user_id = Arc::new(mapper);
        self
    }
}

impl<S, B> Transform<S, ServiceRequest> for Analytics
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type InitError = ();
    type Transform = AnalyticsMiddleware<S>;
    type Future = Ready<Result<Self::Transform, Self::InitError>>;

    fn new_transform(&self, service: S) -> Self::Future {
        let client = Client::builder()
            .timeout(Duration::from_secs(10))
            .build()
            .unwrap_or_else(|_| Client::new());

        ready(Ok(AnalyticsMiddleware {
            api_key: Arc::new(self.api_key.clone()),
            config: Arc::new(self.config.clone()),
            buffer: Arc::new(Mutex::new(RequestBuffer::new())),
            client: Arc::new(client),
            service,
        }))
    }
}

pub struct AnalyticsMiddleware<S> {
    api_key: Arc<String>,
    config: Arc<Config>,
    buffer: Arc<Mutex<RequestBuffer>>,
    client: Arc<Client>,
    service: S,
}

#[derive(Debug, Clone, Serialize)]
struct Payload {
    api_key: String,
    requests: Vec<RequestData>,
    framework: String,
    privacy_level: i32,
}

impl Payload {
    pub fn new(api_key: String, requests: Vec<RequestData>, privacy_level: i32) -> Self {
        Self {
            api_key,
            requests,
            framework: String::from("Actix"),
            privacy_level,
        }
    }
}

async fn post_requests(client: &Client, data: Payload, server_url: &str) {
    let url = format!("{}api/log-request", server_url);
    let _ = client.post(url).json(&data).send().await;
}

fn log_request(
    buffer: &Arc<Mutex<RequestBuffer>>,
    client: &Arc<Client>,
    api_key: &str,
    request_data: RequestData,
    config: &Config,
) {
    let batch = {
        let mut buf = buffer.lock().unwrap();
        buf.requests.push(request_data);
        if buf.last_posted.elapsed().as_secs_f64() > 60.0 {
            buf.last_posted = Instant::now();
            std::mem::take(&mut buf.requests)
        } else {
            vec![]
        }
    };

    if !batch.is_empty() {
        let payload = Payload::new(api_key.to_owned(), batch, config.privacy_level);
        let server_url = config.server_url.clone();
        let client = Arc::clone(client);
        spawn(async move {
            post_requests(&client, payload, &server_url).await;
        });
    }
}

impl<S, B> Service<ServiceRequest> for AnalyticsMiddleware<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    type Response = ServiceResponse<B>;
    type Error = Error;
    type Future = LocalBoxFuture<'static, Result<Self::Response, Self::Error>>;

    forward_ready!(service);

    fn call(&self, req: ServiceRequest) -> Self::Future {
        let start = Instant::now();

        let api_key = Arc::clone(&self.api_key);
        let config = Arc::clone(&self.config);
        let buffer = Arc::clone(&self.buffer);
        let client = Arc::clone(&self.client);
        let hostname = (self.config.get_hostname)(&req);
        let ip_address = if self.config.privacy_level >= 2 {
            None
        } else {
            let val = (self.config.get_ip_address)(&req);
            if val.is_empty() { None } else { Some(val) }
        };
        let path = (self.config.get_path)(&req);
        let method = req.method().to_string();
        let user_agent = (self.config.get_user_agent)(&req);
        let user_id = {
            let val = (self.config.get_user_id)(&req);
            if val.is_empty() { None } else { Some(val) }
        };

        let future = self.service.call(req);

        Box::pin(async move {
            let res = future.await?;
            let response_time = start.elapsed().as_millis().min(u32::MAX as u128) as u32;

            let request_data = RequestData {
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                response_time,
                status: res.status().as_u16(),
                user_id,
                created_at: Utc::now().to_rfc3339(),
            };

            log_request(&buffer, &client, &api_key, request_data, &config);

            Ok(res)
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use actix_web::test::TestRequest;

    #[test]
    fn test_get_ip_cf_connecting_ip() {
        let req = TestRequest::get()
            .insert_header(("cf-connecting-ip", "203.0.113.1"))
            .insert_header(("x-forwarded-for", "10.0.0.1"))
            .to_srv_request();
        assert_eq!(get_ip_address(&req), "203.0.113.1");
    }

    #[test]
    fn test_get_ip_x_forwarded_for_first() {
        let req = TestRequest::get()
            .insert_header(("x-forwarded-for", "203.0.113.1, 10.0.0.1"))
            .to_srv_request();
        assert_eq!(get_ip_address(&req), "203.0.113.1");
    }

    #[test]
    fn test_get_ip_x_real_ip() {
        let req = TestRequest::get()
            .insert_header(("x-real-ip", "203.0.113.1"))
            .to_srv_request();
        assert_eq!(get_ip_address(&req), "203.0.113.1");
    }

    #[test]
    fn test_get_ip_empty_when_no_headers() {
        let req = TestRequest::get().to_srv_request();
        // No proxy headers and no peer addr in test context
        let ip = get_ip_address(&req);
        // Either empty or a loopback — just assert it doesn't panic
        let _ = ip;
    }

    #[test]
    fn test_get_hostname() {
        let req = TestRequest::get()
            .insert_header(("host", "example.com"))
            .to_srv_request();
        assert_eq!(get_hostname(&req), "example.com");
    }

    #[test]
    fn test_get_path() {
        let req = TestRequest::get().uri("/api/users").to_srv_request();
        assert_eq!(get_path(&req), "/api/users");
    }

    #[test]
    fn test_get_user_agent() {
        let req = TestRequest::get()
            .insert_header(("user-agent", "TestAgent/1.0"))
            .to_srv_request();
        assert_eq!(get_user_agent(&req), "TestAgent/1.0");
    }

    #[test]
    fn test_get_user_id_default_empty() {
        let req = TestRequest::get().to_srv_request();
        assert_eq!(get_user_id(&req), "");
    }
}
