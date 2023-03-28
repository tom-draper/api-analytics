import fetch from "node-fetch";

let requests = [];
let last_posted = new Date();

async function logRequest(data) {
  let now = new Date();
  requests.push(data);
  if ((now - last_posted) / 1000 > 60) {
    await fetch("http://213.168.248.206/api/log-request", {
      method: "POST",
      body: JSON.stringify(requests),
      headers: {
        "Content-Type": "application/json",
      },
    });
    requests = [];
    last_posted = now;
  }
}

export function expressAnalytics(apiKey) {
  return (req, res, next) => {
    let start = performance.now();
    next();

    let data = {
      api_key: apiKey,
      hostname: req.headers.host,
      ip_address: req.headers["x-forwarded-for"] || req.socket.remoteAddress,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: res.statusCode,
      method: req.method,
      framework: "Express",
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date(),
    };

    logRequest(data);
  };
}

export function fastifyAnalytics(apiKey) {
  return (req, reply, done) => {
    let start = performance.now();
    done();

    let data = {
      api_key: apiKey,
      hostname: req.headers.host,
      ip_address: req.headers["x-forwarded-for"] || req.socket.remoteAddress,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: reply.statusCode,
      method: req.method,
      framework: "Fastify",
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date(),
    };

    logRequest(data);
  };
}

export function koaAnalytics(apiKey) {
  return async (ctx, next) => {
    let start = performance.now();
    await next();

    let data = {
      api_key: apiKey,
      hostname: ctx.headers.host,
      ip_address: ctx.headers["x-forwarded-for"] || ctx.socket.remoteAddress,
      user_agent: ctx.headers["user-agent"],
      path: ctx.url,
      status: ctx.status,
      method: ctx.method,
      framework: "Koa",
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date(),
    };

    logRequest(data);
  };
}
