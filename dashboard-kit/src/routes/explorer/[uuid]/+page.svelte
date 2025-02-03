<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import generateDemoData from '$lib/demo';
	import formatUUID from '$lib/uuid';
	import { ColumnIndex, columns } from '$lib/consts';
	import { getServerURL } from '$lib/url';
	import { dataStore } from '$lib/dataStore';

	const userID = formatUUID($page.params.uuid);

	async function fetchData() {
		const url = getServerURL();

		let data: DashboardData = { requests: [], user_agents: {} };
		try {
			const response = await fetch(`${url}/api/requests/${userID}/1`, {
				signal: AbortSignal.timeout(250000),
				keepalive: true
			});

			const body = await response.json();
			if (response.ok && response.status === 200) {
				data = body;
				dataStore.set(data);
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

		loading = false;
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

			userAgents = { ...userAgents, ...body.user_agents };

			parseDates(body.requests);
			sortByTime(body.requests);

			data = data.concat(body.requests);
			dataStore.set({requests: data, user_agents: userAgents});

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

	async function getDemoData() {
		return new Promise<DashboardData>((resolve) => {
			const data = generateDemoData();
			setTimeout(() => {
				resolve(data);
			});
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

	let data: RequestsData;
	let userAgents: UserAgents;
	let loading: boolean = true;
	const pageSize = 200_000;

	onMount(async () => {
		dataStore.subscribe((value) => {
			if (value) {
				data = value.requests;
                userAgents = value.user_agents;
			}
		});

		if (!data) {
			const dashboardData = await getDashboardData();
			data = dashboardData.requests;
			userAgents = dashboardData.user_agents;

			if (data.length === pageSize) {
				// Fetch page 2 and onwards if initial fetch didn't get all data
				fetchAdditionalPages();
			} else {
				loading = false;
			}

			parseDates(data);
			sortByTime(data);
		}

		console.log(data);
	});
</script>

<div>
	<h1>Explorer</h1>
</div>
