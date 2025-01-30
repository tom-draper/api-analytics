<script lang="ts">
	import { onMount } from 'svelte';
	import type { DashboardSettings } from '$lib/settings';
	import List from './List.svelte';

	function toggleDisable404() {
		settings.disable404 = !settings.disable404;
	}

	function toggleIgnoreParams() {
		settings.ignoreParams = !settings.ignoreParams;
	}

	function hideSettings() {
		show = false;
	}

	function formatUserID(id: string) {
		if (!id) {
			return ''
		}

		const [ipAddress, userID] = id.split('||');
		if (ipAddress && userID) {
			return `${ipAddress} + ${userID}`
		} else  if (userID) {
			return userID;
		} else {
			return ipAddress;
		}
	}

	let container: HTMLDivElement;
	onMount(() => {
		container.addEventListener('click', (e) => {
			e.stopImmediatePropagation();
		});
	});

	export let show: boolean, settings: DashboardSettings, exportCSV: () => void;
</script>

<div class="background" class:hidden={!show} on:click={hideSettings}>
	<div class="container" bind:this={container}>
		<h2 class="title">Settings</h2>
		<div class="disable404 setting mb-2">
			<div class="setting-label">Disable 404</div>
			<input type="checkbox" name="disable404" id="checkbox" on:change={toggleDisable404} title="Hide requests that returned a 404 status code" />
		</div>
		<div class="disable404 setting mb-8">
			<div class="setting-label">Ignore Params</div>
			<input type="checkbox" name="ignoreParams" id="checkbox" on:change={toggleIgnoreParams} title="Ignore URL parameters when grouping endpoints" />
		</div>
		<div class="setting-title">Filters:</div>
		<div class="setting-filters mb-8 mt-1">
			<div class="setting-filter" class:active={settings.hostname}>
				Hostname: <span>{settings.hostname ?? 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.period}>
				Period: <span>{settings.period === 'All time' ? 'None' : settings.period}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetEndpoint.path}>
				Endpoint: <span>{settings.targetEndpoint.path ?? 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetEndpoint.status}>
				Status: <span>{settings.targetEndpoint.status ?? 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetLocation}>
				Location: <span>{settings.targetLocation ?? 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetUser}>
				User: <span>{formatUserID(settings.targetUser) ?? 'None'}</span>
			</div>
		</div>
		<div class="setting-title">Hidden endpoints:</div>
		<div class="setting mb-4">
			<List bind:items={settings.hiddenEndpoints} placeholder={'/api/v1/example'} />
		</div>
		<div class="export-csv">
			<button class="export-csv-btn" on:click={exportCSV} title="Export data to CSV"
				><svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="size-6"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
					/>
				</svg>
			</button>
		</div>
	</div>
</div>

<style scoped>
	h2 {
		color: var(--highlight);
		margin-bottom: 1em;
		font-weight: 700;
	}
	.background {
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(4px);
		height: 100vh;
		width: 100%;
		display: grid;
		place-items: center;
		position: fixed;
		top: 0;
		cursor: pointer;
		z-index: 10;
	}
	.container {
		background: var(--background);
		border-radius: 6px;
		width: 42em;
		border: 1px solid #2e2e2e;
		color: var(--faded-text);
		z-index: 20;
		position: absolute;
		cursor: default !important;
		pointer-events: bounding-box;
		padding: 1.5em 2.5em;
		position: relative;
	}
	.title {
		font-size: 1.8em;
		font-weight: 600;
		text-align: left;
	}
	.setting-title {
		text-align: left;
	}
	.hidden {
		display: none;
	}
	.setting {
		display: flex;
	}
	.setting-label {
		margin-right: 10px;
	}

	.setting-filters {
		text-align: left;
		font-size: 0.9em;
		color: #707070;
		display: flex;
		flex-wrap: wrap;
		gap: 5px 10px;
	}
	.setting-filter {
		margin: 2px 0;
		background: rgba(0, 0, 0, 0.7);
		background: var(--dark-background);
		border-radius: 4px;
		padding: 5px 10px;
	}
	.active {
		background: var(--highlight);
		color: var(--background);
	}

	input {
		margin: 0;
	}
	input[type='checkbox'] {
		align-self: center;
	}

	svg {
		width: 22px;
		height: 22px;
	}

	#checkbox {
		height: 12px;
		width: 12px;
		margin-top: 2px;
		cursor: pointer;
	}

	.export-csv {
		position: absolute;
		text-align: right;
		right: 2.5em;
		top: 1.8em;
	}
	.export-csv-btn {
		background: var(--dark-background);
		color: var(--dim-text);
		border: 1px solid #2e2e2e;
		padding: 0.8em;
		cursor: pointer;
		border-radius: 4px;
		font-size: 0.85em;
	}
	.export-csv-btn:hover {
		background: var(--highlight);
		color: var(--background);
	}

	@media screen and (max-width: 800px) {
		.container {
			width: 95%;
			padding: 1.5em 2em;
		}
		.export-csv {
			top: 2em;
			right: 1.5em;
		}
	}
</style>
