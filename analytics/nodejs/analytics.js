import fetch from "node-fetch";

let requests = [];
let lastPosted = new Date();

/** API Analytics config to define custom mapper functions that can overwrite
 * the default functionality when extracting data from the request. */
export class Config {
  constructor() {
    this.getPath = null;
    this.getHostname = null;
    this.getIPAddress = null;
    this.getUserAgent = null;
    this.getUserID = null;
  }
}

async function logRequest(apiKey, requestData, framework) {
  if (apiKey === "" || apiKey === null) {
    return;
  }
  const now = new Date();
  requests.push(requestData);
  if (now - lastPosted > 60000) {
    await fetch("https://www.apianalytics-server.com/api/log-request", {
      method: "POST",
      body: JSON.stringify({
        api_key: apiKey,
        requests: requests,
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

/**
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} config - Configuration for API Analytics
 * @returns {function}
 */
export function expressAnalytics(apiKey, config = new Config()) {
  return (req, res, next) => {
    const start = performance.now();
    next();

    const requestData = {
      hostname: getHostname(res, config),
      ip_address: getIPAddress(res, config),
      user_agent: getUserAgent(res, config),
      path: getPath(res, config),
      status: res.statusCode,
      method: req.method,
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date().toISOString(),
    };

    logRequest(apiKey, requestData, "Express");
  };
}

/**
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} config - Configuration for API Analytics
 * @returns {function}
 */
export function fastifyAnalytics(apiKey, config = new Config()) {
  return (req, reply, done) => {
    const start = performance.now();
    done();

    const requestData = {
      hostname: getHostname(res, config),
      ip_address: getIPAddress(res, config),
      user_agent: getUserAgent(res, config),
      path: getPath(res, config),
      status: reply.statusCode,
      method: req.method,
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date().toISOString(),
    };

    logRequest(apiKey, requestData, "Fastify");
  };
}

/**
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} config - Configuration for API Analytics
 * @returns {function}
 */
export function koaAnalytics(apiKey, config = new Config()) {
  return async (ctx, next) => {
    const start = performance.now();
    await next();

    const requestData = {
      hostname: getHostname(ctx, config),
      ip_address: getIPAddress(ctx, config),
      user_agent: getUserAgent(ctx, config),
      path: getPath(ctx, config),
      status: ctx.status,
      method: ctx.method,
      response_time: Math.round((performance.now() - start) / 1000),
      created_at: new Date().toISOString(),
    };

    logRequest(apiKey, requestData, "Koa");
  };
}

function getHostname(res, config) {
  if (config.getHostname) {
    return config.getHostname();
  }
  return res.headers.host;
}

function getIPAddress(res, config) {
  if (config.getIPAddress) {
    return config.getIPAddress();
  }
  return (
    req.headers["x-forwarded-for"] ||
    req.headers["X-Forwarded-For"] ||
    req.socket.remoteAddress
  );
}

function getUserAgent(res, config) {
  if (config.getUserAgent) {
    return config.getUserAgent();
  }
  return res.headers["user-agent"] || res.headers["User-Agent"];
}

function getPath(res, config) {
  if (config.getPath) {
    return config.getPath();
  }
  return res.url;
}

function getUserID(res, config) {
  if (config.getUserID) {
    return config.getUserID();
  }
  return null;
}
