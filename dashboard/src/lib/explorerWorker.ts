/// <reference lib="webworker" />
import { ColumnIndex } from '$lib/consts';
import { defaultFilter, applyFilter, type Filter } from '$lib/filter';
import { applySearch } from '$lib/search';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; userAgents: Record<number, string>; filter: Filter | null; query: string }
	| { type: 'filter'; filter: Filter; query: string }
	| { type: 'search'; query: string };

let cachedRequests: RequestsData = [];
let cachedUserAgents: Record<number, string> = {};
let sidebarFiltered: RequestsData = [];

function getResponseTimeRange(data: RequestsData): [number, number] {
	let min = Infinity;
	let max = 0;
	for (const row of data) {
		const rt = row[ColumnIndex.ResponseTime] as number;
		if (rt < min) min = rt;
		if (rt > max) max = rt;
	}
	return [min === Infinity ? 0 : min, max];
}

function filterAndSearch(filter: Filter, query: string): RequestsData {
	sidebarFiltered = applyFilter(cachedRequests, filter);
	return query ? applySearch(sidebarFiltered, query, cachedUserAgents) : sidebarFiltered;
}

self.onmessage = (e: MessageEvent<WorkerMessage>) => {
	const msg = e.data;

	if (msg.type === 'init') {
		const requests = msg.requests as RequestsData;
		for (let i = 0; i < requests.length; i++) {
			requests[i][ColumnIndex.CreatedAt] = new Date(requests[i][ColumnIndex.CreatedAt] as string);
		}
		requests.sort((a, b) =>
			(a[ColumnIndex.CreatedAt] as Date).getTime() - (b[ColumnIndex.CreatedAt] as Date).getTime()
		);
		cachedRequests = requests;
		cachedUserAgents = msg.userAgents;

		if (msg.filter) {
			const filtered = filterAndSearch(msg.filter, msg.query);
			self.postMessage({ type: 'filtered', filtered });
		} else {
			sidebarFiltered = cachedRequests;
			const filter = defaultFilter(cachedRequests);
			const [rtMin, rtMax] = getResponseTimeRange(cachedRequests);
			self.postMessage({ type: 'ready', filter, rtMin, rtMax });
		}
	} else if (msg.type === 'filter') {
		const filtered = filterAndSearch(msg.filter, msg.query);
		self.postMessage({ type: 'filtered', filtered });
	} else if (msg.type === 'search') {
		const filtered = msg.query
			? applySearch(sidebarFiltered, msg.query, cachedUserAgents)
			: sidebarFiltered;
		self.postMessage({ type: 'filtered', filtered });
	}
};
