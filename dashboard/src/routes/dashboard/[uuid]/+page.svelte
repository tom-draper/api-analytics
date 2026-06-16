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
	import Versions from '$components/dashboard/Versions.svelte';
	import UsageTime from '$components/dashboard/UsageTime.svelte';
	import DaysOfWeek from '$components/dashboard/DaysOfWeek.svelte';
	import Locations from '$components/dashboard/Locations.svelte';
	import Device from '$components/dashboard/device/Device.svelte';
	import { dateInPeriod } from '$lib/period';
	import generateDemoData from '$lib/demo';
	import formatUUID from '$lib/uuid';
	import Settings from '$components/dashboard/Settings.svelte';
	import { initSettings, parseSettingsFromURL, type DashboardSettings } from '$lib/settings';
	import type { NotificationState } from '$lib/notification';
	import Notification from '$components/dashboard/Notification.svelte';
	import exportCSV from '$lib/exportData';
	import { ColumnIndex, columns, loadingMessages, pageSize } from '$lib/consts';
	import Error from '$components/Error.svelte';
	import TopUsers from '$components/dashboard/TopUsers.svelte';
	import { fetchPage, fetchPageRaw } from '$lib/fetchRequests';
	import Navigation from '$components/dashboard/Navigation.svelte';
	import { dataStore } from '$lib/dataStore';
	import Referrers from '$components/dashboard/Referrers.svelte';
	import UserIDList from '$components/dashboard/UserIDList.svelte';
	import Loading from '$components/Loading.svelte';
	import { get } from 'svelte/store';
	import { untrack } from 'svelte';
	import type { AggregatedData } from '$lib/aggregate';

	const userID = formatUUID(page.params.uuid ?? '');

	async function fetchData() {
		let data: DashboardData = { requests: [], userAgents: {} };
		const result = await fetchPage(userID, 1);
		if (result.ok) {
			data = { requests: result.body.requests, userAgents: result.body.user_agents };
			dataStore.set(data);
		} else {
			fetchStatus = { failed: true, message: result.message, status: result.status };
		}
		return data;
	}

	async function fetchAdditionalPages() {
		let page = 2;
		let requests: number;
		do {
			requests = await fetchAdditionalPage(page);
			page++;
		} while (requests >= pageSize);

		loading = false;
	}

	async function fetchAdditionalPage(page: number) {
		const body = await fetchPageRaw(userID, page);
		if (!body || !data) return 0;

		const mostRecent = new Date(body.requests[body.requests.length - 1][ColumnIndex.CreatedAt]);
		const inPeriod = dateInPeriod(mostRecent, settings.period);

		// Send only the new page to the worker; it merges into its cache and
		// re-aggregates only when the page falls within the current period.
		worker?.postMessage({
			type: 'append',
			requests: body.requests,
			userAgents: body.user_agents,
			reaggregate: inPeriod,
			settings: $state.snapshot(settings)
		});

		// Keep the local copy current for the template (userAgents, request count).
		// In-period pages reassign data so dependent components re-render; out-of-period
		// pages mutate in place to avoid a re-render they wouldn't affect.
		Object.assign(data.userAgents, body.user_agents);
		if (inPeriod) {
			data = { ...data, requests: data.requests.concat(body.requests) };
		} else {
			data.requests.push(...body.requests);
		}

		dataStore.set(data);
		return body.requests.length;
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

	let data = $state.raw<DashboardData | undefined>(undefined);
	let settings = $state(parseSettingsFromURL(page.url.searchParams));
	let showSettings = $state(false);
	let hostnames = $state<string[]>([]);
	let notification = $state<NotificationState>({
		message: '',
		style: 'error',
		show: false
	});
	let aggregated = $state.raw<AggregatedData | undefined>(undefined);
	let loading = $state(true);
	let fetchStatus = $state<
		{ failed: boolean; status: number | null; message: string } | undefined
	>(undefined);
	let worker = $state.raw<Worker | undefined>(undefined);

	// Sends the initial dataset to the worker for caching + first aggregation.
	function postInit() {
		if (!worker || !data) return;
		worker.postMessage({
			type: 'init',
			requests: data.requests,
			userAgents: data.userAgents,
			settings: $state.snapshot(settings)
		});
	}

	// When settings change: re-filter using the requests already cached in the worker.
	// worker/data are untracked so this only fires on settings changes; postInit and
	// the per-page appends handle aggregation during the initial load.
	$effect(() => {
		const s = $state.snapshot(settings);
		const w = untrack(() => worker);
		if (!w || !untrack(() => data)) return;
		w.postMessage({ type: 'filter', settings: s });
	});

	async function loadData() {
		const storeData = get(dataStore);
		if (storeData && storeData.requests.length > 0) {
			data = storeData;
			postInit();
			loading = false;
			return;
		}

		const dashboardData = await getDashboardData();
		data = dashboardData;
		dataStore.set(dashboardData);
		postInit();

		if (dashboardData.requests.length >= pageSize) {
			fetchAdditionalPages();
		} else {
			loading = false;
		}
	}

	onMount(() => {
		const w = new Worker(new URL('$lib/dashboardWorker.ts', import.meta.url), {
			type: 'module'
		});
		w.onmessage = (e) => {
			const msg = e.data;
			if (msg.type === 'export') {
				exportCSV(msg.current, columns, data!.userAgents);
				return;
			}
			if (msg.hostnames !== undefined) hostnames = msg.hostnames;
			// Defer the render by one rAF so the CSS animation isn't starved
			// by the synchronous Plotly chart initialisation that follows
			requestAnimationFrame(() => {
				aggregated = msg.aggregated;
			});
		};
		worker = w;

		loadData();

		return () => w.terminate();
	});
</script>

<Settings
	bind:show={showSettings}
	bind:settings
	exportCSV={() => {
		worker?.postMessage({ type: 'export' });
	}}
/>
<Notification state={notification} />
{#if aggregated && data && data.requests.length > 0}
	<div class="dashboard">
		<Navigation bind:settings bind:showSettings bind:hostnames />

		<div class="dashboard-content">
			<div class="left">
				<div class="row">
					<Logo {loading} />
					<SuccessRate
						rate={aggregated.successRate}
						buckets={aggregated.successBuckets}
					/>
				</div>
				<div class="row">
					<Requests
						buckets={aggregated.requestBuckets}
						count={aggregated.requestCount}
						prevCount={aggregated.prevRequestCount}
						firstDate={aggregated.firstRequestDate}
						lastDate={aggregated.lastRequestDate}
						period={settings.period}
					/>
					<Users
						buckets={aggregated.userBuckets}
						count={aggregated.userCount}
						prevCount={aggregated.prevUserCount}
						firstDate={aggregated.firstRequestDate}
						lastDate={aggregated.lastRequestDate}
						period={settings.period}
					/>
				</div>
				<ResponseTimes
					freqTimes={aggregated.rtFreqTimes}
					freqCounts={aggregated.rtFreqCounts}
					LQ={aggregated.rtLQ}
					median={aggregated.rtMedian}
					UQ={aggregated.rtUQ}
				/>
				<Endpoints
					endpointFreq={aggregated.endpointFreq}
					bind:targetPath={settings.targetEndpoint.path}
					bind:targetStatus={settings.targetEndpoint.status}
				/>
				<Versions
					versionCount={aggregated.versionCount}
					hasMultiple={aggregated.versionHasMultiple}
					bind:targetVersion={settings.targetVersion}
				/>
			</div>
			<div class="right">
				<Activity activityBuckets={aggregated.activityBuckets} period={aggregated.period} />
				<div class="grid-row">
					<Locations
						locationBars={aggregated.locationBars}
						bind:targetLocation={settings.targetLocation}
					/>
					<Device
						uaIdCount={aggregated.uaIdCount}
						userAgents={data.userAgents}
						bind:targetClient={settings.targetClient}
						bind:targetDeviceType={settings.targetDeviceType}
						bind:targetOS={settings.targetOS}
					/>
				</div>
				<div class="flex chart-row">
					<div class="flex-grow">
						<UsageTime
							hourlyBuckets={aggregated.hourlyBuckets}
							bind:targetHour={settings.targetHour}
						/>
						<DaysOfWeek
							weekdayBuckets={aggregated.weekdayBuckets}
							bind:targetWeekday={settings.targetWeekday}
						/>
						<TopUsers
							users={aggregated.topUsers}
							userIDActive={aggregated.topUserIDActive}
							locationsActive={aggregated.topLocationsActive}
							bind:targetUser={settings.targetUser}
						/>
					</div>
					<div class="referrer-col">
						{#if aggregated.referrerAvailable}
							<Referrers
								referrerBars={aggregated.referrerBars}
								bind:targetReferrer={settings.targetReferrer}
							/>
						{/if}
					</div>
				</div>
			</div>
		</div>
	</div>
{:else if aggregated && data && data.requests.length <= 0}
	<Error status={400} message="" />
{:else if fetchStatus && fetchStatus.failed}
	<Error status={fetchStatus.status} message={fetchStatus.message} />
{:else}
	<div class="placeholder">
		<div class="spinner-container">
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
		width: min(500px, 90vw);
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
		color: var(--dim-text);
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
		.chart-row {
			flex-direction: column;
		}
		.referrer-col {
			order: -1;
			width: 100%;
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
	@media screen and (max-width: 1070px) {
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
