/** API Analytics config to define custom mapper functions that can overwrite
 * the default functionality when extracting data from the request. */
export class Config {
	constructor() {
		/**
		 * Controls client identification by IP address.
		 * - 0: Sends client IP to the server to be stored and client location is inferred.
		 * - 1: Sends the client IP to the server only for the location to be inferred
		 *      and stored, with the IP discarded afterwards.
		 * - 2: Avoids sending the client IP address to the server.
		 * Defaults to 0.
		 * @type {number}
		 */
		this.privacyLevel = 0;

		/**
		 * For self-hosting. Points to the server URL to post requests to.
		 * @type {string}
		 */
		this.serverUrl = "https://www.apianalytics-server.com/";

		/** @type {(req: any) => string} */
		this.getPath = Mappers.getPath;

		/** @type {(req: any) => string} */
		this.getHostname = Mappers.getHostname;

		/** @type {(req: any) => string|null} */
		this.getIPAddress = Mappers.getIPAddress;

		/** @type {(req: any) => string} */
		this.getUserAgent = Mappers.getUserAgent;

		/** @type {(req: any) => string|null} */
		this.getUserID = Mappers.getUserID;
	}
}

export class Mappers {
	static getHostname(req) {
		return req.headers.host ?? "";
	}

	static getIPAddress(req) {
		if (req.headers["cf-connecting-ip"]) {
			return req.headers["cf-connecting-ip"];
		}
		if (req.headers["x-forwarded-for"]) {
			return req.headers["x-forwarded-for"].split(",")[0].trim();
		}
		if (req.headers["x-real-ip"]) {
			return req.headers["x-real-ip"];
		}
		return req.socket?.remoteAddress ?? null;
	}

	static getUserAgent(req) {
		return req.headers["user-agent"] ?? "";
	}

	static getPath(req) {
		return req.url;
	}

	static getUserID(_req) {
		return null;
	}
}

/**
 * @typedef {{
 *   hostname: string;
 *   ip_address: string|null;
 *   user_agent: string;
 *   path: string;
 *   status: number;
 *   method: string;
 *   response_time: number;
 *   user_id: string|null;
 *   created_at: string;
 * }} RequestData
 */

export class Analytics {
	/**
	 * @param {string} apiKey
	 * @param {string} framework
	 * @param {Config} config
	 */
	constructor(apiKey, framework, config = new Config()) {
		this.apiKey = apiKey;
		this.framework = framework;
		this.config = config;
		this.requests = [];
		this.lastPosted = new Date();
	}

	/** @param {RequestData} requestData */
	async logRequest(requestData) {
		if (!this.apiKey) return;

		this.requests.push(requestData);
		const now = new Date();
		if (now - this.lastPosted > 60000 && this.requests.length > 0) {
			this.lastPosted = now;
			const requestsToSend = this.requests;
			this.requests = [];

			try {
				await fetch(this.getServerEndpoint(), {
					method: "POST",
					body: JSON.stringify({
						api_key: this.apiKey,
						requests: requestsToSend,
						framework: this.framework,
						privacy_level: this.config.privacyLevel,
					}),
					headers: { "Content-Type": "application/json" },
				});
			} catch (error) {
				console.error("Failed to send analytics data:", error);
			}
		}
	}

	getServerEndpoint() {
		const base = this.config.serverUrl.endsWith("/")
			? this.config.serverUrl
			: this.config.serverUrl + "/";
		return base + "api/log-request";
	}
}

/**
 * Returns the IP address from a Node.js-style request, respecting privacy level.
 * @param {any} req
 * @param {Config} config
 * @returns {string|null}
 */
export function getIPAddress(req, config) {
	if (config.privacyLevel >= 2) return null;
	return config.getIPAddress(req);
}
