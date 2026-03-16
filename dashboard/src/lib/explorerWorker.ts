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
let lastQuery = '';
let lastSearchResult: RequestsData = [];

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

function search(query: string): RequestsData {
	if (!query) {
		lastQuery = '';
		lastSearchResult = sidebarFiltered;
		return sidebarFiltered;
	}
	// If the new query extends the previous one, the result can only be a subset —
	// search within the cached result instead of the full sidebarFiltered set.
	const base = lastQuery.length > 0 && query.startsWith(lastQuery)
		? lastSearchResult
		: sidebarFiltered;
	const result = applySearch(base, query, cachedUserAgents);
	lastQuery = query;
	lastSearchResult = result;
	return result;
}

function filterAndSearch(filter: Filter, query: string): RequestsData {
	sidebarFiltered = applyFilter(cachedRequests, filter);
	// Sidebar changed — incremental cache is no longer valid.
	lastQuery = '';
	lastSearchResult = [];
	return search(query);
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
		console.log('[Worker] filter message received, timespan:', msg.filter.timespan);
		const filtered = filterAndSearch(msg.filter, msg.query);
		console.log('[Worker] filtered count:', filtered.length);
		self.postMessage({ type: 'filtered', filtered });
	} else if (msg.type === 'search') {
		const filtered = search(msg.query);
		self.postMessage({ type: 'filtered', filtered });
	}
};
