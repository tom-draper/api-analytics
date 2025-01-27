type FrameworkExamples = {
	[framework: string]: {
		install: string;
		codeFile?: string;
		example: string;
	};
};

const frameworkExamples: FrameworkExamples = {
	Django: {
		install: 'pip install api-analytics',
		codeFile: 'settings.py',
		example: `ANALYTICS_API_KEY = <API-KEY>

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]`,
	},
	Flask: {
		install: 'pip install api-analytics',
		example: `from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <API-KEY>)  # Add middleware

@app.get('/')
def root():
    return {'message': 'Hello, World!'}

if __name__ == '__main__':
    app.run()`,
	},
	FastAPI: {
		install: 'pip install fastapi-analytics',
	    example: `import uvicorn
from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, api_key=<API-KEY>)  # Add middleware

@app.get('/')
async def root():
    return {'message': 'Hello, World!'}

if __name__ == '__main__':
    uvicorn.run('app:app', reload=True)`,
	},
	Tornado: {
		install: 'pip install tornado-analytics',
		example: `import asyncio
from tornado.web import Application
from api_analytics.tornado import Analytics

# Inherit from the Analytics middleware class
class MainHandler(Analytics):
    def __init__(self, app, res):
        super().__init__(app, res, <API-KEY>)  # Provide api key
    
    def get(self):
        self.write({'message': 'Hello, World!'})

def make_app():
    return Application([
        (r'/', MainHandler),
    ])

if __name__ == '__main__':
    app = make_app()
    app.listen(8080)
    IOLoop.instance().start()`,
	},
	Express: {
		install: 'npm install node-api-analytics',
		example: `import express from 'express';
import { expressAnalytics } from 'node-api-analytics';

const app = express();

app.use(expressAnalytics(<API-KEY>)); // Add middleware

app.get('/', (req, res) => {
    res.send({ message: 'Hello, World!' });
});

app.listen(8080, () => {
    console.log('Server listening at http://localhost:8080');
})`,
	},
	Fastify: {
		install: 'npm install node-api-analytics',
		example: `import Fastify from 'fastify';
import { fastifyAnalytics } from 'node-api-analytics';

const fastify = Fastify();

fastify.addHook('onRequest', fastifyAnalytics(<API-KEY>)); // Add middleware

fastify.get('/', function (request, reply) {
    reply.send({ message: 'Hello, World!' });
})

fastify.listen({ port: 8080 }, function (err, address) {
    console.log('Server listening at http://localhost:8080');
    if (err) {
        fastify.log.error(err);
        process.exit(1);
    }
})`,
	},
	Koa: {
		install: 'npm install node-api-analytics',
		example: `import Koa from "koa";
import { koaAnalytics } from 'node-api-analytics';

const app = new Koa();

app.use(koaAnalytics(<API-KEY>)); // Add middleware

app.use((ctx) => {
    ctx.body = { message: 'Hello, World!' };
});

app.listen(8080, () =>
    console.log('Server listening at https://localhost:8080')
); `,
	},
	Gin: {
		install:
			'go get -u github.com/tom-draper/api-analytics/analytics/go/gin',
		example: `package main

import(
    "net/http"
    "github.com/gin-gonic/gin"
    analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
)

func root(c * gin.Context) {
    jsonData:= []byte(\`{"message": "Hello, World!"}\`)
    c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
    router := gin.Default()
    
    router.Use(analytics.Analytics(<API-KEY>)) // Add middleware

    router.GET("/", root)
    router.Run(":8080")
}`,
	},
	Echo: {
		install:
			'go get -u github.com/tom-draper/api-analytics/analytics/go/echo',
		example: `package main

import (
    "net/http"
    echo "github.com/labstack/echo/v4"
    analytics "github.com/tom-draper/api-analytics/analytics/go/echo"
)

func root(c echo.Context) error {
    jsonData := []byte(\`{"message": "Hello, World!"}\`)
    return c.JSON(http.StatusOK, jsonData)
}

func main() {
    router := echo.New()

    router.Use(analytics.Analytics(<API-KEY>))

    router.GET("/", root)
    router.Start(":8080")
}`,
	},
	Fiber: {
		install:
			'go get -u github.com/tom-draper/api-analytics/analytics/go/fiber',
		example: `package main

import (
    "github.com/gofiber/fiber/v2"
    analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"
)

func root(c *fiber.Ctx) error {
    jsonData := []byte(\`{"message": "Hello, World!"}\`)
    return c.SendString(string(jsonData))
}

func main() {
    app := fiber.New()

    app.Use(analytics.Analytics(<API-KEY>)) // Add middleware

    app.Get("/", root)
    app.Listen(":8080")
}`,
	},
	Chi: {
		install:
			'go get -u github.com/tom-draper/api-analytics/analytics/go/chi',
		example: `package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    analytics "github.com/tom-draper/api-analytics/analytics/go/chi"
)

func root(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    jsonData := []byte(\`{"message": "Hello, World!"}\`)
    w.Write(jsonData)
}

func main() {
    router := chi.NewRouter()

    router.Use(analytics.Analytics(<API-KEY>)) // Add middleware

    router.GET("/", root)
    router.Run(":8080")
}`,
	},
	Actix: {
		install: 'cargo add actix-analytics',
		example: `use actix_web::{get, web, App, HttpServer, Responder, Result};
use serde::Serialize;
use actix_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

#[get("/")]
async fn index() -> Result<impl Responder> {
    let data = JsonData {
        message: "Hello, World!".to_string(),
    };
    Ok(web::Json(data))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .wrap(Analytics::new(<API-KEY>))  // Add middleware
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}`,
	},
	Axum: {
		install: 'cargo add axum-analytics',
		example: `use axum::{
    routing::get,
    Json, Router,
};
use serde::Serialize;
use std::net::SocketAddr;
use tokio;
use axum_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

async fn root() -> Json<JsonData> {
    let data = JsonData {
        message: "Hello, World!".to_string(),
    };
    Json(data)
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .layer(Analytics::new(<API-KEY>))  // Add middleware
        .route("/", get(root));

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}`,
	},
	Rocket: {
		install: 'cargo add rocket-analytics',
		example: `#[macro_use]
extern crate rocket;
use rocket::serde::json::Json;
use serde::Serialize;
use rocket_analytics::Analytics;

#[derive(Serialize)]
pub struct JsonData {
    message: String,
}

#[get("/")]
fn root() -> Json<JsonData> {
    let data = JsonData {
        message: "Hello, World!".to_string(),
    };
    Json(data)
}

#[launch]
fn rocket() -> _ {
    rocket::build()
        .mount("/", routes![root])
        .attach(Analytics::new(<API-KEY>))  // Add middleware
}`,
	},
	Rails: {
		install: 'gem install api_analytics',
		codeFile: 'config/application.rb',
		example: `require 'rails'
require 'api_analytics'

Bundler.require(*Rails.groups)

module RailsMiddleware
  class Application < Rails::Application
    config.load_defaults 6.1
    config.api_only = true

    config.middleware.use ::Analytics::Rails, <API-KEY> # Add middleware
  end
end`,
	},
	Sinatra: {
		install: 'gem install api_analytics',
		example: `require 'sinatra'
require 'api_analytics'

use Analytics::Sinatra, <API-KEY> # Add middleware

before do
    content_type 'application/json'
end

get '/' do
    {message: 'Hello, World!'}.to_json
end`,
	},
	Laravel: {
		install: 'coming soon',
		codeFile: 'app/Http/Kernel.php',
		example: `protected $middleware = [
    \\App\\Http\\Middleware\\Analytics::class,
    ...
]`,
	},
	'ASP.NET Core': {
		install: 'dotnet add package APIAnalytics.AspNetCore',
		example: `using analytics;
using Microsoft.AspNetCore.Mvc;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

app.UseAnalytics(<API-KEY>); // Add middleware

app.MapGet("/", () =>
{
    return Results.Ok(new OkObjectResult(new { message = "Hello, World!" }));
});

app.Run();`,
	},
};

export default frameworkExamples;
