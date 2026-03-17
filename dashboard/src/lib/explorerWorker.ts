/// <reference lib="webworker" />
import { ColumnIndex } from '$lib/consts';
import { defaultFilter, applyFilter, type Filter } from '$lib/filter';
import { applySearch } from '$lib/search';
import { statusBad, statusError, statusRedirect, statusSuccess } from '$lib/status';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; userAgents: Record<number, string>; filter: Filter | null; query: string }
	| { type: 'filter'; filter: Filter; query: string }
	| { type: 'search'; query: string };

let cachedRequests: RequestsData = [];
let cachedUserAgents: Record<number, string> = {};
let sidebarFiltered: RequestsData = [];
let lastQuery = '';
let lastSearchResult: RequestsData = [];

type Bucket = { center: number; count: number };

function computeHistogram(data: RequestsData, min: number, max: number, col: number, n: number): Bucket[] {
	const range = max - min;
	if (range === 0) return [{ center: min, count: data.length }];
	const counts = new Array<number>(n).fill(0);
	for (const row of data) {
		const v = col === ColumnIndex.CreatedAt
			? (row[col] as Date).getTime()
			: row[col] as number;
		if (v == null) continue;
		counts[Math.min(n - 1, Math.floor((v - min) / range * n))]++;
	}
	return counts.map((count, i) => ({ center: min + (i + 0.5) / n * range, count }));
}

function computeCounts(data: RequestsData) {
	const methods: Record<number, number> = {};
	const hostnames: Record<string, number> = {};
	const locations: Record<string, number> = {};
	const referrers: Record<string, number> = {};
	let success = 0, redirect = 0, client = 0, server = 0;
	for (const row of data) {
		const status = row[ColumnIndex.Status] as number;
		if (statusSuccess(status)) success++;
		else if (statusRedirect(status)) redirect++;
		else if (statusBad(status)) client++;
		else if (statusError(status)) server++;
		const method = row[ColumnIndex.Method] as number;
		if (method != null) methods[method] = (methods[method] ?? 0) + 1;
		const hostname = row[ColumnIndex.Hostname] as string;
		if (hostname) hostnames[hostname] = (hostnames[hostname] ?? 0) + 1;
		const location = row[ColumnIndex.Location] as string;
		if (location) locations[location] = (locations[location] ?? 0) + 1;
		const referrer = row[ColumnIndex.Referrer] as string;
		if (referrer) referrers[referrer] = (referrers[referrer] ?? 0) + 1;
	}
	return { status: { success, redirect, client, server }, methods, hostnames, locations, referrers };
}

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

function filterAndSearch(filter: Filter, query: string): { filtered: RequestsData; counts: ReturnType<typeof computeCounts> } {
	sidebarFiltered = applyFilter(cachedRequests, filter);
	// Sidebar changed — incremental cache is no longer valid.
	lastQuery = '';
	lastSearchResult = [];
	return { filtered: search(query), counts: computeCounts(sidebarFiltered) };
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
			const { filtered, counts } = filterAndSearch(msg.filter, msg.query);
			self.postMessage({ type: 'filtered', filtered, counts });
		} else {
			sidebarFiltered = cachedRequests;
			const filter = defaultFilter(cachedRequests);
			const [rtMin, rtMax] = getResponseTimeRange(cachedRequests);
			const N = 60;
			const timespanBuckets = computeHistogram(cachedRequests, filter.timespan[0], filter.timespan[1], ColumnIndex.CreatedAt, N);
			const rtBuckets = computeHistogram(cachedRequests, rtMin, rtMax, ColumnIndex.ResponseTime, N);
			const counts = computeCounts(cachedRequests);
			self.postMessage({ type: 'ready', filter, rtMin, rtMax, timespanBuckets, rtBuckets, counts });
		}
	} else if (msg.type === 'filter') {
		const { filtered, counts } = filterAndSearch(msg.filter, msg.query);
		self.postMessage({ type: 'filtered', filtered, counts });
	} else if (msg.type === 'search') {
		const filtered = search(msg.query);
		self.postMessage({ type: 'filtered', filtered, counts: computeCounts(sidebarFiltered) });
	}
};
