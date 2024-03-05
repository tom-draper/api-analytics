use axum::{body::Body, extract::ConnectInfo, http::Request, response::Response};
use chrono::Utc;
use futures::future::BoxFuture;
use http::{
    header::{HeaderValue, HOST, USER_AGENT},
    Extensions, HeaderMap,
};
use lazy_static::lazy_static;
use reqwest::blocking::Client;
use serde::Serialize;
use std::sync::Mutex;
use std::{
    net::{IpAddr, SocketAddr},
    task::{Context, Poll},
    thread::spawn,
    time::Instant,
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
            created_at,
        }
    }
}

#[derive(Clone)]
pub struct Analytics {
    api_key: String,
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        Self { api_key }
    }
}

impl<S> Layer<S> for Analytics {
    type Service = AnalyticsMiddleware<S>;

    fn layer(&self, inner: S) -> Self::Service {
        AnalyticsMiddleware {
            api_key: self.api_key.clone(),
            inner,
        }
    }
}

#[derive(Clone)]
pub struct AnalyticsMiddleware<S> {
    api_key: String,
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

fn extract_ip_address(req: &Request<Body>) -> String {
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

lazy_static! {
    static ref REQUESTS: Mutex<Vec<RequestData>> = Mutex::new(vec![]);
    static ref LAST_POSTED: Mutex<Instant> = Mutex::new(Instant::now());
}

#[derive(Debug, Clone, Serialize)]
struct Payload {
    api_key: String,
    requests: Vec<RequestData>,
    framework: String,
}

impl Payload {
    pub fn new(api_key: String, requests: Vec<RequestData>) -> Self {
        Self {
            api_key,
            requests,
            framework: String::from("Axum"),
        }
    }
}

fn post_requests(data: Payload) {
    let _ = Client::new()
        .post("https://www.apianalytics-server.com/api/log-request")
        .json(&data)
        .send();
}

fn log_request(api_key: String, request_data: RequestData) {
    REQUESTS.lock().unwrap().push(request_data);
    if LAST_POSTED.lock().unwrap().elapsed().as_secs_f64() > 60.0 {
        let payload = Payload::new(api_key, REQUESTS.lock().unwrap().to_vec());
        REQUESTS.lock().unwrap().clear();
        spawn(|| post_requests(payload));
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
    type Future = BoxFuture<'static, Result<Self::Response, Self::Error>>;

    fn poll_ready(&mut self, cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        self.inner.poll_ready(cx)
    }

    fn call(&mut self, req: Request<Body>) -> Self::Future {
        let now = Instant::now();

        let api_key = self.api_key.clone();
        let hostname = req
            .headers()
            .get(HOST)
            .map(|x| x.to_string())
            .unwrap_or_default();
        let ip_address = extract_ip_address(&req);
        let path = req.uri().path().to_owned();
        let method = req.method().to_string();
        let user_agent = req
            .headers()
            .get(USER_AGENT)
            .map(|x| x.to_string())
            .unwrap_or_default();

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
                Utc::now().to_rfc3339(),
            );

            log_request(api_key, request_data);

            Ok(res)
        })
    }
}
