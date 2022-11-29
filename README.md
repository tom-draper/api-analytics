# API Analytics

A lightweight API analytics solution, complete with a dashboard.

Currently compatible with:
 - Python: <b>Django</b>, <b>Flask</b> and <b>FastAPI</b>
 - Node.js: <b>Express</b>, <b>Fastify</b> and <b>Koa</b>
 - Go: <b>Gin</b>, <b>Echo</b>, <b>Fiber</b> and <b>Chi</b>

## Getting Started

### 1. Generate an API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your API analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there is minimal impact on your APIs performance.

#### Django

```bash
python -m pip install api-analytics
```

Set you API key as an environment variable. In `settings.py`:

```py
from os import getenv

ANALYTICS_API_KEY = getenv("API_KEY")

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]
```

#### FastAPI

```bash
pip install api-analytics
```

```py
from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, api_key=<api_key>)

@app.get('/')
async def root():
    return {'message': 'Hello World!'}
```

#### Flask

```bash
pip install api-analytics
```

```py
from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <api_key>)

@app.get('/')
def root():
    return {'message': 'Hello World!'}
```

#### Express

```bash
npm i node-api-analytics
```

```js
import express from 'express';
import { expressAnalytics } from 'node-api-analytics';

const app = express()

app.use(analytics(<api_key>))

app.get('/', (req, res) => {
    res.send({ message: 'Hello World' });
});

app.listen(8080, () => {
    console.log('Server listening at http://localhost:8080');
})
```

#### Fastify

```bash
npm i node-api-analytics
```

```js
import Fastify from 'fastify';
import { fastifyAnalytics } from 'node-api-analytics;

const fastify = Fastify()

fastify.addHook('onRequest', fastifyAnalytics(<api_key>));

fastify.get('/', function (request, reply) {
  reply.send({ message: 'Hello World!' });
})

fastify.listen({ port: 8080 }, function (err, address) {
  console.log('Server listening at http://localhost:8080')
  if (err) {
    fastify.log.error(err);
    process.exit(1);
  }
})
```

#### Koa

```bash
npm i node-api-analytics
```

```js
import Koa from "koa";
import { koaAnalytics } from 'node-api-analytics';

const app = new Koa();

app.use(koaAnalytics(<api_key>));

app.use((ctx) => {
  ctx.body = { message: 'Hello World!' };
});

app.listen(8080, () =>
  console.log('Server listening at https://localhost:8080')
);
```

#### Gin

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/gin
```

```go
package main

import (
	analytics "github.com/tom-draper/api-analytics/analytics/go/gin"
	"net/http"

	"github.com/gin-gonic/gin"
)

func root(c *gin.Context) {
	jsonData := []byte(`{"message": "Hello World!"}`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := gin.Default()
	
	router.Use(analytics.Analytics(<api_key>))

	router.GET("/", root)
	router.Run("localhost:8080")
}
```

#### Echo

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/echo
```

```go
package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	analytics "github.com/tom-draper/api-analytics/analytics/go/echo"
)

func root(c echo.Context) {
	jsonData := []byte(`{"message": "Hello World!"}`)
	c.Data(http.StatusOK, "application/json", jsonData)
}

func main() {
	router := echo.New()

	router.Use(analytics.Analytics(<api_key>))

	router.GET("/", root)
	router.Start("localhost:8080")
}
```

#### Fiber

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/fiber
```

```go
package main

import (
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/fiber"

	"github.com/gofiber/fiber/v2"
)

func root(c *fiber.Ctx) error {
	jsonData := []byte(`{"message": "Hello World!"}`)
	return c.SendString(string(jsonData))
}

func main() {
	app := fiber.New()

	app.Use(analytics.Analytics(<api_key>))

	app.Get("/", root)
	app.Listen(":8080")
}
```

#### Chi

```bash
go get -u github.com/tom-draper/api-analytics/analytics/go/chi
```

```go
package main

import (
	"net/http"
	"os"

	analytics "github.com/tom-draper/api-analytics/analytics/go/chi"

	chi "github.com/go-chi/chi/v5"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"message": "Hello World!"}`)
	w.Write(jsonData)
}

func main() {
	router := chi.NewRouter()

	router.Use(analytics.Analytics(<api_key>))

	router.GET("/", root)
	router.Run("localhost:8080")
}
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.

![dashboard](https://user-images.githubusercontent.com/41476809/204396681-7f38558c-33df-4434-aae8-17703d4422fe.png)

## Data and Security

All data is stored securely in compliance with The EU General Data Protection Regulation (GDPR).

For any given request to your API, data recorded is limited to:
 - API hostname
 - Path requested by user
 - User's operating system
 - User's browser
 - Request method (GET, POST, PUT, etc.)
 - Time of request
 - Status code
 - Response time
 - API framework (FastAPI, Flask, Express etc.)

Data collected is only used by the analytics dashboard.

When using API Analytics to collect analytics for your API, you are anonymous, with the API key the only link between you and you API's analytics. Should you lose your API key, you will have no method to access your API analytics.

You can use the same API key across multiple APIs, but all your data will appear in the same dashboard.
