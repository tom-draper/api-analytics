import { getServerURL } from '$lib/url';

export type PageBody = {
	requests: any[];
	user_agents: Record<number, string>;
};

export type FetchResult =
	| { ok: true; body: PageBody }
	| { ok: false; status: number | null; message: string };

/**
 * Fetches a single page of requests from the API. Returns a discriminated
 * result: the parsed body on success, or the status/message on failure.
 */
export async function fetchPage(userID: string, page: number): Promise<FetchResult> {
	const url = getServerURL();
	try {
		const response = await fetch(`${url}/api/requests/${userID}/${page}`, {
			signal: AbortSignal.timeout(250000),
			keepalive: true
		});
		const body = await response.json();
		if (response.ok && response.status === 200) {
			return { ok: true, body };
		}
		return { ok: false, status: body?.status ?? null, message: body?.message ?? '' };
	} catch {
		return { ok: false, status: 500, message: 'Internal server error.' };
	}
}

/**
 * Fetches a single page of requests. Returns the parsed body on success, or
 * null if the request failed or returned no data.
 */
export async function fetchPageRaw(userID: string, page: number): Promise<PageBody | null> {
	const result = await fetchPage(userID, page);
	if (!result.ok || result.body.requests.length <= 0) return null;
	return result.body;
}
