use axum::{body::Body, extract::ConnectInfo, http::Request, response::Response};
use chrono::Utc;
use futures::Future;
use http::{
    header::{HeaderValue, HOST, USER_AGENT},
    Extensions, HeaderMap,
};
use reqwest::Client;
use serde::Serialize;
use std::{
    net::{IpAddr, SocketAddr},
    pin::Pin,
    sync::Arc,
    task::{Context, Poll},
    time::{Duration, Instant},
};
use tokio::sync::RwLock;
use tokio::time::interval;
use tower::{Layer, Service};

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

type StringMapper = dyn for<'a> Fn(&Request<Body>) -> String + Send + Sync;

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
            server_url: "https://www.apianalytics-server.com/".to_string(),
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

fn get_hostname(req: &Request<Body>) -> String {
    req.headers()
        .get(HOST)
        .map(|x| x.to_string())
        .unwrap_or_default()
}

fn get_ip_address(req: &Request<Body>) -> String {
    let headers = req.headers();
    let extensions = req.extensions();
    if let Some(val) = ip_from_cf_connecting_ip(headers) {
        return val.to_string();
    }
    if let Some(val) = ip_from_x_forwarded_for(headers) {
        return val.to_string();
    }
    if let Some(val) = ip_from_x_real_ip(headers) {
        return val.to_string();
    }
    if let Some(val) = ip_from_connect_info(extensions) {
        return val.to_string();
    }
    String::new()
}

fn ip_from_cf_connecting_ip(headers: &HeaderMap) -> Option<IpAddr> {
    headers
        .get("cf-connecting-ip")
        .and_then(|hv| hv.to_str().ok())
        .and_then(|s| s.trim().parse::<IpAddr>().ok())
}

fn ip_from_x_forwarded_for(headers: &HeaderMap) -> Option<IpAddr> {
    headers
        .get("x-forwarded-for")
        .and_then(|hv| hv.to_str().ok())
        .and_then(|s| s.split(',').find_map(|s| s.trim().parse::<IpAddr>().ok()))
}

fn ip_from_x_real_ip(headers: &HeaderMap) -> Option<IpAddr> {
    headers
        .get("x-real-ip")
        .and_then(|hv| hv.to_str().ok())
        .and_then(|s| s.parse::<IpAddr>().ok())
}

fn ip_from_connect_info(extensions: &Extensions) -> Option<IpAddr> {
    extensions
        .get::<ConnectInfo<SocketAddr>>()
        .map(|ConnectInfo(addr)| addr.ip())
}

fn get_path(req: &Request<Body>) -> String {
    req.uri().path().to_owned()
}

fn get_user_agent(req: &Request<Body>) -> String {
    req.headers()
        .get(USER_AGENT)
        .map(|x| x.to_string())
        .unwrap_or_default()
}

fn get_user_id(_req: &Request<Body>) -> String {
    String::new()
}

#[derive(Clone)]
pub struct Analytics {
    config: Config,
    requests: Arc<RwLock<Vec<RequestData>>>,
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        let requests: Arc<RwLock<Vec<RequestData>>> = Arc::new(RwLock::new(vec![]));
        let requests_clone = Arc::clone(&requests);
        let config = Config::default();
        let privacy_level = config.privacy_level;
        let server_url = config.server_url.clone();

        tokio::spawn(async move {
            let client = Client::builder()
                .timeout(Duration::from_secs(10))
                .build()
                .unwrap_or_else(|_| Client::new());

            let mut ticker = interval(Duration::from_secs(60));
            ticker.tick().await; // skip the immediate first tick
            loop {
                ticker.tick().await;
                let batch = {
                    let mut reqs = requests_clone.write().await;
                    std::mem::take(&mut *reqs)
                };
                if !batch.is_empty() {
                    let payload = Payload::new(api_key.clone(), batch, privacy_level);
                    post_requests(&client, payload, &server_url).await;
                }
            }
        });

        Self { config, requests }
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
        F: Fn(&Request<Body>) -> String + Send + Sync + 'static,
    {
        self.config.get_hostname = Arc::new(mapper);
        self
    }

    pub fn with_ip_address_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&Request<Body>) -> String + Send + Sync + 'static,
    {
        self.config.get_ip_address = Arc::new(mapper);
        self
    }

    pub fn with_path_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&Request<Body>) -> String + Send + Sync + 'static,
    {
        self.config.get_path = Arc::new(mapper);
        self
    }

    pub fn with_user_agent_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&Request<Body>) -> String + Send + Sync + 'static,
    {
        self.config.get_user_agent = Arc::new(mapper);
        self
    }

    pub fn with_user_id_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&Request<Body>) -> String + Send + Sync + 'static,
    {
        self.config.get_user_id = Arc::new(mapper);
        self
    }
}

impl<S> Layer<S> for Analytics {
    type Service = AnalyticsMiddleware<S>;

    fn layer(&self, inner: S) -> Self::Service {
        AnalyticsMiddleware {
            config: Arc::new(self.config.clone()),
            requests: Arc::clone(&self.requests),
            inner,
        }
    }
}

#[derive(Clone)]
pub struct AnalyticsMiddleware<S> {
    config: Arc<Config>,
    requests: Arc<RwLock<Vec<RequestData>>>,
    inner: S,
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
            framework: "Axum".to_string(),
            privacy_level,
        }
    }
}

async fn post_requests(client: &Client, data: Payload, server_url: &str) {
    let url = format!("{}api/log-request", server_url);
    match client.post(url).json(&data).send().await {
        Ok(resp) if resp.status().is_success() => {}
        Ok(resp) => {
            eprintln!(
                "Failed to send analytics data. Server responded with status: {}",
                resp.status()
            );
        }
        Err(err) => {
            eprintln!("Error sending analytics data: {}", err);
        }
    }
}

impl<S> Service<Request<Body>> for AnalyticsMiddleware<S>
where
    S: Service<Request<Body>, Response = Response> + Clone + Send + Sync + 'static,
    S::Future: Send + 'static,
{
    type Response = S::Response;
    type Error = S::Error;
    type Future =
        Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>> + Send + 'static>>;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, req: Request<Body>) -> Self::Future {
        let now = Instant::now();

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
        let requests = Arc::clone(&self.requests);

        let future = self.inner.call(req);

        Box::pin(async move {
            let res: Response = future.await?;
            let response_time = now.elapsed().as_millis().min(u32::MAX as u128) as u32;

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

            requests.write().await.push(request_data);

            Ok(res)
        })
    }
}
