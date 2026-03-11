import { Analytics, Config, getIPAddress } from "@api-analytics/core";

/**
 * Koa middleware for logging analytics data.
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config] - Configuration for API Analytics
 * @returns {(ctx: import('koa').Context, next: () => Promise<void>) => Promise<void>}
 */
export function koaAnalytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Koa", config);
	return async (ctx, next) => {
		const start = performance.now();
		await next();
		analytics.logRequest({
			hostname: config.getHostname(ctx),
			ip_address: getIPAddress(ctx, config),
			user_agent: config.getUserAgent(ctx),
			path: config.getPath(ctx),
			status: ctx.status,
			method: ctx.method,
			response_time: Math.round(performance.now() - start),
			user_id: config.getUserID(ctx),
			created_at: new Date().toISOString(),
		});
	};
}
