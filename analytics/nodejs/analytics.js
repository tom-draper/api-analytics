import fetch from "node-fetch";

/** API Analytics config to define custom mapper functions that can overwrite
 * the default functionality when extracting data from the request. */
export class Config {
	constructor() {
		/**
		 * Controls client identification by IP address.
		 * - 0: Sends client IP to the server to be stored and client location is
		 * inferred.
		 * - 1: Sends the client IP to the server only for the location to be
		 * inferred and stored, with the IP discarded afterwards.
		 * - 2: Avoids sending the client IP address to the server. Providing a
		 * custom `getUserID` mapping function becomes the only method for client
		 * identification.
		 * Defaults to 0.
		 * @type {number}
		 * @public
		 */
		this.privacyLevel = 0;

		/**
		 * For self-hosting. Points to the public server url to post
		 * requests to.
		 * @type {string}
		 * @public
		 */
		this.serverUrl = "https://www.apianalytics-server.com/";

		/**
		 * A mapping function that takes a request and returns the path
		 * stored within the request.
		 * Assigning a value will overrides the default behaviour.
		 * @type {(req: Request) => string}
		 * @public
		 */
		this.getPath = Mappers.getPath;

		/**
		 * A mapping function that takes a request and returns the hostname
		 * stored within the request.
		 * Assigning a value will overrides the default behaviour.
		 * @type {(req: Request) => string}
		 * @public
		 */
		this.getHostname = Mappers.getHostname;

		/**
		 * A mapping function that takes a request and returns the IP address
		 * stored within the request.
		 * Assigning a value will overrides the default behaviour.
		 * @type {(req: Request) => string}
		 * @public
		 */
		this.getIPAddress = Mappers.getIPAddress;

		/**
		 * A mapping function that takes a request and returns the user agent
		 * stored within the request.
		 * Assigning a value will overrides the default behaviour.
		 * @type {(req: Request) => string}
		 * @public
		 */
		this.getUserAgent = Mappers.getUserAgent;

		/**
		 * A mapping function that takes a request and returns a user ID
		 * stored within the request. Always returns `null` by default.
		 * Assigning a value allows for tracking a custom user ID specific to your API
		 * such as an API key or client ID. If left as the default value, user
		 * identification may rely on client IP address only (depending on
		 * config `privacyLevel`).
		 * @type {(req: Request) => string}
		 * @public
		 */
		this.getUserID = Mappers.getUserID;
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
	 * @param {Config} [config=new Config()] - Configuration for API Analytics
	 */
	constructor(apiKey, framework, config = new Config()) {
		this.apiKey = apiKey;
		this.framework = framework;
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
		if (!this.apiKey) {
			return;
		}

		this.requests.push(requestData);
		const now = new Date();
		if (now - this.lastPosted > 60000 && this.requests.length > 0) {
			this.lastPosted = now;

			const url = this.getServerEndpoint();
			try {
				await fetch(url, {
					method: "POST",
					body: JSON.stringify({
						api_key: this.apiKey,
						requests: this.requests,
						framework: this.framework,
						privacy_level: this.config.privacyLevel,
					}),
					headers: {
						"Content-Type": "application/json",
					},
				});
			} catch (error) {
				console.error("Failed to send analytics data:", error);
			}

			this.requests = [];
		}
	}

	getServerEndpoint() {
		if (this.config.serverUrl.endsWith("/")) {
			return this.config.serverUrl + "api/log-request";
		}
		return this.config.serverUrl + "/api/log-request";
	}
}

/**
 * Express middleware for logging analytics data.
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config=new Config()] - Configuration for API Analytics
 * @returns {(req: Request, res: Response, next: NextFunction) => void}
 * @typedef {import('express').Request} Request
 * @typedef {import('express').Response} Response
 * @typedef {import('express').NextFunction} NextFunction
 */
export function expressAnalytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Express", config);
	return (req, res, next) => {
		const start = performance.now();
		
		res.on('finish', () => {
			const requestData = {
				hostname: config.getHostname(req),
				ip_address: getIPAddress(req, config),
				user_agent: config.getUserAgent(req),
				path: config.getPath(req),
				status: res.statusCode,
				method: req.method,
				response_time: Math.round(performance.now() - start),
				user_id: config.getUserID(req),
				created_at: new Date().toISOString(),
			};

			analytics.logRequest(requestData);
		});
		
		next();
	};
}

/**
 * Fastify middleware for logging analytics data.
 *
 * @param {FastifyInstance} fastify - The Fastify instance.
 * @param {string} apiKey - The API key for the analytics service.
 * @param {Config} [config=new Config()] - The configuration object (optional).
 * @returns {void}
 *
 * @typedef {import('fastify').FastifyInstance} FastifyInstance
 */
export function useFastifyAnalytics(fastify, apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Fastify", config);

	fastify.addHook("onRequest", (request, reply, done) => {
		request.startTime = performance.now();
		done();
	});

	fastify.addHook("onResponse", (request, reply, done) => {
		const responseTime = Math.round(performance.now() - request.startTime);
		const requestData = {
			hostname: config.getHostname(request),
			ip_address: getIPAddress(request, config),
			user_agent: config.getUserAgent(request),
			path: config.getPath(request),
			status: reply.statusCode,
			method: request.method,
			response_time: responseTime,
			user_id: config.getUserID(request),
			created_at: new Date().toISOString(),
		};

		analytics.logRequest(requestData);
		done();
	});
}

/**
 * Koa middleware for logging analytics data.
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config=new Config()] - Configuration for API Analytics
 * @returns {(ctx: Koa.Context, next: any) => Promise<void>}
 */
export function koaAnalytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Koa", config);

	return async (ctx, next) => {
		const start = performance.now();
		await next();

		const requestData = {
			hostname: config.getHostname(ctx),
			ip_address: getIPAddress(ctx, config),
			user_agent: config.getUserAgent(ctx),
			path: config.getPath(ctx),
			status: ctx.status,
			method: ctx.method,
			response_time: Math.round(performance.now() - start),
			user_id: config.getUserID(ctx),
			created_at: new Date().toISOString(),
		};

		analytics.logRequest(requestData);
	};
}

/**
 * Hono middleware for logging analytics data.
 *
 * @param {string} apiKey - The API key for the analytics service.
 * @param {Config} [config=new Config()] - The configuration object (optional).
 * @returns {(req: Hono.Request, res: Hono.Response, next: Function) => Promise<void>}
 */
export function honoAnalytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Hono", config);

	return async (c, next) => {
		const start = performance.now();
		await next(); // Call next middleware/route handler
		const responseTime = Math.round(performance.now() - start);

		const requestData = {
			hostname: config.getHostname(c.req),
			ip_address: getIPAddress(c.req, config),
			user_agent: config.getUserAgent(c.req),
			path: config.getPath(c.req),
			status: c.res.status,
			method: c.req.method,
			response_time: responseTime,
			user_id: config.getUserID(c.req),
			created_at: new Date().toISOString(),
		};

		analytics.logRequest(requestData);
	};
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
	return config.getIPAddress(req);
}

class Mappers {
	/**
	 * Gets the hostname from the request, using the custom mapping function if
	 * provided, or the default behaviour if not.
	 * @param {Request} req
	 * @returns {string}
	 */
	static getHostname(req) {
		return req.headers.host;
	}

	/**
	 * Gets the IP address from the request, using the custom mapping function if
	 * provided, or the default behaviour if not.
	 * @param {Request} req
	 * @returns {string}
	 */
	static getIPAddress(req) {
		const forwardedFor =
			req.headers["x-forwarded-for"] || req.headers["X-Forwarded-For"];
		return forwardedFor
			? forwardedFor.split(",")[0].trim()
			: req.socket.remoteAddress;
	}

	/**
	 * Gets the user agent from the request, using the custom mapping function if
	 * provided, or the default behaviour if not.
	 * @param {Request} req
	 * @returns {string}
	 */
	static getUserAgent(req) {
		return req.headers["user-agent"] || req.headers["User-Agent"];
	}

	/**
	 * Gets the path from the request, using the custom mapping function if
	 * provided, or the default behaviour if not.
	 * @param {Request} req
	 * @returns {string}
	 */
	static getPath(req) {
		return req.url;
	}

	/**
	 * Gets the user agent from the request, using the custom mapping function if
	 * provided, or no user ID is used.
	 * @param {Request} req
	 * @returns {string}
	 */
	static getUserID(req) {
		return null;
	}
}
