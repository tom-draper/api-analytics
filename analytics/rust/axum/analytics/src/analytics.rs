use axum::{body::Body, extract::ConnectInfo, http::Request, response::Response};
use chrono::Utc;
use futures::Future;
use http::{
    header::{HeaderValue, HOST, USER_AGENT},
    Extensions, HeaderMap,
};
use lazy_static::lazy_static;
use reqwest::blocking::Client;
use serde::Serialize;
use std::{
    net::{IpAddr, SocketAddr},
    task::{Context, Poll},
    thread::spawn,
    time::Instant,
};
use std::{
    pin::Pin,
    sync::{Arc, Mutex},
};
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
            server_url: String::from("https://www.apianalytics-server.com/"),
            get_hostname: Arc::new(get_hostname),
            get_ip_address: Arc::new(get_ip_address),
            get_path: Arc::new(get_path),
            get_user_agent: Arc::new(get_user_agent),
            get_user_id: Arc::new(get_user_id),
        }
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
    "".to_string()
}

#[derive(Clone)]
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
            api_key: Arc::new(self.api_key.clone()),
            config: Arc::new(self.config.clone()),
            inner,
        }
    }
}

#[derive(Clone)]
pub struct AnalyticsMiddleware<S> {
    api_key: Arc<String>,
    config: Arc<Config>,
    inner: S,
}

pub trait HeaderValueExt {
    fn to_string(&self) -> String;
}

impl HeaderValueExt for HeaderValue {
    fn to_string(&self) -> String {
        self.to_str().unwrap_or_default().to_string()
    }
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

lazy_static! {
    static ref REQUESTS: Mutex<Vec<RequestData>> = Mutex::new(vec![]);
    static ref LAST_POSTED: Mutex<Instant> = Mutex::new(Instant::now());
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
            framework: String::from("Axum"),
            privacy_level,
        }
    }
}

fn post_requests(data: Payload, server_url: String) {
    let _ = Client::new()
        .post(server_url + "api/log-request")
        .json(&data)
        .send();
}

fn log_request(api_key: &str, request_data: RequestData, config: &Config) {
    REQUESTS.lock().unwrap().push(request_data);
    if LAST_POSTED.lock().unwrap().elapsed().as_secs_f64() > 60.0 {
        let payload = Payload::new(
            api_key.to_owned(),
            REQUESTS.lock().unwrap().to_vec(),
            config.privacy_level,
        );
        let server_url = config.server_url.to_owned();
        REQUESTS.lock().unwrap().clear();
        spawn(|| post_requests(payload, server_url));
        *LAST_POSTED.lock().unwrap() = Instant::now();
    }
}

impl<S> Service<Request<Body>> for AnalyticsMiddleware<S>
where
    S: Service<Request<Body>, Response = Response> + Send + 'static,
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

        let api_key = Arc::clone(&self.api_key);
        let config = Arc::clone(&self.config);
        let hostname = (self.config.get_hostname)(&req);
        let ip_address = (self.config.get_ip_address)(&req);
        let path = (self.config.get_path)(&req);
        let method = req.method().to_string();
        let user_agent = (self.config.get_user_agent)(&req);
        let user_id = (self.config.get_user_id)(&req);

        let future = self.inner.call(req);

        Box::pin(async move {
            let res: Response = future.await?;

            let request_data = RequestData::new(
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                res.status().as_u16(),
                now.elapsed().as_millis().try_into().unwrap(),
                user_id,
                Utc::now().to_rfc3339(),
            );

            log_request(&api_key, request_data, &config);

            Ok(res)
        })
    }
}
