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
	import Search from '$components/explorer/Search.svelte';
	import { type Filter } from '$lib/filter';

	const userID = formatUUID(page.params.uuid);

	type FilterBounds = { timespan: [number, number]; rt: [number, number] };

	let data = $state.raw<DashboardData | undefined>(undefined);
	let filteredRequests = $state.raw<RequestsData>([]);
	let filter = $state<Filter | undefined>(undefined);
	let filterBounds = $state.raw<FilterBounds | null>(null);
	let initialFilter = $state.raw<Filter | null>(null);
	let worker = $state.raw<Worker | undefined>(undefined);
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;
	let searching = $state(false);

	const filtersActive = $derived(
		filter !== undefined &&
			initialFilter !== null &&
			(filter.timespan[0] !== initialFilter.timespan[0] ||
				filter.timespan[1] !== initialFilter.timespan[1] ||
				!filter.status.success ||
				!filter.status.redirect ||
				!filter.status.client ||
				!filter.status.server ||
				Object.values(filter.methods).some((v) => !v) ||
				Object.values(filter.hostnames).some((v) => !v) ||
				Object.values(filter.locations).some((v) => !v) ||
				Object.values(filter.referrers).some((v) => !v) ||
				filter.responseTime[0] !== 0 ||
				filter.responseTime[1] !== Infinity ||
				filter.paths.size > 0)
	);

	// When data changes: send full requests to worker for parse/sort/filter
	$effect(() => {
		if (!worker || !data) return;
		worker.postMessage({
			type: 'init',
			requests: data.requests,
			userAgents: data.userAgents,
			filter: untrack(() => (filter ? $state.snapshot(filter) : null)),
			query: untrack(() => searchQuery)
		});
	});

	// When filter changes: re-filter using cached requests in worker
	$effect(() => {
		const f = $state.snapshot(filter);
		const w = untrack(() => worker);
		if (!w || !f || !untrack(() => data)) return;
		searching = true;
		w.postMessage({ type: 'filter', filter: f, query: untrack(() => searchQuery) });
	});

	// When search query changes: debounce then send to worker
	$effect(() => {
		const q = searchQuery;
		const w = untrack(() => worker);
		if (!w || !untrack(() => data)) return;
		clearTimeout(searchTimeout);
		searching = true;
		searchTimeout = setTimeout(() => w.postMessage({ type: 'search', query: q }), 50);
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
				requestAnimationFrame(() => { searching = false; });
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
	<header class="fixed left-0 right-0 top-0 z-10 flex h-[52px] items-center border-b border-[var(--border)]">
		<Search bind:value={searchQuery} loading={searching} />
	</header>
	<div class="flex mt-[52px] h-[calc(100vh-52px)] overflow-hidden">
		<Navigation
			bind:filter
			{filtersActive}
			{filterBounds}
			{resetFilter}
		/>
		<Viewer {filteredRequests} totalCount={data?.requests.length ?? 0} />
	</div>
</main>
