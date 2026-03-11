import { Analytics, Config } from "@api-analytics/core";

function getIPAddress(request, config) {
	if (config.privacyLevel >= 2) return null;
	return (
		request.headers.get("cf-connecting-ip") ??
		request.headers.get("x-forwarded-for")?.split(",")[0].trim() ??
		request.headers.get("x-real-ip") ??
		null
	);
}

/**
 * Elysia plugin for logging analytics data.
 *
 * @param {import('elysia').Elysia} app
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config] - Configuration for API Analytics
 */
export function elysiaAnalytics(app, apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Elysia", config);
	return app
		.derive(() => ({ _analyticsStart: performance.now() }))
		.onAfterHandle(({ request, set, _analyticsStart }) => {
			const url = new URL(request.url);
			const status = typeof set.status === "number" ? set.status : 200;
			analytics.logRequest({
				hostname: url.hostname,
				ip_address: getIPAddress(request, config),
				user_agent: request.headers.get("user-agent") ?? "",
				path: url.pathname,
				status,
				method: request.method,
				response_time: Math.round(performance.now() - _analyticsStart),
				user_id: config.getUserID(request),
				created_at: new Date().toISOString(),
			});
		});
}
