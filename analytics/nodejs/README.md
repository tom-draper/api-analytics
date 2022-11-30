# Node API Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
npm i node-api-analytics
```

#### Express

```js
import express from 'express';
import { expressAnalytics } from 'node-api-analytics';

const app = express();

app.use(expressAnalytics(<api_key>));  // Add middleware

app.get("/", (req, res) => {
    res.send({message: "Hello World"});
});

app.listen(8080, () => {
    console.log('Server listening at localhost:8080');
})
```

#### Fastify

```js
import Fastify from 'fastify';
import { fastifyAnalytics } from 'node-api-analytics;

const fastify = Fastify({
  logger: true,
})

fastify.addHook('onRequest', fastifyAnalytics(<api_key>));  // Add middleware

fastify.get('/', function (request, reply) {
  reply.send({ message: 'Hello World!' });
})

fastify.listen({ port: 8080 }, function (err, address) {
  console.log('Server listening at https://localhost:8080');
  if (err) {
    fastify.log.error(err);
    process.exit(1);
  }
})
```

#### Koa

```js
import Koa from "koa";
import { koaAnalytics } from "node-api-analytics";

const app = new Koa();

app.use(koaAnalytics(<api_key>));  // Add middleware

app.use((ctx) => {
  ctx.body = { message: "Hello World!" };
});

app.listen(8080, () =>
  console.log('Server listening at https://localhost:8080')
);
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.
