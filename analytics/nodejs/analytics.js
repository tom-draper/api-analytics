import fetch from "node-fetch";

let methodMap = {
  GET: 0,
  POST: 1,
  PUT: 2,
  PATCH: 3,
  DELETE: 4,
  OPTIONS: 5,
  CONNECT: 6,
  HEAD: 7,
  TRACE: 8,
};

async function logRequest(data) {
  fetch("https://api-analytics-server.vercel.app/api/log-request", {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
      "Content-Type": "application/json",
    },
  });
}

export function expressAnalytics(apiKey) {
  return (req, res, next) => {
    let start = performance.now();
    next();
    let elapsed = Math.round((performance.now() - start) / 1000);

    let data = {
      api_key: apiKey,
      hostname: req.headers.host,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: res.statusCode,
      method: methodMap[req.method],
      response_time: elapsed,
      framework: 4,
    };

    logRequest(data);
  };
}

export function fastifyAnalytics(apiKey) {
  return (req, reply, done) => {
    let start = performance.now();
    done();
    let elapsed = Math.round((performance.now() - start) / 1000);

    let data = {
      api_key: apiKey,
      hostname: req.headers.host,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: reply.statusCode,
      method: methodMap[req.method],
      response_time: elapsed,
      framework: 5,
    };

    logRequest(data);
  };
}

export function koaAnalytics(apiKey) {
  return async (ctx, next) => {
    let start = performance.now();
    await next();
    let elapsed = Math.round((performance.now() - start) / 1000);

    let data = {
      api_key: apiKey,
      hostname: ctx.headers.host,
      user_agent: ctx.headers["user-agent"],
      path: ctx.url,
      status: ctx.status,
      method: methodMap[ctx.method],
      response_time: elapsed,
      framework: 6,
    };

    logRequest(data);
  };
}
