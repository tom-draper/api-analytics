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
use std::{
    future::{ready, Ready},
    time::Instant,
};

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
    created_at: Utc,
}
impl Data {
    pub fn new(
        api_key: String,
        hostname: String,
        ip_address: String,
        path: String,
        user_agent: String,
        method: String,
        response_time: u32,
        status: u16,
        created_at: Utc,
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
            framework: String::from("Actix"),
            created_at,
        }
    }
}

pub struct Analytics {
    api_key: String,
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        Self { api_key }
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
        ready(Ok(AnalyticsMiddleware {
            api_key: self.api_key.clone(),
            requests: Vec::new(),
            last_posted: Instant::now(),
            service,
        }))
    }
}

pub struct AnalyticsMiddleware<S> {
    api_key: String,
    requests: Vec<Data>,
    last_posted: Instant,
    service: S,
}

pub trait HeaderValueExt {
    fn to_string(&self) -> String;
}

impl HeaderValueExt for HeaderValue {
    fn to_string(&self) -> String {
        self.to_str().unwrap_or_default().to_string()
    }
}

fn extract_ip_address(req: &ServiceRequest) -> String {
    if let Some(val) = req.peer_addr() {
        return val.ip().to_string();
    };
    return String::new();
}

pub trait Logger {
    fn log_request(&self, data: Data);
}

async fn post(requests: Vec<Data>) {
    let _ = Client::new()
        .post("http://213.168.248.206/api/log-request")
        .json(&requests)
        .send()
        .await;
}

impl<S, B> Logger for AnalyticsMiddleware<S>
where
    S: Service<ServiceRequest, Response = ServiceResponse<B>, Error = Error>,
    S::Future: 'static,
    B: 'static,
{
    fn log_request(&self, data: Data) {
        self.requests.push(data);
        if self.last_posted.elapsed().as_secs_f64() > 60.0 {
            spawn(async { post(self.requests).await });
            self.requests.clear();
            self.last_posted = Instant::now();
        }
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
        let now = Instant::now();

        let api_key = self.api_key.clone();
        let hostname = req
            .headers()
            .get(HOST)
            .map(|x| x.to_string())
            .unwrap_or_default();
        let ip_address = extract_ip_address(&req);
        let path = req.path().to_string();
        let method = req.method().to_string();
        let user_agent = req
            .headers()
            .get(USER_AGENT)
            .map(|x| x.to_string())
            .unwrap_or_default();

        let future = self.service.call(req);

        Box::pin(async move {
            let res = future.await?;
            let elapsed = now.elapsed().as_millis();

            let data = Data::new(
                api_key,
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                elapsed.try_into().unwrap(),
                res.status().as_u16(),
                Utc::now(),
            );

            self.log_request(data);
            Ok(res)
        })
    }
}
