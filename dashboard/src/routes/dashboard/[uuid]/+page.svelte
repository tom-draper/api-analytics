<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Requests from '$components/dashboard/Requests.svelte';
	import Logo from '$components/dashboard/Logo.svelte';
	import ResponseTimes from '$components/dashboard/ResponseTimes.svelte';
	import Users from '$components/dashboard/Users.svelte';
	import Endpoints from '$components/dashboard/endpoints/Endpoints.svelte';
	import SuccessRate from '$components/dashboard/SuccessRate.svelte';
	import Activity from '$components/dashboard/activity/Activity.svelte';
	import Version from '$components/dashboard/Version.svelte';
	import UsageTime from '$components/dashboard/UsageTime.svelte';
	import Location from '$components/dashboard/Location.svelte';
	import Device from '$components/dashboard/device/Device.svelte';
	import { dateInPeriod, isPeriod } from '$lib/period';
	import generateDemoData from '$lib/demo';
	import formatUUID from '$lib/uuid';
	import Settings from '$components/dashboard/Settings.svelte';
	import { initSettings, type DashboardSettings } from '$lib/settings';
	import type { NotificationState } from '$lib/notification';
	import Notification from '$components/dashboard/Notification.svelte';
	import exportCSV from '$lib/exportData';
	import { ColumnIndex, columns, loadingMessages, pageSize } from '$lib/consts';
	import Error from '$components/Error.svelte';
	import TopUsers from '$components/dashboard/TopUsers.svelte';
	import { getServerURL } from '$lib/url';
	import Navigation from '$components/dashboard/Navigation.svelte';
	import { dataStore } from '$lib/dataStore';
	import Health from '$components/dashboard/health/Health.svelte';
	import Referrer from '$components/dashboard/Referrer.svelte';
	import Loading from '$components/Loading.svelte';
	import { get } from 'svelte/store';
	import { untrack } from 'svelte';

	const userID = formatUUID(page.params.uuid);

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
			} else {
				fetchStatus = {
					failed: true,
					message: body.message || '',
					status: body.status || null
				};
			}
		} catch (e) {
			console.log(e);
			fetchStatus = {
				failed: true,
				message: 'Internal server error.',
				status: 500
			};
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

			Object.assign(data.userAgents, body.user_agents);

			parseDates(body.requests);
			sortByTime(body.requests);

			const mostRecent = body.requests[body.requests.length - 1][ColumnIndex.CreatedAt];
			if (dateInPeriod(mostRecent, settings.period)) {
				// Trigger dashboard re-render
				data = { ...data, requests: data.requests.concat(body.requests) };
			} else {
				// Avoid triggering dashboard re-render
				data.requests.push(...body.requests);
			}

			dataStore.set(data);
			console.log(data);

			return body.requests.length;
		} catch (e) {
			console.log(e);
			return 0;
		}
	}

	function isDemo() {
		return page.params.uuid === 'demo';
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

	function getSettings() {
		const settings = initSettings();

		const period = page.url.searchParams.get('period');
		if (period && isPeriod(period)) {
			settings.period = period;
		}
		const hostname = page.url.searchParams.get('hostname');
		if (hostname) {
			settings.hostname = hostname;
		}
		const location = page.url.searchParams.get('location');
		if (location) {
			settings.targetLocation = location;
		}
		const path = page.url.searchParams.get('path');
		if (path) {
			settings.targetEndpoint.path = path;
		}
		const status = page.url.searchParams.get('status');
		if (status) {
			settings.targetEndpoint.status = parseInt(status);
		}
		const userID = page.url.searchParams.get('userID');
		if (userID) {
			if (settings.targetUser === null) {
				settings.targetUser = {
					ipAddress: '',
					userID: '',
					composite: false
				};
			}
			settings.targetUser.userID = userID;
		}

		const ipAddress = page.url.searchParams.get('ipAddress');
		if (ipAddress) {
			if (settings.targetUser === null) {
				settings.targetUser = {
					ipAddress: '',
					userID: '',
					composite: false
				};
			}
			settings.targetUser.ipAddress = ipAddress;
		}

		return settings;
	}

	let data = $state.raw<DashboardData | undefined>(undefined);
	let settings = $state(getSettings());
	let showSettings = $state(false);
	let hostnames = $state<string[]>([]);
	let notification = $state<NotificationState>({
		message: '',
		style: 'error',
		show: false
	});
	let periodData = $state.raw<{ current: RequestsData; previous: RequestsData } | undefined>(undefined);
	let loading = $state(true);
	let fetchStatus = $state<{ failed: boolean; status: number; message: string } | undefined>(undefined);
	let worker = $state.raw<Worker | undefined>(undefined);

	// When data changes: send full requests to worker cache, then filter
	$effect(() => {
		if (!worker || !data) return;
		worker.postMessage({ type: 'init', requests: data.requests, settings: untrack(() => $state.snapshot(settings)) });
	});

	// When settings change: re-filter using cached requests in worker
	$effect(() => {
		const s = $state.snapshot(settings);
		if (!worker || !untrack(() => data)) return;
		worker.postMessage({ type: 'filter', settings: s });
	});

	onMount(async () => {
		const w = new Worker(new URL('$lib/dashboardWorker.ts', import.meta.url), { type: 'module' });
		w.onmessage = (e) => {
			const { current, previous, hostnames: h } = e.data;
			periodData = { current, previous };
			if (h !== undefined) hostnames = h;
		};
		worker = w;

		const storeData = get(dataStore);
		if (storeData && storeData.requests.length > 0) {
			parseDates(storeData.requests);
			sortByTime(storeData.requests);
			data = storeData;
			loading = false;
		} else {
			const dashboardData = await getDashboardData();

			if (dashboardData.requests.length === pageSize) {
				fetchAdditionalPages();
			} else {
				loading = false;
			}

			parseDates(dashboardData.requests);
			sortByTime(dashboardData.requests);
			data = dashboardData;
			dataStore.set(dashboardData);
		}

		return () => w.terminate();
	});
</script>

<Settings
	bind:show={showSettings}
	bind:settings
	exportCSV={() => {
		exportCSV(periodData.current, columns, data.userAgents);
	}}
/>
<Notification state={notification} />
{#if periodData && data.requests.length > 0}
	<div class="dashboard">
		<Navigation bind:settings bind:showSettings bind:hostnames />

		<div class="dashboard-content">
			<div class="left">
				<div class="row">
					<Logo bind:loading />
					<SuccessRate data={periodData.current} />
				</div>
				<div class="row">
					<Requests
						data={periodData.current}
						prevData={periodData.previous}
						period={settings.period}
					/>
					<Users
						data={periodData.current}
						prevData={periodData.previous}
						period={settings.period}
					/>
				</div>
				<ResponseTimes data={periodData.current} />
				<Endpoints
					data={periodData.current}
					bind:targetPath={settings.targetEndpoint.path}
					bind:targetStatus={settings.targetEndpoint.status}
					bind:ignoreParams={settings.ignoreParams}
				/>
				<Version data={periodData.current} />
			</div>
			<div class="right">
				<Activity data={periodData.current} period={settings.period} />
				<div class="grid-row">
					<Location data={periodData.current} bind:targetLocation={settings.targetLocation} />
					<Device data={periodData.current} userAgents={data.userAgents} />
				</div>
				<div class="flex">
					<div class="flex-grow">
						<!-- <Health data={periodData.current} /> -->
						<UsageTime data={periodData.current} />
						<TopUsers data={periodData.current} bind:targetUser={settings.targetUser} />
					</div>
					<div>
						<!-- <Referrer data={periodData.current} bind:targetReferrer={settings.targetReferrer} bind:ignoreParams={settings.ignoreParams}/> -->
					</div>
				</div>
			</div>
		</div>
	</div>
{:else if periodData && data.requests.length <= 0}
	<Error status={400} message="" />
{:else if fetchStatus && fetchStatus.failed}
	<Error status={fetchStatus.status} message={fetchStatus.message} />
{:else}
	<div class="placeholder">
		<div class="spinner-container">
			<!-- <div class="loader"></div> -->
			<div class="w-[25px]">
				<Loading />
			</div>
			<div class="loading-text-container">
				{#each loadingMessages as message, i}
					<div class="loading-text">{message}</div>
				{/each}
			</div>
		</div>
	</div>
{/if}

<style scoped>
	.dashboard {
		min-height: 90vh;
	}
	.dashboard {
		margin: 1.4em 5em 5em;
	}
	.dashboard-content {
		margin-top: 1.4em;
		display: flex;
		position: relative;
	}
	.row {
		display: flex;
		margin-bottom: 2em;
	}
	.grid-row {
		display: flex;
	}
	.left {
		margin: 0 2em;
	}
	.right {
		flex-grow: 1;
		margin-right: 2em;
	}

	.placeholder {
		min-height: 80vh;
		display: grid;
		place-items: center;
		text-align: center;
	}

	.spinner-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 3.8rem;
		width: 500px;
	}

	/* Multi-text animation container */
	.loading-text-container {
		position: relative;
		height: 1.5em;
		width: 100%;
		text-align: center;
		overflow: hidden;
	}

	/* Individual loading texts */
	.loading-text {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		font-size: 14px;
		color: #707070;
		opacity: 0;
		/* Ensure each text sits on its own GPU layer */
		will-change: opacity;
		transform: translateZ(0);
		animation-timing-function: ease-in-out;
	}
	.loading-text:nth-child(1) {
		animation: fadeText 28s infinite 0s;
	}
	.loading-text:nth-child(2) {
		animation: fadeText 28s infinite 3.5s;
	}
	.loading-text:nth-child(3) {
		animation: fadeText 28s infinite 7s;
	}
	.loading-text:nth-child(4) {
		animation: fadeText 28s infinite 10.5s;
	}
	.loading-text:nth-child(5) {
		animation: fadeText 28s infinite 14s;
	}
	.loading-text:nth-child(6) {
		animation: fadeText 28s infinite 17.5s;
	}
	.loading-text:nth-child(7) {
		animation: fadeText 28s infinite 21s;
	}
	.loading-text:nth-child(8) {
		animation: fadeText 28s infinite 24.5s;
	}

	@keyframes fadeText {
		0% {
			opacity: 0;
			transform: translateY(10px);
		}
		5% {
			opacity: 1;
			transform: translateY(0);
		}
		12.5% {
			opacity: 1;
			transform: translateY(0);
		}
		17.5% {
			opacity: 0;
			transform: translateY(-10px);
		}
		100% {
			opacity: 0;
			transform: translateY(-10px);
		}
	}

	@keyframes gradient-shift {
		0%,
		100% {
			background-position: 0% 50%; /* Start at the left */
		}
		50% {
			background-position: 100% 50%; /* End at the right */
		}
	}

	@media screen and (max-width: 1600px) {
		.grid-row {
			flex-direction: column;
		}
	}
	@media screen and (max-width: 1300px) {
		.dashboard {
			margin: 0;
		}
		.dashboard-content {
			margin: 1.4em 1em 3.5em;
		}
	}
	@media screen and (max-width: 1030px) {
		.dashboard-content {
			flex-direction: column;
		}
		.right,
		.left {
			margin: 0 2em;
		}
	}
	@media screen and (max-width: 660px) {
		.right,
		.left {
			margin: 0;
		}
	}
	@media screen and (max-width: 450px) {
		.dashboard-content {
			margin: 1.4em 1em 3.5em;
		}
	}
</style>
