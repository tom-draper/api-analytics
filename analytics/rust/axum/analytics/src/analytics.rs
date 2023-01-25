use axum::{body::Body, extract::ConnectInfo, http::Request, response::Response};
use futures::future::BoxFuture;
use http::{
    header::{HeaderValue, HOST, USER_AGENT},
    Extensions, HeaderMap,
};
use reqwest::blocking::Client;
use serde::Serialize;
use std::{
    net::{IpAddr, SocketAddr},
    task::{Context, Poll},
    thread::spawn,
    time::Instant,
};
use tower::{Layer, Service};

#[derive(Debug, Serialize)]
struct Data {
    api_key: String,
    hostname: String,
    ip_address: String,
    path: String,
    user_agent: String,
    method: String,
    response_time: u32,
    status: u16,
    framework: String,
}

impl Data {
    pub fn new(
        api_key: String,
        hostname: String,
        ip_address: String,
        path: String,
        user_agent: String,
        method: String,
        status: u16,
        response_time: u32,
    ) -> Self {
        Self {
            api_key,
            hostname,
            ip_address,
            path,
            user_agent,
            method,
            response_time,
            status,
            framework: String::from("Axum"),
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

fn log_request(data: Data) {
    let _ = Client::new()
        .post("https://api-analytics-server.vercel.app/api/log-request")
        .json(&data)
        .send();
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
        let hostname = req.headers().get(HOST).map(|x| x.to_string()).unwrap();
        let ip_address = extract_ip_address(&req);
        let path = req.uri().path().to_owned();
        let method = req.method().to_string();
        let user_agent = req
            .headers()
            .get(USER_AGENT)
            .map(|x| x.to_string())
            .unwrap();

        let future = self.inner.call(req);
        
        Box::pin(async move {
            let res: Response = future.await?;

            let data = Data::new(
                api_key,
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                res.status().as_u16(),
                now.elapsed().as_millis().try_into().unwrap(),
            );

            spawn(|| log_request(data));
            Ok(res)
        })
    }
}
