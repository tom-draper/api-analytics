<script lang="ts">
   import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Requests from '$lib/components/dashboard/Requests.svelte';
	import Logo from '$lib/components/dashboard/Logo.svelte';
	import ResponseTimes from '$lib/components/dashboard/ResponseTimes.svelte';
	import Users from '$lib/components/dashboard/Users.svelte';
	import Endpoints from '$lib/components/dashboard/Endpoints.svelte';
	import SuccessRate from '$lib/components/dashboard/SuccessRate.svelte';
	import Activity from '$lib/components/dashboard/activity/Activity.svelte';
	import Version from '$lib/components/dashboard/Version.svelte';
	import UsageTime from '$lib/components/dashboard/UsageTime.svelte';
	import Location from '$lib/components/dashboard/Location.svelte';
	import Device from '$lib/components/dashboard/device/Device.svelte';
	import { dateInPeriod, dateInPrevPeriod } from '$lib/period';
	import genDemoData from '$lib/demo';
	import formatUUID from '$lib/uuid';
    import Settings from "$lib/components/dashboard/Settings.svelte";
	import type { DashboardSettings, Period } from '$lib/settings';
	import { initSettings } from '$lib/settings';
	import type { NotificationState } from '$lib/notification';
	import Notification from '$lib/components/dashboard/Notification.svelte';
	import exportCSV from '$lib/exportData';
	import { ColumnIndex, columns } from '$lib/consts';
	import Error from '$lib/components/dashboard/Error.svelte';
	import TopUsers from '$lib/components/dashboard/TopUsers.svelte';
	import { getServerURL } from '$lib/url';
	import Navigation from '$lib/components/dashboard/Navigation.svelte';

	const userID = formatUUID($page.params.uuid);

	function allTimePeriod(_: Date) {
		return true;
	}

	function getPeriodData() {
		const inRange = getInRange();
		const inPrevRange = getInPrevRange();

		const current = [];
		const previous = [];
		for (let i = 0; i < data.length; i++) {
			// Created inverted version of the if statement to reduce nesting
			const request = data[i];
			const status = request[ColumnIndex.Status];
			const path = request[ColumnIndex.Path];
			const hostname = request[ColumnIndex.Hostname];
			const location = request[ColumnIndex.Location];
			const ipAddress = request[ColumnIndex.IPAddress];
			const customUserID = request[ColumnIndex.UserID];
			if (
				(settings.targetUser === null || settings.targetUser === `${ipAddress} ${customUserID}`) &&
				(!settings.disable404 || status !== 404) &&
				(settings.targetEndpoint.path === null || settings.targetEndpoint.path === path) &&
				(settings.targetEndpoint.status === null || settings.targetEndpoint.status === status) &&
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

	function setPeriodData() {
		periodData = getPeriodData();
	}

	function getInRange() {
		let inRange: (date: Date) => boolean;
		if (settings.period === 'All time') {
			inRange = allTimePeriod;
		} else {
			inRange = (date: Date) => {
				return dateInPeriod(date, settings.period);
			};
		}
		return inRange;
	}

	function getInPrevRange() {
		let inPrevRange: (date: Date) => boolean;
		if (settings.period === 'All time') {
			inPrevRange = allTimePeriod;
		} else {
			inPrevRange = (date) => {
				return dateInPrevPeriod(date, settings.period);
			};
		}
		return inPrevRange;
	}

	function isHiddenEndpoint(endpoint: string) {
		const firstChar = endpoint.charAt(0);
		const lastChar = endpoint.charAt(endpoint.length - 1);
		return (
			settings.hiddenEndpoints.has(endpoint) ||
			(firstChar === '/' && settings.hiddenEndpoints.has(endpoint.slice(1))) ||
			(lastChar === '/' && settings.hiddenEndpoints.has(endpoint.slice(0, -1))) ||
			(firstChar !== '/' && settings.hiddenEndpoints.has('/' + endpoint)) ||
			(lastChar !== '/' && settings.hiddenEndpoints.has(endpoint + '/')) ||
			wildCardMatch(endpoint)
		);
	}

	function wildCardMatch(endpoint: string) {
		if (endpoint.charAt(endpoint.length - 1) !== '/') {
			endpoint = endpoint + '/';
		}

		for (let value of settings.hiddenEndpoints) {
			if (value.charAt(value.length - 1) !== '*') {
				continue;
			}
			value = value.slice(0, value.length - 1); // Remove asterisk
			// Format both paths with a starting '/' and no trailing '/'
			if (value.charAt(0) !== '/') {
				value = '/' + value;
			}
			if (value.charAt(value.length - 1) !== '/') {
				value = value + '/';
			}
			if (endpoint.slice(0, value.length) === value) {
				return true;
			}
		}
		return false;
	}

	function setPeriod(value: Period) {
		settings.period = value;
	}

	function sortedFrequencies(freq: ValueCount): string[] {
		return Object.entries(freq)
			.sort((a, b) => {
				return b[1] - a[1];
			})
			.map((value) => value[0]);
	}

	function getHostnames() {
		const hostnameFreq: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const hostname = data[i][ColumnIndex.Hostname];
			if (hostname === null || hostname === '' || hostname === 'null') {
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

	function setHostnames() {
		hostnames = getHostnames();
	}

	async function fetchData() {
		const url = getServerURL();

		let data: DashboardData = { requests: [], user_agents: {} };
		try {
			const response = await fetch(`${url}/api/requests/${userID}/1`);
			if (response.ok && response.status === 200) {
				data = await response.json();
			} else {
				fetchStatus.failed = true;
			}
		} catch (e) {
			fetchStatus.failed = true;
			fetchStatus.reason = e;
			console.log(e.message);
		}

		return data;
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
	let settings: DashboardSettings = initSettings();
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
	// let changingPeriod: boolean = false;
	let loading: boolean = true;
	const fetchStatus: { failed: boolean; reason: string } = {
		failed: false,
		reason: ''
	};
	let endpointsRendered: boolean = false;
	const pageSize = 200_000;
	onMount(async () => {
        const dashboardData = await getDashboardData();
        data = dashboardData.requests;
        userAgents = dashboardData.user_agents;

		// loading = true;
		if (data.length === pageSize) {
			// Fetch page 2 and onwards if initial fetch didn't get all data
			fetchAdditionalPage(2);
		} else {
			loading = false;
		}

		setHostnames();
		parseDates(data);
		sortByTime(data);
		setPeriod(settings.period);

		console.log(data);
	});

	async function fetchAdditionalPage(page: number) {
		try {
			const url = getServerURL();
			const response = await fetch(`${url}/api/requests/${userID}/${page}`, {
				signal: AbortSignal.timeout(180000)
			});
			if (response.status !== 200) {
				loading = false;
				return;
			}

			const json = await response.json();
			if (json.requests.length <= 0) {
				loading = false;
				return;
			}

			userAgents = { ...userAgents, ...json.user_agents };

			parseDates(json.requests);

			json.requests?.sort((a, b) => {
				return a[ColumnIndex.CreatedAt].getTime() - b[ColumnIndex.CreatedAt].getTime();
			});

			const mostRecent = json.requests[json.requests.length - 1][ColumnIndex.CreatedAt];
			if (dateInPeriod(mostRecent, settings.period)) {
				// Trigger dashboard re-render
				data = data.concat(json.requests);
			} else {
				Object.assign(data, data.concat(json.requests));
			}

			setHostnames();

			console.log(data);

			if (json.requests.length === pageSize) {
				await fetchAdditionalPage(page + 1);
			} else {
				loading = false;
			}
		} catch (e) {
			console.log(e);
			loading = false;
		}
	}

	async function getDashboardData() {
		if ($page.params.uuid === 'demo') {
			return genDemoData();
        }

        return await fetchData();
	}

	function getUserAgent(id: number) {
		if (!(id in userAgents)) {
			return ''
		} 

		return userAgents[id];
	}

	function refreshData() {
		if (data === undefined) {
			return;
		}

		setPeriodData();
	}

	// If any settings are updated or target path/location is reset, refresh data with this new filter
	$: if (settings.targetEndpoint.path !== undefined && settings.targetLocation !== undefined) {
		refreshData();
	}
</script>

{#if periodData && data.length > 0}
	<div class="dashboard">
		<Navigation bind:settings={settings} bind:showSettings={showSettings} bind:hostnames={hostnames}/>

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
					bind:endpointsRendered
				/>
				<Version data={periodData.current} bind:endpointsRendered />
			</div>
			<div class="right">
				<Activity data={periodData.current} period={settings.period} />
				<div class="grid-row">
					<Location data={periodData.current} bind:targetLocation={settings.targetLocation} />
					<Device data={periodData.current} {getUserAgent} />
				</div>
				<UsageTime data={periodData.current} />
				<TopUsers data={periodData.current} bind:targetUser={settings.targetUser} />
			</div>
		</div>
	</div>
{:else if periodData && data.length <= 0}
	<Error reason={'no-requests'} description="" />
{:else if fetchStatus.failed}
	<Error reason={'error'} description={fetchStatus.reason} />
{:else}
	<div class="placeholder">
		<div class="spinner">
			<div class="loader"></div>
		</div>
	</div>
{/if}
<Settings
	bind:show={showSettings}
	bind:settings
	exportCSV={() => {
		exportCSV(periodData.current, columns, userAgents);
	}}
/>
<Notification state={notification} />

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
	@media screen and (max-width: 600px) {
		.right,
		.left {
			margin: 0 1em;
		}
	}
	@media screen and (max-width: 450px) {
		.dashboard-content {
			margin: 1.4em 0em 3.5em;
		}
			margin: 2.5em 1em 0;
	}
</style>
