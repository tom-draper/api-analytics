<script lang="ts">
	import { onMount } from 'svelte';
	import Requests from '../components/dashboard/Requests.svelte';
	import Logo from '../components/dashboard/Logo.svelte';
	import ResponseTimes from '../components/dashboard/ResponseTimes.svelte';
	import Users from '../components/dashboard/Users.svelte';
	import Endpoints from '../components/dashboard/Endpoints.svelte';
	import Footer from '../components/Footer.svelte';
	import SuccessRate from '../components/dashboard/SuccessRate.svelte';
	import Activity from '../components/dashboard/activity/Activity.svelte';
	import Version from '../components/dashboard/Version.svelte';
	import UsageTime from '../components/dashboard/UsageTime.svelte';
	import Location from '../components/dashboard/Location.svelte';
	import Device from '../components/dashboard/device/Device.svelte';
	import { dateInPeriod, dateInPrevPeriod } from '../lib/period';
	import genDemoData from '../lib/demo';
	import formatUUID from '../lib/uuid';
	import Settings from '../components/dashboard/Settings.svelte';
	import type { DashboardSettings, Period } from '../lib/settings';
	import { initSettings } from '../lib/settings';
	import type { NotificationState } from '../lib/notification';
	import Dropdown from '../components/dashboard/Dropdown.svelte';
	import Notification from '../components/dashboard/Notification.svelte';
	import exportCSV from '../lib/exportData';
	import { ColumnIndex, columns } from '../lib/consts';
	import Error from '../components/dashboard/Error.svelte';
	import TopUsers from '../components/dashboard/TopUsers.svelte';
	import { getServerURL } from '../lib/url';

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
				(settings.targetEndpoint.path === null ||
					settings.targetEndpoint.path === path) &&
				(settings.targetEndpoint.status === null ||
					settings.targetEndpoint.status === status) &&
				(settings.targetLocation === null ||
					settings.targetLocation === location) &&
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
		console.log('setting data');
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
			(firstChar === '/' &&
				settings.hiddenEndpoints.has(endpoint.slice(1))) ||
			(lastChar === '/' &&
				settings.hiddenEndpoints.has(endpoint.slice(0, -1))) ||
			(firstChar !== '/' &&
				settings.hiddenEndpoints.has('/' + endpoint)) ||
			(lastChar !== '/' &&
				settings.hiddenEndpoints.has(endpoint + '/')) ||
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

		userID = formatUUID(userID);
		try {
			const response = await fetch(`${url}/api/requests/${userID}/1`);
			if (response.ok && response.status === 200) {
				const data: DashboardData = await response.json();
				return data;
			} else {
				fetchStatus.failed = true;
			}
		} catch (e) {
			fetchStatus.failed = true;
			fetchStatus.reason = e;
			console.log(e.message);
		}
	}

	function parseDates(data: RequestsData) {
		for (let i = 0; i < data.length; i++) {
			data[i][ColumnIndex.CreatedAt] = new Date(
				data[i][ColumnIndex.CreatedAt],
			);
		}
	}

	function sortByTime(data: RequestsData) {
		data.sort((a, b) => {
			return (
				a[ColumnIndex.CreatedAt].getTime() -
				b[ColumnIndex.CreatedAt].getTime()
			);
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
		show: false,
	};
	let periodData: {
		current: RequestsData;
		previous: RequestsData;
	};
	// let changingPeriod: boolean = false;
	const timePeriods: Period[] = [
		'24 hours',
		'Week',
		'Month',
		'6 months',
		'Year',
		'All time',
	];
	let loading: boolean = true;
	const fetchStatus: { failed: boolean; reason: string } = {
		failed: false,
		reason: '',
	};
	let endpointsRendered: boolean = false;
	const pageSize = 200_000;
	onMount(async () => {
		({ requests: data, user_agents: userAgents } =
			await getDashboardData());

		// loading = true;
		if (data.length === pageSize) {
			// Fetch page 2 and onwards if initial fetch didn't get all data
			fetchAdditionalPage(2);
		} else {
			loading = false;
		}

		setPeriod(settings.period);
		setHostnames();
		parseDates(data);
		sortByTime(data);

		setPeriodData();

		console.log(data);
	});

	async function fetchAdditionalPage(page: number) {
		try {
			const url = getServerURL();
			const response = await fetch(
				`${url}/api/requests/${userID}/${page}`,
				{ signal: AbortSignal.timeout(180000) },
			);
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
				return (
					a[ColumnIndex.CreatedAt].getTime() -
					b[ColumnIndex.CreatedAt].getTime()
				);
			});

			const mostRecent =
				json.requests[json.requests.length - 1][ColumnIndex.CreatedAt];
			if (dateInPeriod(mostRecent, settings.period)) {
				// Trigger dashboard re-render
				data = data.concat(json.requests);
			} else {
				Object.assign(data, data.concat(json.requests));
			}

			// setPeriod(settings.period);
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
		let data: DashboardData;
		if (demo) {
			data = genDemoData();
		} else {
			data = await fetchData();
		}
		return data;
	}

	function getUserAgent(id: number) {
		let userAgent: string;
		if (id in userAgents) {
			userAgent = userAgents[id];
		} else {
			userAgent = '';
		}
		return userAgent;
	}

	function refreshData() {
		if (data === undefined) {
			return;
		}

		setPeriodData();
	}

	// If any settings are updated or target path/location is reset, refresh data with this new filter
	$: if (
		settings.targetEndpoint.path !== undefined &&
		settings.targetLocation !== undefined
	) {
		refreshData();
	}

	export let location: string, userID: string, demo: boolean;
</script>

{#if periodData && data.length > 0}
	<div class="dashboard">
		<div class="button-nav">
			<div class="donate">
				<a
					target="_blank"
					href="https://www.buymeacoffee.com/tomdraper"
					class="donate-link">Donate</a
				>
			</div>
			<button
				class="settings"
				on:click={() => {
					showSettings = true;
				}}
			>
				<img class="settings-icon" src="../img/icons/cog.png" alt="" />
			</button>
			{#if hostnames.length > 1}
				<div class="dropdown-container">
					<Dropdown
						options={hostnames.slice(0, 25)}
						bind:selected={settings.hostname}
						defaultOption={'All hostnames'}
					/>
				</div>
			{/if}
			<div class="nav-btn time-period">
				{#each timePeriods as period}
					<button
						class="time-period-btn"
						class:time-period-btn-active={settings.period ===
							period}
						on:click={() => {
							setPeriod(period);
						}}
					>
						{period}
					</button>
				{/each}
			</div>
		</div>
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
					<Location
						data={periodData.current}
						bind:targetLocation={settings.targetLocation}
					/>
					<Device data={periodData.current} {getUserAgent} />
				</div>
				<UsageTime data={periodData.current} />
				<TopUsers data={periodData.current} bind:targetUser={settings.targetUser}/>
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
			<div class="loader" />
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
<Footer />

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
	.button-nav {
		margin: 2.5em 2em 0;
		display: flex;
	}
	.time-period {
		display: flex;
		border: 1px solid #2e2e2e;
		border-radius: 4px;
		overflow: hidden;
		height: 27px;
	}
	.time-period-btn {
		background: var(--background);
		padding: 3px 12px;
		border: none;
		color: var(--dim-text);
		cursor: pointer;
	}
	.time-period-btn:hover {
		background: #161616;
	}
	.time-period-btn-active:hover {
		background: var(--highlight);
		color: black;
	}
	.time-period-btn-active {
		background: var(--highlight);
		color: black;
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

	.settings {
		background: transparent;
		outline: none;
		border: none;
		margin-right: 10px;
		cursor: pointer;
		text-align: right;
	}
	.donate {
		margin-left: auto;
		font-size: 0.85em;
		display: grid;
		place-items: center;
		margin-right: 1em;
	}

	.dropdown-container {
		margin-right: 10px;
	}

	.donate-link {
		color: rgb(73, 73, 73);
		color: rgb(82, 82, 82);
		color: #464646;
		transition: 0.1s;
	}
	.donate-link:hover {
		color: var(--highlight);
	}
	.settings-icon {
		width: 20px;
		height: 20px;
		filter: contrast(0.45);
		margin-top: 2px;
		transition: 0.1s;
	}
	.settings-icon:hover {
		filter: contrast(0.01);
	}

	@media screen and (max-width: 800px) {
		.donate {
			display: none;
		}
		.settings {
			margin-left: auto;
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
		.button-nav {
			margin: 2.5em 3em 0;
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
	@media screen and (max-width: 800px) {
		.button-nav {
			flex-direction: column;
		}
		.dropdown-container {
			margin-left: auto;
			margin-right: 0;
			margin: -30px 0 0 auto;
		}
		.time-period {
			margin-top: 15px;
		}
		.time-period-btn {
			flex: 1;
		}
		.settings {
			margin-left: 0;
			margin-right: auto;
		}
	}
	@media screen and (max-width: 600px) {
		.right,
		.left {
			margin: 0 1em;
		}
		.time-period {
			right: 1em;
		}
		.button-nav {
			margin: 2.5em 2em 0;
		}
		.time-period-btn {
			padding: 3px 0;
		}
	}
	@media screen and (max-width: 500px) {
		.time-period-btn {
			flex-grow: 1;
			flex: auto;
		}
	}
	@media screen and (max-width: 450px) {
		.dashboard-content {
			margin: 1.4em 0em 3.5em;
		}
		.button-nav {
			margin: 2.5em 1em 0;
		}
	}
</style>
