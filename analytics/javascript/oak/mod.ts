import type { Context, Next } from "@oak/oak";

export interface Config {
	/**
	 * Controls client identification by IP address.
	 * - 0: Sends client IP to the server to be stored and client location is inferred.
	 * - 1: Sends the client IP to the server only for location inference, then discarded.
	 * - 2: Avoids sending the client IP address to the server.
	 * Defaults to 0.
	 */
	privacyLevel?: number;
	/** Override for self-hosting. */
	serverUrl?: string;
	/** Custom user ID extractor. */
	getUserID?: (ctx: Context) => string | null;
}

interface RequestData {
	hostname: string;
	ip_address: string | null;
	user_agent: string;
	path: string;
	status: number;
	method: string;
	response_time: number;
	user_id: string | null;
	created_at: string;
}

class Analytics {
	private apiKey: string;
	private serverUrl: string;
	private privacyLevel: number;
	private requests: RequestData[] = [];
	private lastPosted: Date = new Date();

	constructor(apiKey: string, serverUrl: string, privacyLevel: number) {
		this.apiKey = apiKey;
		this.serverUrl = serverUrl;
		this.privacyLevel = privacyLevel;
	}

	async logRequest(requestData: RequestData): Promise<void> {
		if (!this.apiKey) return;

		this.requests.push(requestData);
		const now = new Date();
		if (now.getTime() - this.lastPosted.getTime() > 60000 && this.requests.length > 0) {
			this.lastPosted = now;
			const requestsToSend = this.requests;
			this.requests = [];

			try {
				await fetch(this.getServerEndpoint(), {
					method: "POST",
					body: JSON.stringify({
						api_key: this.apiKey,
						requests: requestsToSend,
						framework: "Oak",
						privacy_level: this.privacyLevel,
					}),
					headers: { "Content-Type": "application/json" },
				});
			} catch {
				// Silently fail — analytics should never affect the application
			}
		}
	}

	private getServerEndpoint(): string {
		const base = this.serverUrl.endsWith("/") ? this.serverUrl : this.serverUrl + "/";
		return base + "api/log-request";
	}
}

function getIPAddress(ctx: Context, privacyLevel: number): string | null {
	if (privacyLevel >= 2) return null;
	return (
		ctx.request.headers.get("cf-connecting-ip") ??
		ctx.request.headers.get("x-forwarded-for")?.split(",")[0].trim() ??
		ctx.request.headers.get("x-real-ip") ??
		ctx.request.ip ??
		null
	);
}

/**
 * Oak middleware for logging analytics data.
 *
 * @example
 * ```ts
 * import { Application } from "@oak/oak";
 * import { oakAnalytics } from "@api-analytics/oak";
 *
 * const app = new Application();
 * app.use(oakAnalytics("your-api-key"));
 * ```
 */
export function oakAnalytics(apiKey: string, config: Config = {}) {
	const {
		privacyLevel = 0,
		serverUrl = "https://www.apianalytics-server.com/",
		getUserID = () => null,
	} = config;

	const analytics = new Analytics(apiKey, serverUrl, privacyLevel);

	return async (ctx: Context, next: Next): Promise<void> => {
		const start = performance.now();
		await next();

		analytics.logRequest({
			hostname: ctx.request.url.hostname,
			ip_address: getIPAddress(ctx, privacyLevel),
			user_agent: ctx.request.headers.get("user-agent") ?? "",
			path: ctx.request.url.pathname,
			status: ctx.response.status ?? 200,
			method: ctx.request.method,
			response_time: Math.round(performance.now() - start),
			user_id: getUserID(ctx),
			created_at: new Date().toISOString(),
		});
	};
}
