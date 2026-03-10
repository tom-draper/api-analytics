import { Analytics, Config } from "@api-analytics/core";

function getIPAddress(req, config) {
	if (config.privacyLevel >= 2) return null;
	return (
		req.header("cf-connecting-ip") ??
		req.header("x-forwarded-for")?.split(",")[0].trim() ??
		req.header("x-real-ip") ??
		null
	);
}

/**
 * Hono middleware for logging analytics data.
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config] - Configuration for API Analytics
 * @returns {import('hono').MiddlewareHandler}
 */
export function honoAnalytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Hono", config);
	return async (c, next) => {
		const start = performance.now();
		await next();
		analytics.logRequest({
			hostname: new URL(c.req.url).hostname,
			ip_address: getIPAddress(c.req, config),
			user_agent: c.req.header("user-agent") ?? "",
			path: new URL(c.req.url).pathname,
			status: c.res.status,
			method: c.req.method,
			response_time: Math.round(performance.now() - start),
			user_id: config.getUserID(c.req),
			created_at: new Date().toISOString(),
		});
	};
}
