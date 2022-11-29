use actix_web::http::header::{HeaderValue, USER_AGENT};
use actix_web::{
    dev::{forward_ready, Service, ServiceRequest, ServiceResponse, Transform},
    Error,
};
use futures::future::LocalBoxFuture;
use serde::{Deserialize, Serialize};
use std::future::{ready, Ready};
use std::time::Instant;

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

#[derive(Debug, Serialize, Deserialize)]
struct Data {
    api_key: String,
    hostname: String,
    path: String,
    user_agent: String,
    method: u32,
    response_time: u32,
    status: u16,
    framework: u32,
}

impl Data {
    fn method_map(method: &str) -> u32 {
        match method {
            "GET" => return 0,
            "POST" => return 1,
            "PUT" => return 2,
            "PATCH" => return 3,
            "DELETE" => return 4,
            "OPTIONS" => return 5,
            "CONNECT" => return 6,
            "HEAD" => return 7,
            "TRACE" => return 8,
            &_ => todo!(),
        }
    }

    pub fn new(
        api_key: String,
        hostname: String,
        path: String,
        user_agent: String,
        method: String,
        response_time: u32,
        status: u16,
    ) -> Self {
        Self {
            api_key,
            hostname,
            path,
            user_agent,
            method: Data::method_map(&method),
            response_time,
            status,
            framework: 9,
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

fn log_request(data: Data) {
    reqwest::Client::new()
        .post("https://api-analytics.vercel.app/log-request")
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
        let hostname = "".to_string();
        let path = req.path().to_string();
        let user_agent: String = req
            .headers()
            .get(USER_AGENT)
            .map(|x| x.to_string())
            .unwrap();
        let method = req.method().to_string();

        let now = Instant::now();
        let fut = self.service.call(req);

        Box::pin(async move {
            let res = fut.await?;
            let elapsed = now.elapsed().as_millis();
            let status = res.status().as_u16();

            let data = Data::new(
                api_key,
                hostname,
                path,
                user_agent,
                method,
                elapsed.try_into().unwrap(),
                status,
            );
            log_request(data);
            Ok(res)
        })
    }
}
