use chrono::Utc;
use reqwest::Client;
use rocket::fairing::{Fairing, Info, Kind};
use rocket::http::Status;
use rocket::request::{FromRequest, Outcome};
use rocket::{Data, Request, Response};
use serde::Serialize;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

#[derive(Debug, Clone, Serialize)]
struct RequestData {
    hostname: String,
    ip_address: Option<String>,
    path: String,
    user_agent: String,
    method: String,
    response_time: u32,
    status: u16,
    user_id: Option<String>,
    created_at: String,
}

type StringMapper = dyn for<'a> Fn(&Request<'a>) -> String + Send + Sync;

struct Config {
    privacy_level: i32,
    server_url: String,
    get_hostname: Box<StringMapper>,
    get_ip_address: Box<StringMapper>,
    get_path: Box<StringMapper>,
    get_user_agent: Box<StringMapper>,
    get_user_id: Box<StringMapper>,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            privacy_level: 0,
            server_url: String::from("https://www.apianalytics-server.com/"),
            get_hostname: Box::new(get_hostname),
            get_ip_address: Box::new(get_ip_address),
            get_path: Box::new(get_path),
            get_user_agent: Box::new(get_user_agent),
            get_user_id: Box::new(get_user_id),
        }
    }
}

fn get_hostname(req: &Request) -> String {
    req.host()
        .map(|h| h.to_string())
        .unwrap_or_default()
}

fn get_ip_address(req: &Request) -> String {
    if let Some(val) = req
        .headers()
        .get_one("cf-connecting-ip")
        .map(|s| s.trim().to_owned())
        .filter(|s| !s.is_empty())
    {
        return val;
    }

    if let Some(val) = req
        .headers()
        .get_one("x-forwarded-for")
        .and_then(|s| s.split(',').next())
        .map(|s| s.trim().to_owned())
        .filter(|s| !s.is_empty())
    {
        return val;
    }

    if let Some(val) = req
        .headers()
        .get_one("x-real-ip")
        .map(|s| s.trim().to_owned())
        .filter(|s| !s.is_empty())
    {
        return val;
    }

    req.client_ip()
        .map(|ip| ip.to_string())
        .unwrap_or_default()
}

fn get_path(req: &Request) -> String {
    req.uri().path().to_string()
}

fn get_user_agent(req: &Request) -> String {
    req.headers()
        .get_one("User-Agent")
        .unwrap_or_default()
        .to_owned()
}

fn get_user_id(_req: &Request) -> String {
    String::new()
}

struct RequestBuffer {
    requests: Vec<RequestData>,
    last_posted: Instant,
}

impl RequestBuffer {
    fn new() -> Self {
        Self {
            requests: Vec::new(),
            last_posted: Instant::now(),
        }
    }
}

pub struct Analytics {
    api_key: String,
    config: Config,
    buffer: Arc<Mutex<RequestBuffer>>,
    client: Arc<Client>,
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(10))
            .build()
            .unwrap_or_else(|_| Client::new());

        Self {
            api_key,
            config: Config::default(),
            buffer: Arc::new(Mutex::new(RequestBuffer::new())),
            client: Arc::new(client),
        }
    }

    pub fn with_privacy_level(mut self, privacy_level: i32) -> Self {
        self.config.privacy_level = privacy_level;
        self
    }

    pub fn with_server_url(mut self, server_url: String) -> Self {
        self.config.server_url = if server_url.ends_with('/') {
            server_url
        } else {
            server_url + "/"
        };
        self
    }

    pub fn with_hostname_mapper<F>(mut self, mapper: F) -> Self
    where
        F: for<'a> Fn(&Request<'a>) -> String + Send + Sync + 'static,
    {
        self.config.get_hostname = Box::new(mapper);
        self
    }

    pub fn with_ip_address_mapper<F>(mut self, mapper: F) -> Self
    where
        F: for<'a> Fn(&Request<'a>) -> String + Send + Sync + 'static,
    {
        self.config.get_ip_address = Box::new(mapper);
        self
    }

    pub fn with_path_mapper<F>(mut self, mapper: F) -> Self
    where
        F: for<'a> Fn(&Request<'a>) -> String + Send + Sync + 'static,
    {
        self.config.get_path = Box::new(mapper);
        self
    }

    pub fn with_user_agent_mapper<F>(mut self, mapper: F) -> Self
    where
        F: for<'a> Fn(&Request<'a>) -> String + Send + Sync + 'static,
    {
        self.config.get_user_agent = Box::new(mapper);
        self
    }

    pub fn with_user_id_mapper<F>(mut self, mapper: F) -> Self
    where
        F: for<'a> Fn(&Request<'a>) -> String + Send + Sync + 'static,
    {
        self.config.get_user_id = Box::new(mapper);
        self
    }
}

#[derive(Clone)]
pub struct Start<T = Instant>(T);

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
            framework: String::from("Rocket"),
            privacy_level,
        }
    }
}

async fn post_requests(client: &Client, data: Payload, server_url: &str) {
    let url = format!("{}api/log-request", server_url);
    let _ = client.post(url).json(&data).send().await;
}

#[rocket::async_trait]
impl Fairing for Analytics {
    fn info(&self) -> Info {
        Info {
            name: "API Analytics",
            kind: Kind::Request | Kind::Response,
        }
    }

    async fn on_request(&self, req: &mut Request<'_>, _data: &mut Data<'_>) {
        req.local_cache(|| Start::<Option<Instant>>(Some(Instant::now())));
    }

    async fn on_response<'r>(&self, req: &'r Request<'_>, res: &mut Response<'r>) {
        let start = &req.local_cache(|| Start::<Option<Instant>>(None)).0;

        let hostname = (self.config.get_hostname)(req);
        let ip_address = if self.config.privacy_level >= 2 {
            None
        } else {
            let val = (self.config.get_ip_address)(req);
            if val.is_empty() { None } else { Some(val) }
        };
        let method = req.method().to_string();
        let user_agent = (self.config.get_user_agent)(req);
        let path = (self.config.get_path)(req);
        let user_id = {
            let val = (self.config.get_user_id)(req);
            if val.is_empty() { None } else { Some(val) }
        };
        let response_time = start
            .map(|s| s.elapsed().as_millis().min(u32::MAX as u128) as u32)
            .unwrap_or(0);

        let request_data = RequestData {
            hostname,
            ip_address,
            path,
            user_agent,
            method,
            response_time,
            status: res.status().code,
            user_id,
            created_at: Utc::now().to_rfc3339(),
        };

        let batch = {
            let mut buf = self.buffer.lock().unwrap();
            buf.requests.push(request_data);
            if buf.last_posted.elapsed().as_secs_f64() > 60.0 {
                buf.last_posted = Instant::now();
                std::mem::take(&mut buf.requests)
            } else {
                vec![]
            }
        };

        if !batch.is_empty() {
            let payload = Payload::new(self.api_key.clone(), batch, self.config.privacy_level);
            let server_url = self.config.server_url.clone();
            let client = Arc::clone(&self.client);
            tokio::spawn(async move {
                post_requests(&client, payload, &server_url).await;
            });
        }
    }
}

// Allows a route to access the start time
#[rocket::async_trait]
impl<'r> FromRequest<'r> for Start {
    type Error = ();

    async fn from_request(request: &'r Request<'_>) -> Outcome<Self, ()> {
        match &*request.local_cache(|| Start::<Option<Instant>>(None)) {
            Start(Some(start)) => Outcome::Success(Start(start.to_owned())),
            Start(None) => Outcome::Error((Status::InternalServerError, ())),
        }
    }
}
