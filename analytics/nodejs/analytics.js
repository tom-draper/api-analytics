import fetch from "node-fetch";

/** API Analytics config to define custom mapper functions that can overwrite
 * the default functionality when extracting data from the request. */
export class Config {
  constructor() {
    /**
     * A custom mapping function that takes a request and returns the path 
     * stored within the request.
     * If set, it overrides the default behaviour of API Analytics.
     * @type {((req: Request) => string) | null}
     * @public
     */
    this.getPath = null;

    /**
     * A custom mapping function that takes a request and returns the hostname 
     * stored within the request.
     * If set, it overrides the default behaviour of API Analytics.
     * @type {((req: Request) => string) | null}
     * @public
     */
    this.getHostname = null;

    /**
     * A custom mapping function that takes a request and returns the IP address
     * stored within the request.
     * If set, it overrides the default behaviour of API Analytics.
     * @type {((req: Request) => string) | null}
     * @public
     */
    this.getIPAddress = null;

    /**
     * A custom mapping function that takes a request and returns the user agent
     * stored within the request.
     * If set, it overrides the default behaviour of API Analytics.
     * @type {((req: Request) => string) | null}
     * @public
     */
    this.getUserAgent = null;

    /**
     * A custom mapping function that takes a request and returns a user ID 
     * stored within the request.
     * If set, this can be used to track a custom user ID specific to your API 
     * such as an API key or client ID. If left as `null`, no custom user ID 
     * will be used, and user identification will rely on client IP address only.
     * @type {((req: Request) => string) | null}
     * @public
     */
    this.getUserID = null;

    /** 
     * Controls client identification by IP address. 
     * - 0: Sends client IP to the server to be stored and client location is 
     * inferred. 
     * - 1: Sends the client IP to the server only for the location to be 
     * inferred and stored, with the IP discarded afterwards.
     * - 2: Avoids sending the client IP address to the server. Providing a 
     * custom `get_user_id` mapping function becomes the only method for client 
     * identification.
     * Defaults to 0.
     * @type {number}
     * @public
     */
    this.privacyLevel = 0;
  }
}

/**
 * A request object, holding information related to a request, that will be sent 
 * the server to be logged.
 * @typedef {{
 *   hostname: string;
 *   ip_address: string;
 *   user_agent: string;
 *   path: string;
 *   status: number;
 *   method: string;
 *   response_time: number;
 *   user_id: string;
 *   created_at: string;
 * }} RequestData
 */

class Analytics {
  /**
   * @param {string} apiKey - API Key for API Analytics
   * @param {string} framework - Web framework that is being used
   * @param {Config} config - Configuration for API Analytics
   */
  constructor(apiKey, framework, config = new Config()) {
    this.apiKey = apiKey;
    this.framework = framework
    this.config = config;
    this.requests = [];
    this.lastPosted = new Date();
  }

  /**
   * Logs a request to storage. If time interval has elapsed, post all stored 
   * logs to the server.
   * @param {RequestData} requestData - Request data to be logged
   * @returns {Promise<void>}
   */
  async logRequest(requestData) {
    if (this.apiKey === "" || this.apiKey === null) {
      return;
    }
    this.requests.push(requestData);
    const now = new Date();
    if (now - this.lastPosted > 60000) {
      await fetch("https://www.apianalytics-server.com/api/log-request", {
        method: "POST",
        body: JSON.stringify({
          api_key: this.apiKey,
          requests: this.requests,
          framework: this.framework,
          privacy_level: this.config.privacyLevel
        }),
        headers: {
          "Content-Type": "application/json",
        },
      });
      this.requests = [];
      this.lastPosted = now;
    }
  }
}

/**
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} config - Configuration for API Analytics
 * @returns {function}
 */
export function expressAnalytics(apiKey, config = new Config()) {
  const analytics = new Analytics(apiKey, config, "Express");
  return (req, res, next) => {
    const start = performance.now();
    next();

    const requestData = {
      hostname: getHostname(req, config),
      ip_address: getIPAddress(req, config),
      user_agent: getUserAgent(req, config),
      path: getPath(req, config),
      status: res.statusCode,
      method: req.method,
      response_time: Math.round((performance.now() - start) / 1000),
      user_id: getUserID(req, config),
      created_at: new Date().toISOString(),
    };

    analytics.logRequest(requestData);
  };
}

/**
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} config - Configuration for API Analytics
 * @returns {function}
 */
export function fastifyAnalytics(apiKey, config = new Config()) {
  const analytics = new Analytics(apiKey, config, "Fastify");
  return (req, reply, done) => {
    const start = performance.now();
    done();

    const requestData = {
      hostname: getHostname(req, config),
      ip_address: getIPAddress(req, config),
      user_agent: getUserAgent(req, config),
      path: getPath(req, config),
      status: reply.statusCode,
      method: req.method,
      response_time: Math.round((performance.now() - start) / 1000),
      user_id: getUserID(req, config),
      created_at: new Date().toISOString(),
    };

    analytics.logRequest(requestData);
  };
}

/**
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} config - Configuration for API Analytics
 * @returns {function}
 */
export function koaAnalytics(apiKey, config = new Config()) {
  const analytics = new Analytics(apiKey, config, "Koa");
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
      user_id: getUserID(ctx, config),
      created_at: new Date().toISOString(),
    };

    analytics.logRequest(requestData);
  };
}

/**
 * Gets the hostname from the request, using the custom mapping function if 
 * provided, or the default behaviour if not.
 * @param {Request} req 
 * @param {Config} config 
 * @returns {string}
 */
function getHostname(req, config) {
  if (config.getHostname) {
    return config.getHostname(req);
  }
  return req.headers.host;
}

/**
 * Gets the IP address from the request, using the custom mapping function if 
 * provided, or the default behaviour if not.
 * @param {Request} req 
 * @param {Config} config 
 * @returns {string}
 */
function getIPAddress(req, config) {
  if (config.privacyLevel >= 2) {
    return null;
  }

  if (config.getIPAddress) {
    return config.getIPAddress(req);
  }
  return (
    req.headers["x-forwarded-for"] ||
    req.headers["X-Forwarded-For"] ||
    req.socket.remoteAddress
  );
}

/**
 * Gets the user agent from the request, using the custom mapping function if 
 * provided, or the default behaviour if not.
 * @param {Request} req 
 * @param {Config} config 
 * @returns {string}
 */
function getUserAgent(req, config) {
  if (config.getUserAgent) {
    return config.getUserAgent(req);
  }
  return req.headers["user-agent"] || req.headers["User-Agent"];
}

/**
 * Gets the path from the request, using the custom mapping function if 
 * provided, or the default behaviour if not.
 * @param {Request} req 
 * @param {Config} config 
 * @returns {string}
 */
function getPath(req, config) {
  if (config.getPath) {
    return config.getPath(req);
  }
  return req.url;
}

/**
 * Gets the user agent from the request, using the custom mapping function if 
 * provided, or no user ID is used.
 * @param {Request} req 
 * @param {Config} config 
 * @returns {string}
 */
function getUserID(req, config) {
  if (config.getUserID) {
    return config.getUserID(req);
  }
  return null;
}
