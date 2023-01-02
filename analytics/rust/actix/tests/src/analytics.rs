use actix_web::http::header::{HeaderValue, USER_AGENT, HOST};
use actix_web::{
    dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform},
    Error,
};
use std::thread::spawn;
use futures::future::LocalBoxFuture;
use serde::Serialize;
use std::future::{ready, Ready};
use std::time::Instant;
use reqwest::blocking::Client;


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
        response_time: u32,
        status: u16,
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

fn log_request(data: Data) {
    let _ = Client::new()
        .post("https://api-analytics-server.vercel.app/api/log-request")
        .json(&data)
        .send();
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
        let api_key = self.api_key.clone();
        let hostname = req.headers().get(HOST).map(|x| x.to_string()).unwrap();
        let mut ip_address = String::new();
        if let Some(val) = req.peer_addr() {
            ip_address = val.ip().to_string();
        };
        let path = req.path().to_string();
        let method = req.method().to_string();
        let user_agent = req.headers().get(USER_AGENT).map(|x| x.to_string()).unwrap();

        let now = Instant::now();
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
            );

            spawn(|| log_request(data));
            Ok(res)
        })
    }
}

