import { Analytics, Config, getIPAddress } from "@api-analytics/core";
import { defineEventHandler } from "h3";

/**
 * H3 middleware for logging analytics data.
 *
 * @example
 * ```js
 * import { createApp } from 'h3';
 * import { h3Analytics } from '@api-analytics/h3';
 *
 * const app = createApp();
 * app.use(h3Analytics('your-api-key'));
 * ```
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config] - Configuration for API Analytics
 * @returns {import('h3').EventHandler}
 */
export function h3Analytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "H3", config);

	return defineEventHandler((event) => {
		const start = performance.now();

		event.node.res.on("finish", () => {
			analytics.logRequest({
				hostname: config.getHostname(event.node.req),
				ip_address: getIPAddress(event.node.req, config),
				user_agent: config.getUserAgent(event.node.req),
				path: config.getPath(event.node.req),
				status: event.node.res.statusCode,
				method: event.node.req.method ?? "GET",
				response_time: Math.round(performance.now() - start),
				user_id: config.getUserID(event.node.req),
				created_at: new Date().toISOString(),
			});
		});
	});
}
