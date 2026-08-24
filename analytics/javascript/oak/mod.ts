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
	private flushPromise: Promise<void> | null = null;

	constructor(apiKey: string, serverUrl: string, privacyLevel: number) {
		this.apiKey = apiKey;
		this.serverUrl = serverUrl;
		this.privacyLevel = privacyLevel;

		// Flush even when traffic goes idle. Without this timer, a service that
		// receives only one request after startup retains it until a later request
		// happens to arrive more than a minute later.
		if (this.apiKey) {
			setInterval(() => {
				void this.flush();
			}, 60_000);
		}
	}

	async logRequest(requestData: RequestData): Promise<void> {
		if (!this.apiKey) return;

		this.requests.push(requestData);
		const now = new Date();
		if (now.getTime() - this.lastPosted.getTime() > 60000 && this.requests.length > 0) {
			await this.flush();
		}
	}

	private async flush(): Promise<void> {
		if (!this.apiKey || this.requests.length === 0) return;
		if (this.flushPromise) return this.flushPromise;

		const requestsToSend = this.requests;
		this.requests = [];
		this.flushPromise = (async () => {
			try {
				const response = await fetch(this.getServerEndpoint(), {
					method: "POST",
					body: JSON.stringify({
						api_key: this.apiKey,
						requests: requestsToSend,
						framework: "Oak",
						privacy_level: this.privacyLevel,
					}),
					headers: { "Content-Type": "application/json" },
				});
				if (!response.ok) {
					throw new Error(`Analytics server responded with status: ${response.status}`);
				}
				this.lastPosted = new Date();
			} catch {
				// Analytics must not affect the application, but a transient failure
				// must not discard the batch.
				this.requests = requestsToSend.concat(this.requests);
			} finally {
				this.flushPromise = null;
			}
		})();
		return this.flushPromise;
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
export function oakAnalytics(apiKey: string, config: Config = {}): (ctx: Context, next: Next) => Promise<void> {
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
