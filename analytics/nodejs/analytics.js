import fetch from "node-fetch";

let requests = [];
let lastPosted = new Date();

async function logRequest(apiKey, requestData, framework) {
  if (apiKey === "" || apiKey == null) {
    return;
  }
  let now = new Date();
  requests.push(requestData);
  if ((now - lastPosted) / 1000 > 60) {
    await fetch("http://213.168.248.206/api/log-request", {
      method: "POST",
      body: JSON.stringify({
        api_key: apiKey,
        requests: requestData,
        framework: framework,
      }),
      headers: {
        "Content-Type": "application/json",
      },
    });
    requests = [];
    lastPosted = now;
  }
}

export function expressAnalytics(apiKey) {
  return (req, res, next) => {
    let start = performance.now();
    next();

    let requestData = {
      hostname: req.headers.host,
      ip_address: req.headers["x-forwarded-for"] || req.socket.remoteAddress,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: res.statusCode,
      method: req.method,
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date().toISOString(),
    };

    logRequest(apiKey, requestData, "Express");
  };
}

export function fastifyAnalytics(apiKey) {
  return (req, reply, done) => {
    let start = performance.now();
    done();

    let data = {
      hostname: req.headers.host,
      ip_address: req.headers["x-forwarded-for"] || req.socket.remoteAddress,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: reply.statusCode,
      method: req.method,
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date().toISOString(),
    };

    logRequest(apiKey, data, "Fastify");
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
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date().toISOString(),
    };

    logRequest(apiKey, data, "Koa");
  };
}
