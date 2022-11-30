use axum::{
    response::Response,
    body::Body,
    http::Request,
};
use http::header::{HOST, USER_AGENT, HeaderValue};
use futures::future::BoxFuture;
use tower::{Service, Layer};
use std::task::{Context, Poll};
use std::thread::spawn;
use std::time::Instant;
use serde::Serialize;


#[derive(Debug, Serialize)]
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
            _ => panic!("invalid method"),
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
            framework: 10,
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
        AnalyticsMiddleware { api_key: self.api_key.clone(), inner }
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

fn log_request(data: Data) {
    let _ = reqwest::blocking::Client::new()
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
        let api_key = self.api_key.clone();
        let hostname = req.headers().get(HOST).map(|x| x.to_string()).unwrap();
        let path = req.uri().path().to_owned();
        let method = req.method().to_string();
        let user_agent = req.headers().get(USER_AGENT).map(|x| x.to_string()).unwrap();

        let now = Instant::now();
        let future = self.inner.call(req);

        Box::pin(async move {
            let res: Response = future.await?;
            let elapsed = now.elapsed().as_millis();

            let data = Data::new(
                api_key,
                hostname,
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
