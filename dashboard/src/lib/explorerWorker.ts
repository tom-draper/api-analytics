/// <reference lib="webworker" />
import { ColumnIndex } from '$lib/consts';
import { defaultFilter, applyFilter, type Filter } from '$lib/filter';

type WorkerMessage =
	| { type: 'init'; requests: RequestsData; filter: Filter | null }
	| { type: 'filter'; filter: Filter };

let cachedRequests: RequestsData = [];

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

		if (msg.filter) {
			const filtered = applyFilter(cachedRequests, msg.filter);
			self.postMessage({ type: 'filtered', filtered });
		} else {
			const filter = defaultFilter(cachedRequests);
			const [rtMin, rtMax] = getResponseTimeRange(cachedRequests);
			self.postMessage({ type: 'ready', filter, rtMin, rtMax });
		}
	} else if (msg.type === 'filter') {
		const filtered = applyFilter(cachedRequests, msg.filter);
		self.postMessage({ type: 'filtered', filtered });
	}
};
