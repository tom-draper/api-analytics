import { Analytics, Config, getIPAddress } from "@api-analytics/core";

/**
 * NestJS middleware for logging analytics data.
 *
 * Apply globally in main.ts:
 *   app.use(nestjsAnalytics('your-api-key'));
 *
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config] - Configuration for API Analytics
 * @returns {(req: any, res: any, next: () => void) => void}
 */
export function nestjsAnalytics(apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "NestJS", config);
	return (req, res, next) => {
		const start = performance.now();
		res.on("finish", () => {
			analytics.logRequest({
				hostname: config.getHostname(req),
				ip_address: getIPAddress(req, config),
				user_agent: config.getUserAgent(req),
				path: config.getPath(req),
				status: res.statusCode,
				method: req.method,
				response_time: Math.round(performance.now() - start),
				user_id: config.getUserID(req),
				created_at: new Date().toISOString(),
			});
		});
		next();
	};
}
