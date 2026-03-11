type FrameworkExamples = {
	[framework: string]: {
		install: string;
		codeFile?: string;
		example: string;
	};
};

const frameworkExamples: FrameworkExamples = {
	Django: {
		install: 'pip install api-analytics[django]',
		codeFile: 'settings.py',
		example: `ANALYTICS_API_KEY = <API-KEY>

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]`,
	},
	Flask: {
		install: 'pip install api-analytics[flask]',
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
		install: 'pip install api-analytics[fastapi]',
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
		install: 'pip install api-analytics[tornado]',
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
		install: 'npm install @api-analytics/express',
		example: `import express from 'express';
import { expressAnalytics } from '@api-analytics/express';

const app = express();

app.use(expressAnalytics(<API-KEY>)); // Add middleware

app.get('/', (req, res) => {
    res.send({ message: 'Hello World!' });
});

app.listen(8080, () => {
    console.log('Server listening at http://localhost:8080');
});`,
	},
	Fastify: {
		install: 'npm install @api-analytics/fastify',
		example: `import Fastify from 'fastify';
import { fastifyAnalytics } from '@api-analytics/fastify';

const fastify = Fastify();

fastifyAnalytics(fastify, <API-KEY>); // Add middleware

fastify.get('/', (request, reply) => {
    reply.send({ message: 'Hello World!' });
});

fastify.listen({ port: 8080 }, (err) => {
    if (err) {
        fastify.log.error(err);
        process.exit(1);
    }
    console.log('Server listening at http://localhost:8080');
});`,
	},
	Koa: {
		install: 'npm install @api-analytics/koa',
		example: `import Koa from 'koa';
import { koaAnalytics } from '@api-analytics/koa';

const app = new Koa();

app.use(koaAnalytics(<API-KEY>)); // Add middleware

app.use((ctx) => {
    ctx.body = { message: 'Hello World!' };
});

app.listen(8080, () => {
    console.log('Server listening at http://localhost:8080');
});`,
	},
	Hono: {
		install: 'npm install @api-analytics/hono',
		example: `import { Hono } from 'hono';
import { serve } from '@hono/node-server';
import { honoAnalytics } from '@api-analytics/hono';

const app = new Hono();

app.use('*', honoAnalytics(<API-KEY>)); // Add middleware

app.get('/', (c) => c.json({ message: 'Hello World!' }));

serve(app, (info) => {
    console.log('Server listening at http://localhost:' + info.port);
});`,
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
    data := map[string]string{
        "message": "Hello, World!",
    }
    c.JSON(http.StatusOK, data)
}

func main() {
    r := gin.Default()
    
    r.Use(analytics.Analytics(<API-KEY>)) // Add middleware

    r.GET("/", root)
    r.Run(":8080")
}`,
	},
	Echo: {
		install:
			'go get -u github.com/tom-draper/api-analytics/analytics/go/echo',
		example: `package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    analytics "github.com/tom-draper/api-analytics/analytics/go/echo"
)

func root(c echo.Context) error {
    data := map[string]string{
        "message": "Hello, World!",
    }
    return c.JSON(http.StatusOK, data)
}

func main() {
    e := echo.New()

    e.Use(analytics.Analytics(<API-KEY>))

    e.GET("/", root)
    e.Start(":8080")
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
    data := map[string]string{
        "message": "Hello, World!",
    }
    return c.JSON(data)
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
    "encoding/json"
    "net/http"
    chi "github.com/go-chi/chi/v5"
    analytics "github.com/tom-draper/api-analytics/analytics/go/chi"
)

func root(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{
        "message": "Hello, World!",
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)

    err := json.NewEncoder(w).Encode(data)
    if err != nil {
        http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
        return
    }
}

func main() {
    r := chi.NewRouter()

    r.Use(analytics.Analytics(<API-KEY>)) // Add middleware

    r.Get("/", root)
    http.ListenAndServe(":8080", r)
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
		example: `use axum::{routing::get, Json, Router};
use axum_analytics::Analytics;
use serde::Serialize;
use std::net::SocketAddr;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

async fn root() -> Json<JsonData> {
    let data = JsonData {
        message: "Hello World!".to_string(),
    };
    Json(data)
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .route("/", get(root))
        .layer(Analytics::new(<API-KEY>));  // Add middleware

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
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
		install: 'composer require api-analytics/laravel',
		codeFile: 'app/Http/Kernel.php',
		example: `protected $middleware = [
    \\ApiAnalytics\\Laravel\\AnalyticsMiddleware::class,
    ...
]`,
	},
	'ASP.NET Core': {
		install: 'dotnet add package APIAnalytics.AspNetCore',
		example: `using Analytics;

var builder = WebApplication.CreateBuilder(args);

var app = builder.Build();

app.UseAnalytics(<API-KEY>); // Add middleware

app.MapGet("/", () => Results.Ok(new { message = "Hello, World!" }));

app.Run();`,
	},
};

export default frameworkExamples;

export const frameworkLanguages: Record<string, string> = {
	FastAPI: 'python', Flask: 'python', Django: 'python', Tornado: 'python',
	Express: 'javascript', Fastify: 'javascript', Koa: 'javascript', Hono: 'javascript',
	Gin: 'go', Echo: 'go', Fiber: 'go', Chi: 'go',
	Actix: 'rust', Axum: 'rust', Rocket: 'rust',
	Rails: 'ruby', Sinatra: 'ruby',
	Laravel: 'php',
	'ASP.NET Core': 'csharp',
};
