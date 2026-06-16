/// <reference lib="webworker" />
import { getPeriodData, getHostnames } from '$lib/periodFilter';
import { aggregate } from '$lib/aggregate';
import { ColumnIndex } from '$lib/consts';
import type { DashboardSettings } from '$lib/settings';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; userAgents: UserAgents; settings: DashboardSettings }
	| {
			type: 'append';
			requests: RequestsData;
			userAgents: UserAgents;
			reaggregate: boolean;
			settings: DashboardSettings;
	  }
	| { type: 'filter'; settings: DashboardSettings }
	| { type: 'export' };

let cachedRequests: RequestsData | null = null;
let cachedUserAgents: UserAgents = {};
let cachedHostnames: string[] = [];
let cachedCurrent: RequestsData = [];

function parseDates(requests: RequestsData): void {
	for (let i = 0; i < requests.length; i++) {
		requests[i][ColumnIndex.CreatedAt] = new Date(
			requests[i][ColumnIndex.CreatedAt] as unknown as string
		);
	}
}

const byCreatedAt = (a: RequestsData[number], b: RequestsData[number]) =>
	(a[ColumnIndex.CreatedAt] as Date).getTime() - (b[ColumnIndex.CreatedAt] as Date).getTime();

/** Merge two arrays already sorted ascending by CreatedAt into a new sorted array. */
function mergeSorted(a: RequestsData, b: RequestsData): RequestsData {
	const out: RequestsData = new Array(a.length + b.length);
	let i = 0;
	let j = 0;
	let k = 0;
	while (i < a.length && j < b.length) {
		out[k++] = byCreatedAt(a[i], b[j]) <= 0 ? a[i++] : b[j++];
	}
	while (i < a.length) out[k++] = a[i++];
	while (j < b.length) out[k++] = b[j++];
	return out;
}

self.onmessage = (e: MessageEvent<WorkerMessage>) => {
	const msg = e.data;

	if (msg.type === 'export') {
		self.postMessage({ type: 'export', current: cachedCurrent });
		return;
	}

	if (msg.type === 'init') {
		const requests = msg.requests;
		parseDates(requests);
		requests.sort(byCreatedAt);
		cachedRequests = requests;
		cachedUserAgents = msg.userAgents;
		cachedHostnames = getHostnames(cachedRequests);
	}

	if (msg.type === 'append') {
		const rows = msg.requests;
		parseDates(rows);
		rows.sort(byCreatedAt);
		cachedRequests = cachedRequests ? mergeSorted(cachedRequests, rows) : rows;
		// New object reference so downstream user-agent set caches invalidate
		cachedUserAgents = { ...cachedUserAgents, ...msg.userAgents };
		// Out-of-period pages are cached for later period switches but don't
		// affect the current view, so skip the re-aggregation + re-render.
		if (!msg.reaggregate) return;
		cachedHostnames = getHostnames(cachedRequests);
	}

	if (!cachedRequests) return;

	const { current, previous } = getPeriodData(cachedRequests, msg.settings, cachedUserAgents);
	cachedCurrent = current;
	const aggregated = aggregate(current, previous, msg.settings);
	const sendHostnames = msg.type === 'init' || (msg.type === 'append' && msg.reaggregate);
	self.postMessage({ aggregated, hostnames: sendHostnames ? cachedHostnames : undefined });
};
