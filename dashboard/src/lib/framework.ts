// import hljs from "highlight.js";

// import python from 'highlight.js/lib/languages/python';
// hljs.registerLanguage('python', python);

let frameworkExamples = {
    Django: {
        install: "pip install api-analytics",
        codeFile: 'settings.py',
        example: `ANALYTICS_API_KEY = <api_key>

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]`
    },
    Flask: {
        install: "pip install api-analytics",
        codeFile: '',
        example: `from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, api_key=<api_key>)  # Add middleware

@app.get('/')
async def root():
    return {'message': 'Hello World!'}`
    },
    FastAPI: {
        install: "pip install api-analytics",
        codeFile: '',
        example: `from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <api_key>)  # Add middleware

@app.get('/')
def root():
    return {'message': 'Hello World!'}`
    },
    Tornado: {
        install: "pip install api-analytics",
        codeFile: '',
        example: `import asyncio
from tornado.web import Application

from api_analytics.tornado import Analytics

# Inherit from the Analytics middleware class
class MainHandler(Analytics):
    def __init__(self, app, res):
        super().__init__(app, res, <api_key>)  # Pass api key

    def get(self):
        self.write({'message': 'Hello World!'})

def make_app():
    return Application([
        (r"/", MainHandler),
    ])

async def main():
    app = make_app()
    app.listen(8080)
    await asyncio.Event().wait()

if __name__ == "__main__":
    asyncio.run(main())`
    },
    Express: {
        install: 'npm install node-api-analytics',
        codeFile: '',
        example: `import express from 'express';
import { expressAnalytics } from 'node-api-analytics';

const app = express();

app.use(analytics(<api_key>));  // Add middleware

app.get('/', (req, res) => {
    res.send({ message: 'Hello World' });
});

app.listen(8080, () => {
    console.log('Server listening at http://localhost:8080');
})`
    },
    Fastify: {
        install: 'npm install node-api-analytics',
        codeFile: '',
        example: `import Fastify from 'fastify';
import { fastifyAnalytics } from 'node-api-analytics;

const fastify = Fastify();

fastify.addHook('onRequest', fastifyAnalytics(<api_key>));  // Add middleware

fastify.get('/', function (request, reply) {
  reply.send({ message: 'Hello World!' });
})

fastify.listen({ port: 8080 }, function (err, address) {
  console.log('Server listening at http://localhost:8080');
  if (err) {
    fastify.log.error(err);
    process.exit(1);
  }
})`
    },
    Koa: {
        install: 'npm install node-api-analytics',
        codeFile: '',
        example: `import Koa from "koa";
import { koaAnalytics } from 'node-api-analytics';

const app = new Koa();

app.use(koaAnalytics(<api_key>));  // Add middleware

app.use((ctx) => {
  ctx.body = { message: 'Hello World!' };
});

app.listen(8080, () =>
  console.log('Server listening at https://localhost:8080')
);`
    },
    Gin: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/gin',
        codeFile: '',
        example: `package main

import (
	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
	"net/http"

	"github.com/gin-gonic/gin"
)

func root(c *gin.Context) {
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := gin.Default()
	
	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Run("localhost:8080")
}`
    },
    Echo: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/echo',
        codeFile: '',
        example: `package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	analytics "github.com/tom-draper/api-analytics/analytics/go/echo"
)

func root(c echo.Context) {
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := echo.New()

	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Start("localhost:8080")
}`
    },
    Fiber: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/fiber',
        codeFile: '',
        example: `package main

import (
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"

	"github.com/gofiber/fiber/v2"
)

func root(c *fiber.Ctx) error {
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	return c.SendString(string(jsonData))
}

func main() {
	app := fiber.New()

	app.Use(analytics.Analytics(<api_key>)) // Add middleware

	app.Get("/", root)
	app.Listen(":8080")
}`
    },
    Chi: {
        install: 'go get -u github.com/tom-draper/api-analytics/analytics/go/chi',
        codeFile: '',
        example: `package main

import (
	"net/http"
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/chi"

	chi "github.com/go-chi/chi/v5"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(\`{"message": "Hello World!"}\`)
	w.Write(jsonData)
}

func main() {
	router := chi.NewRouter()

	router.Use(analytics.Analytics(<api_key>)) // Add middleware

	router.GET("/", root)
	router.Run("localhost:8080")
}`
    },
    Actix: {
        install: 'cargo add actix-analytics',
        codeFile: '',
        example: `use actix_web::{get, web, Responder, Result};
use serde::Serialize;
use actix_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

#[get("/")]
async fn index() -> Result<impl Responder> {
    let json_data = JsonData {
        message: "Hello World!".to_string(),
    };
    Ok(web::Json(json_data))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    use actix_web::{App, HttpServer};

    HttpServer::new(|| {
        App::new()
            .wrap(Analytics::new(<api_key>))  // Add middleware
            .service(index)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}`
    },
    Axum: {
        install: 'cargo add axum-analytics',
        codeFile: '',
        example: `use axum::{
    routing::get,
    Json, Router,
};
use serde::Serialize;
use std::net::SocketAddr;
use tokio;
use actum_analytics::Analytics;

#[derive(Serialize)]
struct JsonData {
    message: String,
}

async fn root() -> Json<JsonData> {
    let json_data = JsonData {
        message: "Hello World!".to_string(),
    };
    Json(json_data)
}

#[tokio::main]
async fn main() {
    let app = Router::new()
        .layer(Analytics::new(<api_key>))  // Add middleware
        .route("/", get(root));

    let addr = SocketAddr::from(([127, 0, 0, 1], 8080));
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}`
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

    config.middleware.use ::Analytics::Rails, <api_key> # Add middleware
  end
end`
    },
    Sinatra: {
        install: 'gem install api_analytics',
        codeFile: '',
        example: `require 'sinatra'
require 'api_analytics'

use Analytics::Sinatra, <api_key>

before do
    content_type 'application/json'
end

get '/' do
    {message: 'Hello World!'}.to_json
end`
    }
}

export default frameworkExamples;