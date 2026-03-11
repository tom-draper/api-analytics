import { Analytics, Config } from "@api-analytics/core";

function getIPAddress(req, config) {
	if (config.privacyLevel >= 2) return null;
	return (
		req.headers.get("cf-connecting-ip") ??
		req.headers.get("x-forwarded-for")?.split(",")[0].trim() ??
		req.headers.get("x-real-ip") ??
		null
	);
}

/**
 * Wraps a Bun fetch handler with API Analytics logging.
 *
 * @example
 * ```js
 * Bun.serve({
 *   port: 3000,
 *   fetch: bunAnalytics("your-api-key", (req) => {
 *     return new Response("Hello World");
 *   }),
 * });
 * ```
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {(req: Request) => Response | Promise<Response>} handler - Your Bun fetch handler
 * @param {Config} [config] - Configuration for API Analytics
 * @returns {(req: Request) => Promise<Response>}
 */
export function bunAnalytics(apiKey, handler, config = new Config()) {
	const analytics = new Analytics(apiKey, "Bun", config);
	return async (req) => {
		const start = performance.now();
		const response = await handler(req);
		const url = new URL(req.url);
		analytics.logRequest({
			hostname: url.hostname,
			ip_address: getIPAddress(req, config),
			user_agent: req.headers.get("user-agent") ?? "",
			path: url.pathname,
			status: response.status,
			method: req.method,
			response_time: Math.round(performance.now() - start),
			user_id: config.getUserID(req),
			created_at: new Date().toISOString(),
		});
		return response;
	};
}
