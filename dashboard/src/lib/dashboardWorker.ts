/// <reference lib="webworker" />
import { getPeriodData, getHostnames } from '$lib/periodFilter';
import type { DashboardSettings } from '$lib/settings';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; settings: DashboardSettings }
	| { type: 'filter'; settings: DashboardSettings };

let cachedRequests: RequestsData | null = null;
let cachedHostnames: string[] = [];

self.onmessage = (e: MessageEvent<WorkerMessage>) => {
	const msg = e.data;

	if (msg.type === 'init') {
		cachedRequests = msg.requests;
		cachedHostnames = getHostnames(cachedRequests);
	}

	if (!cachedRequests) return;

	const { current, previous } = getPeriodData(cachedRequests, msg.settings);
	self.postMessage({ current, previous, hostnames: cachedHostnames });
};
