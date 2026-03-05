/// <reference lib="webworker" />
import { getPeriodData, getHostnames } from '$lib/periodFilter';
import { aggregate } from '$lib/aggregate';
import { ColumnIndex } from '$lib/consts';
import type { DashboardSettings } from '$lib/settings';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; userAgents: UserAgents; settings: DashboardSettings }
	| { type: 'filter'; settings: DashboardSettings }
	| { type: 'export' };

let cachedRequests: RequestsData | null = null;
let cachedUserAgents: UserAgents = {};
let cachedHostnames: string[] = [];
let cachedCurrent: RequestsData = [];

self.onmessage = (e: MessageEvent<WorkerMessage>) => {
	const msg = e.data;

	if (msg.type === 'export') {
		self.postMessage({ type: 'export', current: cachedCurrent });
		return;
	}

	if (msg.type === 'init') {
		const requests = msg.requests;
		for (let i = 0; i < requests.length; i++) {
			requests[i][ColumnIndex.CreatedAt] = new Date(requests[i][ColumnIndex.CreatedAt] as string);
		}
		requests.sort((a, b) =>
			(a[ColumnIndex.CreatedAt] as Date).getTime() - (b[ColumnIndex.CreatedAt] as Date).getTime()
		);
		cachedRequests = requests;
		cachedUserAgents = msg.userAgents;
		cachedHostnames = getHostnames(cachedRequests);
	}

	if (!cachedRequests) return;

	const { current, previous } = getPeriodData(cachedRequests, msg.settings);
	cachedCurrent = current;
	const aggregated = aggregate(current, previous, msg.settings);
	self.postMessage({ aggregated, hostnames: msg.type === 'init' ? cachedHostnames : undefined });
};
