<script lang="ts">
	import { onMount } from 'svelte';
	import Card from '$lib/components/monitor/Card.svelte';
	import TrackNew from '$lib/components/monitor/TrackNew.svelte';
	import Notification from '$lib/components/dashboard/Notification.svelte';
	import formatUUID from '$lib/uuid';
	import type { NotificationState } from '$lib/notification';
	import { getServerURL } from '$lib/url';
	import { page } from '$app/stores';

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

	function setPeriod(value: string) {
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
		delete data[url];
		data = data; // Trigger reactivity to update display
	}

	let error = false;
	const periods = ['24h', '7d', '30d', '60d'];
	let period = periods[1];
	let data: MonitorData;
	let notification: NotificationState = {
		message: '',
		style: 'error',
		show: false
	};

	let showTrackNew = false;
	onMount(async () => {
		data = await fetchData();
	});
</script>

<div class="monitoring">
	<div class="status">
		{#if data !== undefined && Object.keys(data).length === 0}
			<div class="status-image">
				<img id="status-image" src="/images/logos/lightning-green.svg" alt="" />
				<div class="status-text">Setup Required</div>
			</div>
		{:else if error}
			<div class="status-image">
				<img id="status-image" src="/images/logos/lightning-red.svg" alt="" />
				<div class="status-text">Systems Down</div>
			</div>
		{:else}
			<div class="status-image">
				<img id="status-image" src="/images/logos/lightning-green.svg" alt="" />
				<div class="status-text">Systems Online</div>
			</div>
		{/if}
	</div>
	<div class="cards-container">
		<div class="controls">
			<div class="add-new text-sm">
				<button class="add-new-btn" class:active={showTrackNew} on:click={toggleShowTrackNew}>
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
				</button>
			</div>
			<div class="period-controls-container text-sm">
				<div class="period-controls">
					{#each periods as _period}
						<button
							class="period-btn"
							class:active={period === _period}
							on:click={() => {
								setPeriod(_period);
							}}
						>
							{_period}
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
	#status-image {
		height: 5em;
		margin-bottom: 2em;
		filter: saturate(1.3);
	}
	.status-text {
		font-size: 2em;
		font-weight: 700;
		color: white;
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

	.period-btn:hover {
		background: #161616

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
		background: var(--light-background);
	}
	.add-new-btn:hover {
		background: radial-gradient(var(--light-background), #3fcf8e10);
	}
	.add-new-btn {
		border: 1px solid #2e2e2e;
		border-radius: 4px;
		padding: 0;
		height: 35px;
		color: var(--highlight);
		width: 35px;
		display: grid;
		place-items: center;
	}
	.active,
	.active:hover {
		background: var(--highlight) !important;
		color: black !important;
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
</style>
