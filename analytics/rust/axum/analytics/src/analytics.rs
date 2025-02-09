use axum::{body::Body, extract::ConnectInfo, http::Request, response::Response};
use chrono::Utc;
use futures::Future;
use http::{
    header::{HeaderValue, HOST, USER_AGENT},
    Extensions, HeaderMap,
};
use lazy_static::lazy_static;
use reqwest::Client;
use serde::Serialize;
use std::{
    net::{IpAddr, SocketAddr},
    task::{Context, Poll},
    time::Instant,
};
use std::{pin::Pin, sync::Arc};
use tokio::sync::RwLock;
use tokio::time::{interval, Duration};
use tower::{Layer, Service};

#[derive(Debug, Clone, Serialize)]
struct RequestData {
    hostname: String,
    ip_address: String,
    path: String,
    user_agent: String,
    method: String,
    response_time: u32,
    status: u16,
    user_id: String,
    created_at: String,
}

impl RequestData {
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        hostname: String,
        ip_address: String,
        path: String,
        user_agent: String,
        method: String,
        status: u16,
        response_time: u32,
        user_id: String,
        created_at: String,
    ) -> Self {
        Self {
            hostname,
            ip_address,
            path,
            user_agent,
            method,
            response_time,
            status,
            user_id,
            created_at,
        }
    }
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
    let extensions = req.extensions();
    let headers = req.headers();
    let mut ip_address = String::new();
    if let Some(val) = ip_from_x_forwarded_for(headers) {
        ip_address = val.to_string();
    } else if let Some(val) = ip_from_x_real_ip(headers) {
        ip_address = val.to_string();
    } else if let Some(val) = ip_from_connect_info(extensions) {
        ip_address = val.to_string();
    }
    ip_address
}

fn ip_from_x_forwarded_for(headers: &HeaderMap) -> Option<IpAddr> {
    headers
        .get("x-forwarded-for")
        .and_then(|hv| hv.to_str().ok())
        .and_then(|s| {
            s.split(',')
                .rev()
                .find_map(|s| s.trim().parse::<IpAddr>().ok())
        })
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
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        let config = Config::default();
        let api_key_clone = api_key.clone();
        let privacy_level = config.privacy_level;
        let server_url = config.server_url.clone();

        tokio::spawn(async move {
            let mut interval = interval(Duration::from_secs(60));
            loop {
                interval.tick().await; // Wait for the next interval

                let mut requests = REQUESTS.write().await;
                if !requests.is_empty() {
                    let payload = Payload::new(
                        api_key_clone.clone(),
                        requests.clone(),
                        privacy_level,
                    );

                    requests.clear(); // Clear requests after reading
                    drop(requests); // Explicitly drop the lock

                    let url = server_url.clone();

                    tokio::spawn(async move { post_requests(payload, url).await });
                }
            }
        });

        Self { config }
    }

    pub fn with_privacy_level(mut self, privacy_level: i32) -> Self {
        self.config.privacy_level = privacy_level;
        self
    }

    pub fn with_server_url(mut self, server_url: String) -> Self {
        if server_url.ends_with("/") {
            self.config.server_url = server_url;
        } else {
            self.config.server_url = server_url + "/";
        }
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
}

impl<S> Layer<S> for Analytics {
    type Service = AnalyticsMiddleware<S>;

    fn layer(&self, inner: S) -> Self::Service {
        AnalyticsMiddleware {
            config: Arc::new(self.config.clone()),
            inner,
        }
    }
}

#[derive(Clone)]
pub struct AnalyticsMiddleware<S> {
    config: Arc<Config>,
    inner: S,
}

lazy_static! {
    static ref REQUESTS: Arc<RwLock<Vec<RequestData>>> = Arc::new(RwLock::new(vec![]));
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

async fn post_requests(data: Payload, server_url: String) {
    let client = Client::new();
    let response = client
        .post(server_url + "api/log-request")
        .json(&data)
        .send()
        .await;

    match response {
        Ok(resp) if resp.status().is_success() => {
            return;
        }
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

async fn log_request(request_data: RequestData) {
    let mut requests = REQUESTS.write().await;
    requests.push(request_data);
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
        let ip_address = (self.config.get_ip_address)(&req);
        let path = (self.config.get_path)(&req);
        let method = req.method().to_string();
        let user_agent = (self.config.get_user_agent)(&req);
        let user_id = (self.config.get_user_id)(&req);

        let future = self.inner.call(req);

        Box::pin(async move {
            let res: Response = future.await?;

            let response_time = now.elapsed().as_millis().min(u32::MAX as u128) as u32;

            let request_data = RequestData::new(
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                res.status().as_u16(),
                response_time,
                user_id,
                Utc::now().to_rfc3339(),
            );

            tokio::spawn(log_request(request_data));

            Ok(res)
        })
    }
}
