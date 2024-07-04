use chrono::Utc;
use lazy_static::lazy_static;
use reqwest::blocking::Client;
use rocket::fairing::{Fairing, Info, Kind};
use rocket::http::Status;
use rocket::request::{FromRequest, Outcome};
use rocket::{Data, Request, Response};
use serde::Serialize;
use std::sync::Mutex;
use std::thread::spawn;
use std::time::Instant;

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
    req.host().unwrap().to_string()
}

fn get_ip_address(req: &Request) -> String {
    req.client_ip().unwrap().to_string()
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
    "".to_string()
}

#[derive(Default)]
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
}

#[derive(Clone)]
pub struct Start<T = Instant>(T);

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
            framework: String::from("Rocket"),
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
    if LAST_POSTED.lock().unwrap().elapsed().as_secs_f64() > 5.0 {
        let payload = Payload::new(
            api_key.to_string(),
            REQUESTS.lock().unwrap().to_vec(),
            config.privacy_level,
        );
        let server_url = config.server_url.to_owned();
        REQUESTS.lock().unwrap().clear();
        spawn(|| post_requests(payload, server_url));
        *LAST_POSTED.lock().unwrap() = Instant::now();
    }
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
        let ip_address = (self.config.get_ip_address)(req);
        let method = req.method().to_string();
        let user_agent = (self.config.get_user_agent)(req);
        let path = (self.config.get_path)(req);
        let user_id = (self.config.get_user_id)(req);

        let request_data = RequestData::new(
            hostname,
            ip_address,
            path,
            user_agent,
            method,
            start.unwrap().elapsed().as_millis() as u32,
            res.status().code,
            user_id,
            Utc::now().to_rfc3339(),
        );

        log_request(&self.api_key, request_data, &self.config);
    }
}

// Allows a route to access the start time
#[rocket::async_trait]
impl<'r> FromRequest<'r> for Start {
    type Error = ();

    async fn from_request(request: &'r Request<'_>) -> Outcome<Self, ()> {
        match &*request.local_cache(|| Start::<Option<Instant>>(None)) {
            Start(Some(start)) => Outcome::Success(Start(start.to_owned())),
            Start(None) => Outcome::Failure((Status::InternalServerError, ())),
        }
    }
}
