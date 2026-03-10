import { Analytics, Config, getIPAddress } from "@api-analytics/core";

/**
 * Fastify plugin for logging analytics data.
 *
 * @param {import('fastify').FastifyInstance} fastify
 * @param {string} apiKey - API Key for API Analytics
 * @param {Config} [config] - Configuration for API Analytics
 */
export function fastifyAnalytics(fastify, apiKey, config = new Config()) {
	const analytics = new Analytics(apiKey, "Fastify", config);
	fastify.addHook("onResponse", (request, reply, done) => {
		analytics.logRequest({
			hostname: config.getHostname(request),
			ip_address: getIPAddress(request, config),
			user_agent: config.getUserAgent(request),
			path: config.getPath(request),
			status: reply.statusCode,
			method: request.method,
			response_time: Math.round(reply.elapsedTime),
			user_id: config.getUserID(request),
			created_at: new Date().toISOString(),
		});
		done();
	});
}
