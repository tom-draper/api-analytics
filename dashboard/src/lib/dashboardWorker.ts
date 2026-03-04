/// <reference lib="webworker" />
import { getPeriodData, getHostnames } from '$lib/periodFilter';
import { aggregate } from '$lib/aggregate';
import type { DashboardSettings } from '$lib/settings';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; userAgents: UserAgents; settings: DashboardSettings }
	| { type: 'filter'; settings: DashboardSettings };

let cachedRequests: RequestsData | null = null;
let cachedUserAgents: UserAgents = {};
let cachedHostnames: string[] = [];

self.onmessage = (e: MessageEvent<WorkerMessage>) => {
	const msg = e.data;

	if (msg.type === 'init') {
		cachedRequests = msg.requests;
		cachedUserAgents = msg.userAgents;
		cachedHostnames = getHostnames(cachedRequests);
	}

	if (!cachedRequests) return;

	const { current, previous } = getPeriodData(cachedRequests, msg.settings);
	const aggregated = aggregate(current, previous, msg.settings);
	self.postMessage({ current, previous, aggregated, hostnames: cachedHostnames });
};
