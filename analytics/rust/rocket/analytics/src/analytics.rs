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

#[derive(Default)]
pub struct Analytics {
    api_key: String,
}

impl Analytics {
    pub fn new(api_key: String) -> Self {
        Self { api_key }
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
}

impl Payload {
    pub fn new(api_key: String, requests: Vec<RequestData>) -> Self {
        Self {
            api_key,
            requests,
            framework: String::from("Rocket"),
        }
    }
}

fn post_requests(data: Payload) {
    let _ = Client::new()
        .post("https://www.apianalytics-server.com/api/log-request")
        .json(&data)
        .send();
}

fn log_request(api_key: &str, request_data: RequestData) {
    REQUESTS.lock().unwrap().push(request_data);
    if LAST_POSTED.lock().unwrap().elapsed().as_secs_f64() > 60.0 {
        let payload = Payload::new(api_key.to_string(), REQUESTS.lock().unwrap().to_vec());
        REQUESTS.lock().unwrap().clear();
        spawn(|| post_requests(payload));
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
        let hostname = req.host().unwrap().to_string();
        let ip_address = req.client_ip().unwrap().to_string();
        let method = req.method().to_string();
        let user_agent = req
            .headers()
            .get_one("User-Agent")
            .unwrap_or_default()
            .to_owned();
        let path = req.uri().path().to_string();

        let request_data = RequestData::new(
            hostname,
            ip_address,
            path,
            user_agent,
            method,
            start.unwrap().elapsed().as_millis() as u32,
            res.status().code,
            Utc::now().to_rfc3339(),
        );

        log_request(&self.api_key, request_data);
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
