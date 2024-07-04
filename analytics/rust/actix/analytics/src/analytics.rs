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
use std::sync::{Arc, Mutex};
use std::{
    future::{ready, Ready},
    time::Instant,
};

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
    pub fn new(
        hostname: String,
        ip_address: String,
        path: String,
        user_agent: String,
        method: String,
        response_time: u32,
        status: u16,
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

type StringMapper = dyn for<'a> Fn(&ServiceRequest) -> String + Send + Sync;

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

fn get_hostname(req: &ServiceRequest) -> String {
    req.headers()
        .get(HOST)
        .map(|x| x.to_string())
        .unwrap_or_default()
}

fn get_ip_address(req: &ServiceRequest) -> String {
    if let Some(val) = req.peer_addr() {
        return val.ip().to_string();
    };
    String::new()
}

fn get_path(req: &ServiceRequest) -> String {
    req.path().to_string()
}

fn get_user_agent(req: &ServiceRequest) -> String {
    req
            .headers()
            .get(USER_AGENT)
            .map(|x| x.to_string())
            .unwrap_or_default()
}

fn get_user_id(_req: &ServiceRequest) -> String {
    String::new()
}

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
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_hostname = Arc::new(mapper);
        self
    }

    pub fn with_ip_address_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_ip_address = Arc::new(mapper);
        self
    }

    pub fn with_path_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_path = Arc::new(mapper);
        self
    }

    pub fn with_user_agent_mapper<F>(mut self, mapper: F) -> Self
    where
        F: Fn(&ServiceRequest) -> String + Send + Sync + 'static,
    {
        self.config.get_user_agent = Arc::new(mapper);
        self
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
            api_key: Arc::new(self.api_key.clone()),
            config: Arc::new(self.config.clone()),
            service,
        }))
    }
}

pub struct AnalyticsMiddleware<S> {
    api_key: Arc<String>,
    config: Arc<Config>,
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
            framework: String::from("Actix"),
            privacy_level,
        }
    }
}

async fn post_requests(data: Payload, server_url: String) {
    let _ = Client::new()
        .post(server_url + "api/log-request")
        .json(&data)
        .send()
        .await;
}

async fn log_request(api_key: &str, request_data: RequestData, config: &Config) {
    REQUESTS.lock().unwrap().push(request_data);
    if LAST_POSTED.lock().unwrap().elapsed().as_secs_f64() > 60.0 {
        let payload = Payload::new(
            api_key.to_owned(),
            REQUESTS.lock().unwrap().to_vec(),
            config.privacy_level,
        );
        let server_url = config.server_url.to_owned();
        REQUESTS.lock().unwrap().clear();
        post_requests(payload, server_url).await;
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

        let api_key = Arc::clone(&self.api_key);
        let config = Arc::clone(&self.config);
        let hostname = (self.config.get_hostname)(&req);
        let ip_address = (self.config.get_ip_address)(&req);
        let path = (self.config.get_path)(&req);
        let method = req.method().to_string();
        let user_agent = (self.config.get_hostname)(&req);
        let user_id = (self.config.get_user_id)(&req);

        let future = self.service.call(req);

        Box::pin(async move {
            let res = future.await?;
            let elapsed = start.elapsed().as_millis();

            let request_data = RequestData::new(
                hostname,
                ip_address,
                path,
                user_agent,
                method,
                elapsed.try_into().unwrap(),
                res.status().as_u16(),
                user_id,
                Utc::now().to_rfc3339(),
            );

            spawn(async move { log_request(&api_key, request_data, &config).await });

            Ok(res)
        })
    }
}
