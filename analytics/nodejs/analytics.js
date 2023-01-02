import fetch from "node-fetch";

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

    let data = {
      api_key: apiKey,
      hostname: req.headers.host,
      ip_address: req.headers['x-forwarded-for'] || req.socket.remoteAddress,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: res.statusCode,
      method: req.method,
      framework: 'Express',
      response_time: Math.round((performance.now() - start) / 1000),
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
      ip_address: req.headers['x-forwarded-for'] || req.socket.remoteAddress,
      user_agent: req.headers["user-agent"],
      path: req.url,
      status: reply.statusCode,
      method: req.method,
      framework: 'Fastify',
      response_time: Math.round((performance.now() - start) / 1000),
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
      ip_address: ctx.headers['x-forwarded-for'] || ctx.socket.remoteAddress,
      user_agent: ctx.headers["user-agent"],
      path: ctx.url,
      status: ctx.status,
      method: ctx.method,
      framework: 'Koa',
      response_time: Math.round((performance.now() - start) / 1000),
    };

    logRequest(data);
  };
}
