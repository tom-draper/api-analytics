import { getServerURL } from '$lib/url';

export type PageBody = {
	requests: any[];
	user_agents: Record<number, string>;
};

/**
 * Fetches a single page of requests from the API.
 * Returns the parsed body on success, or null if the request failed or returned no data.
 */
export async function fetchPageRaw(userID: string, page: number): Promise<PageBody | null> {
	const url = getServerURL();
	try {
		const response = await fetch(`${url}/api/requests/${userID}/${page}`, {
			signal: AbortSignal.timeout(250000),
			keepalive: true
		});
		if (response.status !== 200) return null;
		const body = await response.json();
		if (body.requests.length <= 0) return null;
		return body;
	} catch {
		return null;
	}
}
