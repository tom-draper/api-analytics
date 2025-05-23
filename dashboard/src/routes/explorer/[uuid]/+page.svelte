<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import generateDemoData from '$lib/demo';
	import formatUUID from '$lib/uuid';
	import { ColumnIndex, pageSize } from '$lib/consts';
	import { getServerURL } from '$lib/url';
	import { dataStore } from '$lib/dataStore';
	import Navigation from '$components/explorer/navigation/Navigation.svelte';
	import Viewer from '$components/explorer/Viewer.svelte';
	import { statusBad, statusError, statusRedirect, statusSuccess } from '$lib/status';
	import { defaultFilter, type Filter } from '$lib/filter';
	import { nextDay, toDay } from '$lib/date';

	const userID = formatUUID($page.params.uuid);

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
				console.log(data);
			}
		} catch (e) {
			console.log(e);
		}

		return data;
	}

	async function fetchAdditionalPages() {
		let page = 2;
		let requests: number;
		do {
			requests = await fetchAdditionalPage(page);
			page++;
		} while (requests === pageSize);
	}

	async function fetchAdditionalPage(page: number) {
		const url = getServerURL();

		try {
			const response = await fetch(`${url}/api/requests/${userID}/${page}`, {
				signal: AbortSignal.timeout(250000),
				keepalive: true
			});
			if (response.status !== 200) {
				return 0;
			}

			const body = await response.json();
			if (body.requests.length <= 0) {
				return 0;
			}

			data.userAgents = { ...data.userAgents, ...body.user_agents };

			parseDates(body.requests);
			sortByTime(body.requests);

			data.requests = data.requests.concat(body.requests);

			dataStore.set(data);
			console.log(data);

			return body.requests.length;
		} catch (e) {
			console.log(e);
			return 0;
		}
	}

	function isDemo() {
		return $page.params.uuid === 'demo';
	}

	async function getDashboardData() {
		if (isDemo()) {
			return await getDemoData();
		}

		return await fetchData();
	}

	// async function getDemoData() {
	// 	return new Promise<DashboardData>((resolve) => {
	// 		const data = generateDemoData();
	// 		setTimeout(() => {
	// 			resolve(data);
	// 		});
	// 	});
	// }
	async function getDemoData() {
		return new Promise<DashboardData>((resolve) => {
			const worker = new Worker(new URL('$lib/worker.ts', import.meta.url), {
				type: 'module'
			});

			worker.onmessage = (event) => {
				resolve(event.data);
				worker.terminate(); // Cleanup worker after use
			};

			worker.postMessage(null);
		});
	}

	function parseDates(data: RequestsData) {
		for (let i = 0; i < data.length; i++) {
			data[i][ColumnIndex.CreatedAt] = new Date(data[i][ColumnIndex.CreatedAt]);
		}
	}

	function sortByTime(data: RequestsData) {
		data.sort((a, b) => {
			return a[ColumnIndex.CreatedAt].getTime() - b[ColumnIndex.CreatedAt].getTime();
		});
	}

	function applyFilter(data: RequestsData, filter: Filter) {
		const filteredData: RequestsData = [];
		const startDate = toDay(new Date(filter.timespan[0]));
		const endDate = toDay(nextDay(new Date(filter.timespan[1])));
		for (const row of data) {
			const status = row[ColumnIndex.Status];
			if (
				((filter.status.success && statusSuccess(status)) ||
					(filter.status.redirect && statusRedirect(status)) ||
					(filter.status.client && statusBad(status)) ||
					(filter.status.server && statusError(status))) &&
				row[ColumnIndex.CreatedAt] >= startDate &&
				row[ColumnIndex.CreatedAt] <= endDate &&
				filter.methods[row[ColumnIndex.Method]] &&
				filter.hostnames[row[ColumnIndex.Hostname]]
			) {
				filteredData.push(row);
			}
		}

		return filteredData;
	}

	function resetFilter() {
		filter = defaultFilter(data.requests);
	}

	let data: DashboardData;
	let filteredRequests: RequestsData;
	let filter: Filter;

	$: if (data && filter) {
		filteredRequests = applyFilter(data.requests, filter);
	}

	onMount(async () => {
		dataStore.subscribe((value) => {
			if (value) {
				data = value;
			}
		});

		if (!data) {
			data = await getDashboardData();

			if (data.requests.length === pageSize) {
				// Fetch page 2 and onwards if initial fetch didn't get all data
				fetchAdditionalPages();
			}

			parseDates(data.requests);
			sortByTime(data.requests);

			resetFilter();
		}
	});
</script>

<main>
	<div class="flex">
		<Navigation bind:data bind:filteredRequests bind:filter />
		<Viewer {data} filteredData={filteredRequests} />
	</div>
</main>
