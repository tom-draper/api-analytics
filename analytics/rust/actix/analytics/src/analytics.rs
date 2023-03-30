use actix_rt::spawn;
use actix_web::{
    dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform},
    http::header::{HeaderValue, HOST, USER_AGENT},
    Error,
};
use chrono::Utc;
use futures::future::LocalBoxFuture;
use lazy_static::lazy_static;
use reqwest::Client;
use serde::Serialize;
use std::sync::Mutex;
use std::{
    future::{ready, Ready},
    time::Instant,
};

#[derive(Debug, Clone, Serialize)]
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
    created_at: String,
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
        created_at: String,
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
            service,
        }))
    }
}

pub struct AnalyticsMiddleware<S> {
    api_key: String,
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

lazy_static! {
    static ref REQUESTS: Mutex<Vec<Data>> = Mutex::new(vec![]);
    static ref LAST_POSTED: Mutex<Instant> = Mutex::new(Instant::now());
}

async fn post_requests(requests: Vec<Data>) {
    let _ = Client::new()
        .post("http://213.168.248.206/api/log-request")
        .json(&requests)
        .send()
        .await;
}

async fn log_request(data: Data) {
    REQUESTS.lock().unwrap().push(data);
    if LAST_POSTED.lock().unwrap().elapsed().as_secs_f64() > 60.0 {
        let requests = REQUESTS.lock().unwrap().to_vec();
        REQUESTS.lock().unwrap().clear();
        post_requests(requests).await;
        *LAST_POSTED.lock().unwrap() = Instant::now();
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
            let elapsed = start.elapsed().as_millis();

            let data = Data::new(
                api_key,
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                elapsed.try_into().unwrap(),
                res.status().as_u16(),
                Utc::now().to_rfc3339(),
            );

            spawn(async { log_request(data).await });

            Ok(res)
        })
    }
}
