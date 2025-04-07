<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Requests from '$components/dashboard/Requests.svelte';
	import Logo from '$components/dashboard/Logo.svelte';
	import ResponseTimes from '$components/dashboard/ResponseTimes.svelte';
	import Users from '$components/dashboard/Users.svelte';
	import Endpoints from '$components/dashboard/Endpoints.svelte';
	import SuccessRate from '$components/dashboard/SuccessRate.svelte';
	import Activity from '$components/dashboard/activity/Activity.svelte';
	import Version from '$components/dashboard/Version.svelte';
	import UsageTime from '$components/dashboard/UsageTime.svelte';
	import Location from '$components/dashboard/Location.svelte';
	import Device from '$components/dashboard/device/Device.svelte';
	import { dateInPeriod, dateInPrevPeriod } from '$lib/period';
	import generateDemoData from '$lib/demo';
	import formatUUID from '$lib/uuid';
	import Settings from '$components/dashboard/Settings.svelte';
	import { initSettings, type DashboardSettings } from '$lib/settings';
	import type { NotificationState } from '$lib/notification';
	import Notification from '$components/dashboard/Notification.svelte';
	import exportCSV from '$lib/exportData';
	import { ColumnIndex, columns, pageSize } from '$lib/consts';
	import Error from '$components/Error.svelte';
	import TopUsers from '$components/dashboard/TopUsers.svelte';
	import { getServerURL } from '$lib/url';
	import Navigation from '$components/dashboard/Navigation.svelte';
	import { userTargeted } from '$lib/user';
	import { dataStore } from '$lib/dataStore';
	import Health from '$components/dashboard/health/Health.svelte';
	import { periodParamToPeriod } from '$lib/params';
	import Referrer from '$components/dashboard/Referrer.svelte';

	const userID = formatUUID($page.params.uuid);

	function getPeriodData(data: RequestsData, settings: DashboardSettings) {
		const inRange = getInRange();
		const inPrevRange = getInPrevRange();

		const current = [];
		const previous = [];
		for (const request of data) {
			// Created inverted version of the if statement to reduce nesting
			const status = request[ColumnIndex.Status];
			const path = request[ColumnIndex.Path];
			const referrer = request[ColumnIndex.Path];
			const hostname = request[ColumnIndex.Hostname];
			const location = request[ColumnIndex.Location];
			const ipAddress = request[ColumnIndex.IPAddress];
			const customUserID = request[ColumnIndex.UserID];
			if (
				(settings.targetUser === null ||
					userTargeted(settings.targetUser, ipAddress, customUserID)) &&
				(!settings.disable404 || status !== 404) &&
				(settings.targetEndpoint.path === null || settings.targetEndpoint.path === path) &&
				(settings.targetEndpoint.status === null || settings.targetEndpoint.status === status) &&
				(settings.targetReferrer === null || settings.targetReferrer === referrer) &&
				(settings.targetLocation === null || settings.targetLocation === location) &&
				!isHiddenEndpoint(path) &&
				(settings.hostname === null || settings.hostname === hostname)
			) {
				const date = request[ColumnIndex.CreatedAt];
				if (inRange(date)) {
					current.push(request);
				} else if (inPrevRange(date)) {
					previous.push(request);
				}
			}
		}

		return { current, previous };
	}

	function allTimePeriod(_: Date) {
		return true;
	}

	function getInRange() {
		if (settings.period === 'All time') {
			return allTimePeriod;
		}

		return (date: Date) => {
			return dateInPeriod(date, settings.period);
		};
	}

	function getInPrevRange() {
		if (settings.period === 'All time') {
			return allTimePeriod;
		}

		return (date: Date) => {
			return dateInPrevPeriod(date, settings.period);
		};
	}

	function isHiddenEndpoint(endpoint: string) {
		const normalized = endpoint.replace(/^\/|\/$/g, ''); // Trim leading/trailing slashes
		return (
			settings.hiddenEndpoints.has(endpoint) ||
			settings.hiddenEndpoints.has('/' + normalized) ||
			settings.hiddenEndpoints.has(normalized) ||
			wildCardMatch(endpoint)
		);
	}

	function wildCardMatch(endpoint: string) {
		// Ensure endpoint has a trailing slash
		endpoint = endpoint.endsWith('/') ? endpoint : endpoint + '/';

		for (const hidden of settings.hiddenEndpoints) {
			if (!hidden.endsWith('*')) {
				continue;
			}

			let prefix = hidden.slice(0, -1); // Remove trailing '*'
			prefix = prefix.startsWith('/') ? prefix : '/' + prefix;
			prefix = prefix.endsWith('/') ? prefix : prefix + '/';

			if (endpoint.startsWith(prefix)) {
				return true;
			}
		}
		return false;
	}

	function sortedFrequencies(freq: ValueCount): string[] {
		return Object.entries(freq)
			.sort((a, b) => {
				return b[1] - a[1];
			})
			.map((value) => value[0]);
	}

	function getHostnames(data: RequestsData) {
		const hostnameFreq: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const hostname = data[i][ColumnIndex.Hostname];
			if (hostname === null || hostname === '') {
				continue;
			}
			if (hostname in hostnameFreq) {
				hostnameFreq[hostname]++;
			} else {
				hostnameFreq[hostname] = 1;
			}
		}

		return sortedFrequencies(hostnameFreq);
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

			data.userAgents = { ...data.userAgents, ...body.user_agents };

			parseDates(body.requests);
			sortByTime(body.requests);

			const mostRecent = body.requests[body.requests.length - 1][ColumnIndex.CreatedAt];
			if (dateInPeriod(mostRecent, settings.period)) {
				// Trigger dashboard re-render
				data.requests = data.requests.concat(body.requests);
			} else {
				// Avoid triggering dashboard re-render
				Object.assign(data.requests, data.requests.concat(body.requests));
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

	function getSettings() {
		const settings = initSettings();

		const period = $page.url.searchParams.get('period');
		if (period) {
			settings.period = periodParamToPeriod(period);
		}
		const hostname = $page.url.searchParams.get('hostname');
		if (hostname) {
			settings.hostname = hostname;
		}
		const location = $page.url.searchParams.get('location');
		if (location) {
			settings.targetLocation = location;
		}
		const path = $page.url.searchParams.get('path');
		if (path) {
			settings.targetEndpoint.path = path;
		}
		const status = $page.url.searchParams.get('status');
		if (status) {
			settings.targetEndpoint.status = parseInt(status);
		}
		const userID = $page.url.searchParams.get('userID');
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

		const ipAddress = $page.url.searchParams.get('ipAddress');
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

	let data: DashboardData;
	let settings: DashboardSettings = getSettings();
	let showSettings: boolean = false;
	let hostnames: string[];
	const notification: NotificationState = {
		message: '',
		style: 'error',
		show: false
	};
	let periodData: {
		current: RequestsData;
		previous: RequestsData;
	};
	let loading: boolean = true;
	let fetchStatus: { failed: boolean; status: number; message: string };
	let endpointsRendered: boolean = false;

	// If data or settings are changed, recalcualte data
	$: if (data) {
		periodData = getPeriodData(data.requests, settings);
	}

	$: if (data) {
		hostnames = getHostnames(data.requests);
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
			} else {
				loading = false;
			}

			parseDates(data.requests);
			sortByTime(data.requests);
		}
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
				<Version data={periodData.current} bind:endpointsRendered />
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
		<div class="spinner">
			<div class="loader"></div>
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
	}

	.loader {
		width: 40px;
		height: 40px;
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
