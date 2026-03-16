<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { untrack } from 'svelte';
	import { get } from 'svelte/store';
	import formatUUID from '$lib/uuid';
	import { pageSize } from '$lib/consts';
	import { getServerURL } from '$lib/url';
	import { fetchPageRaw } from '$lib/fetchRequests';
	import { dataStore } from '$lib/dataStore';
	import Navigation from '$components/explorer/navigation/Navigation.svelte';
	import Viewer from '$components/explorer/Viewer.svelte';
	import { type Filter } from '$lib/filter';

	const userID = formatUUID(page.params.uuid);

	type FilterBounds = { timespan: [number, number]; rt: [number, number] };

	let data = $state.raw<DashboardData | undefined>(undefined);
	let filteredRequests = $state.raw<RequestsData>([]);
	let filter = $state<Filter | undefined>(undefined);
	let filterBounds = $state.raw<FilterBounds | null>(null);
	let initialFilter = $state.raw<Filter | null>(null);
	let worker = $state.raw<Worker | undefined>(undefined);

	// When data changes: send full requests to worker for parse/sort/filter
	$effect(() => {
		if (!worker || !data) return;
		worker.postMessage({
			type: 'init',
			requests: data.requests,
			filter: untrack(() => (filter ? $state.snapshot(filter) : null))
		});
	});

	// When filter changes: re-filter using cached requests in worker
	$effect(() => {
		const f = $state.snapshot(filter);
		const w = untrack(() => worker);
		if (!w || !f || !untrack(() => data)) return;
		w.postMessage({ type: 'filter', filter: f });
	});

	function resetFilter() {
		if (initialFilter) {
			filter = structuredClone(initialFilter);
		}
	}

	async function fetchData() {
		const url = getServerURL();
		let data: DashboardData = { requests: [], userAgents: {} };
		try {
			const response = await fetch(`${url}/api/requests/${userID}/1`, {
				signal: AbortSignal.timeout(250000),
				keepalive: true
			});
			const body = await response.json();
			if (response.ok && response.status === 200) {
				data = { requests: body.requests, userAgents: body.user_agents };
				dataStore.set(data);
			}
		} catch (e) {
			console.log(e);
		}
		return data;
	}

	async function fetchAdditionalPages() {
		let pageNum = 2;
		let count: number;
		do {
			count = await fetchAdditionalPage(pageNum);
			pageNum++;
		} while (count === pageSize);
	}

	async function fetchAdditionalPage(pageNum: number) {
		const body = await fetchPageRaw(userID, pageNum);
		if (!body) return 0;
		data = {
			requests: data!.requests.concat(body.requests),
			userAgents: { ...data!.userAgents, ...body.user_agents }
		};
		return body.requests.length;
	}

	function isDemo() {
		return page.params.uuid === 'demo';
	}

	async function getDashboardData() {
		if (isDemo()) return await getDemoData();
		return await fetchData();
	}

	async function getDemoData() {
		return new Promise<DashboardData>((resolve) => {
			const w = new Worker(new URL('$lib/worker.ts', import.meta.url), { type: 'module' });
			w.onmessage = (event) => {
				resolve(event.data);
				w.terminate();
			};
			w.postMessage(null);
		});
	}

	onMount(async () => {
		const w = new Worker(new URL('$lib/explorerWorker.ts', import.meta.url), { type: 'module' });
		w.onmessage = (e) => {
			const msg = e.data;
			if (msg.type === 'ready') {
				initialFilter = msg.filter;
				filterBounds = { timespan: msg.filter.timespan, rt: [msg.rtMin, msg.rtMax] };
				filter = structuredClone(msg.filter);
			} else if (msg.type === 'filtered') {
				filteredRequests = msg.filtered;
			}
		};
		worker = w;

		const storeData = get(dataStore);
		if (storeData && storeData.requests.length > 0) {
			data = storeData;
		} else {
			const dashboardData = await getDashboardData();
			if (dashboardData.requests.length === pageSize) {
				fetchAdditionalPages();
			}
			data = dashboardData;
			dataStore.set(dashboardData);
		}

		return () => w.terminate();
	});
</script>

<main>
	<div class="flex">
		<Navigation
			bind:filter
			{filteredRequests}
			totalCount={data?.requests.length ?? 0}
			{filterBounds}
			{resetFilter}
		/>
		<Viewer {filteredRequests} totalCount={data?.requests.length ?? 0} />
	</div>
</main>
