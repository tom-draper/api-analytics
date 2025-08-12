<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import Card from '$components/monitor/Card.svelte';
	import TrackNew from '$components/monitor/TrackNew.svelte';
	import Notification from '$components/dashboard/Notification.svelte';
	import formatUUID from '$lib/uuid';
	import type { NotificationState } from '$lib/notification';
	import { getServerURL } from '$lib/url';
	import { page } from '$app/stores';
	import Lightning from '$components/Lightning.svelte';
	import type { MonitorPeriod } from '$lib/period';

	const userID = formatUUID($page.params.uuid);

	async function fetchData() {
		const url = getServerURL();

		let data: MonitorData = {};
		try {
			const response = await fetch(`${url}/api/monitor/pings/${userID}`);
			if (response.status === 200) {
				data = await response.json();
			}
		} catch (e) {
			console.log(e);
		}

		console.log(data);

		return data;
	}

	function setPeriod(value: MonitorPeriod) {
		period = value;
		error = false;
	}

	function toggleShowTrackNew() {
		showTrackNew = !showTrackNew;
	}

	function addEmptyMonitor(url: string) {
		data[url] = [];
	}

	function removeMonitor(url: string) {
		data = Object.fromEntries(Object.entries(data).filter(([key]) => key !== url));
		// Reset error flag in case deleted monitor was the cause
		error = false;
	}

	function allPending(data: MonitorData) {
		for (const samples of Object.values(data)) {
			if (samples.length > 0) {
				return false;
			}
		}

		return true;
	}

	function getStatus(data: MonitorData, error: boolean) {
		const monitorCount = Object.keys(data).length;
		if (monitorCount === 0) {
			return 'setup';
		} else if (allPending(data)) {
			return 'pending';
		} else if (error) {
			return 'offline';
		} else {
			return 'online';
		}
	}

	function getStatusTitle(status: Status) {
		switch (status) {
			case 'online':
				return 'Systems Online';
			case 'offline':
				return 'Systems Down';
			case 'pending':
				return 'Status Pending';
			case 'setup':
				return 'Setup Required';
			default:
				return '';
		}
	}

	let error = false;
	const periods: MonitorPeriod[] = ['24h', '7d', '30d', '60d'];
	let period = periods[1];
	let data: MonitorData;
	let showTrackNew = false;
	let notification: NotificationState = {
		message: '',
		style: 'success',
		show: false
	};

	type Status = 'setup' | 'pending' | 'online' | 'offline';

	let status: Status;
	let statusTitle: string;

	let intervalID: NodeJS.Timeout;

	$: if (data) {
		status = getStatus(data, error);
		statusTitle = getStatusTitle(status);
	}

	onMount(async () => {
		data = await fetchData();

		console.log(data);
		data['https://persona-api.vercel.app/v1'] = data['https://www.google.com']
		delete data['https://www.google.com']

		for (let i = 2053; i < 2058; i++) {
			data['https://persona-api.vercel.app/v1'][i].status = 500
			data['https://persona-api.vercel.app/v1'][i].response_time = 0
		}

		const refreshData = async () => {
			data = await fetchData();
		};

		if (intervalID) {
			clearInterval(intervalID);
		}
		intervalID = setInterval(refreshData, 1800000); // Refresh data every 30 minutes
	});

	onDestroy(() => {
		// Cleanup interval on component destroy
		clearInterval(intervalID);
	});
</script>

<div class="monitoring">
	<div class="status min-h-[160px]">
		{#if data}
			<div class="status-image">
				<div
					class="lightning"
					class:text-[var(--highlight)]={status === 'online'}
					class:text-[var(--red)]={status === 'offline'}
					class:text-[#424242]={status === 'setup' || status === 'pending'}
				>
					<Lightning />
				</div>
				<div
					class="status-text"
					class:text-[#bee7c5]={status === 'online'}
					class:text-[#ffc1c1]={status === 'offline'}
					class:text-[#c0c0c0]={status === 'setup' || status === 'pending'}
				>
					{statusTitle}
				</div>
			</div>
		{/if}
	</div>
	<div class="cards-container">
		<div class="controls">
			<div class="add-new text-sm">
				<button
					class="add-new-btn"
					class:active={showTrackNew || (data && Object.keys(data).length === 0)}
					on:click={toggleShowTrackNew}
					aria-label="Add new monitor"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-6"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
					</svg>
					<span>New</span>
				</button>
			</div>
			<div class="period-controls-container text-sm">
				<div class="period-controls">
					{#each periods as p}
						<button
							class="period-btn"
							class:active={period === p}
							on:click={() => {
								setPeriod(p);
							}}
						>
							{p}
						</button>
					{/each}
				</div>
			</div>
		</div>
		{#if data}
			{#if showTrackNew || Object.keys(data).length == 0}
				<TrackNew
					{userID}
					bind:showTrackNew
					monitorCount={Object.keys(data).length}
					bind:notification
					{addEmptyMonitor}
				/>
			{/if}
			{#each Object.keys(data).sort() as url}
				<Card
					bind:url
					{data}
					{userID}
					{period}
					bind:anyError={error}
					bind:notification
					{removeMonitor}
				/>
			{/each}
		{:else}
			<div class="spinner">
				<div class="loader"></div>
			</div>
		{/if}
	</div>
</div>
<Notification bind:state={notification} />

<style scoped>
	.monitoring {
		font-weight: 600;
	}
	.status {
		margin: 13vh 0 9vh;
		display: grid;
		place-items: center;
	}
	.status-image {
		place-items: center;
	}
	.lightning {
		height: 5em;
		margin-bottom: 2em;
		filter: saturate(1.3);
		transition: color 1s ease-out;
	}
	.status-text {
		font-size: 2em;
		font-weight: 700;
		transition: color 1s ease-out;
	}

	.cards-container {
		width: 60%;
		margin: auto;
		padding-bottom: 1em;
	}

	.controls {
		margin: auto;
		width: 60%;

		width: min(100%, 1000px);
		display: flex;
	}
	.add-new {
		flex-grow: 1;
		display: flex;
		justify-content: left;
	}
	.add-new-btn > svg {
		width: 20px;
		height: 20px;
	}
	.period-controls {
		margin-left: auto;
		display: flex;
		justify-content: right;
	}

	.add-new-btn:hover,
	.period-btn:hover {
		background: #161616;
	}

	.add-new-btn:hover {
		color: var(--highlight);
	}

	.add-new-btn:hover svg {
		filter: contrast(1.5);
	}

	.period-controls {
		border: 1px solid #2e2e2e;
		width: fit-content;
		border-radius: 4px;
		overflow: hidden;
	}
	.period-controls-container {
		margin-top: auto;
	}

	button {
		background: var(--background);
		color: var(--dim-text);
		border: none;
		padding: 3px 12px;
		cursor: pointer;
	}
	.add-new-btn {
		display: flex;
	}

	.add-new-btn {
		border: 1px solid #2e2e2e;
		border-radius: 4px;
		padding: 0;
		display: flex;
		place-items: center;
	}
	.add-new-btn > span {
		padding: 0 12px 0 0.2em;
		color: var(--dim-text) !important;
	}

	.active > span {
		color: var(--dark-background) !important;
	}

	.add-new-btn > svg {
		margin: 0 0.5em;
		width: 16px;
		transition: color 0.4s ease-in-out;
	}
	.active > svg {
		transition: none;
	}
	.active,
	.active:hover {
		background: var(--highlight) !important;
		color: var(--dark-background) !important;
	}
	.spinner {
		margin: 3em 0 10em;
	}
	.loader {
		width: 40px;
		height: 40px;
	}

	@media screen and (max-width: 1100px) {
		.cards-container {
			width: 95%;
		}
	}

	@media screen and (max-width: 600px) {
		.status {
			margin: 10vh 0 9vh !important;
			font-size: 0.9em;
		}
	}
</style>
