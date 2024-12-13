<script lang="ts">
	import { onMount } from 'svelte';
	import type { DashboardSettings } from '../../lib/settings';
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

	let container: HTMLDivElement;
	onMount(() => {
		container.addEventListener('click', (e) => {
			e.stopImmediatePropagation();
		});
	});

	export let show: boolean,
		settings: DashboardSettings,
		exportCSV: () => void;
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div class="background" class:hidden={!show} on:click={hideSettings}>
	<div class="container" bind:this={container}>
		<h2 class="title">Settings</h2>
		<div class="disable404 setting">
			<div class="setting-label">Disable 404</div>
			<input
				type="checkbox"
				name="disable404"
				id="checkbox"
				on:change={toggleDisable404}
			/>
		</div>
		<div class="disable404 setting">
			<div class="setting-label">Ignore Params</div>
			<input
				type="checkbox"
				name="ignoreParams"
				id="checkbox"
				on:change={toggleIgnoreParams}
			/>
		</div>
		<div class="setting-title">Filters:</div>
		<div class="setting-filters">
			<div class="setting-filter">
				Hostname: <span class:text-white={settings.hostname}
					>{settings.hostname ?? 'None'}</span
				>
			</div>
			<div class="setting-filter">
				Period: <span class:text-white={settings.period}
					>{settings.period === 'All time'
						? 'None'
						: settings.period}</span
				>
			</div>
			<div class="setting-filter">
				Endpoint: <span class:text-white={settings.targetEndpoint.path}
					>{settings.targetEndpoint.path ?? 'None'}</span
				>
			</div>
			<div class="setting-filter">
				Status: <span class:text-white={settings.targetEndpoint.status}
					>{settings.targetEndpoint.status ?? 'None'}</span
				>
			</div>
			<div class="setting-filter">
				Location: <span class:text-white={settings.targetLocation}
					>{settings.targetLocation ?? 'None'}</span
				>
			</div>
		</div>
		<div class="setting-title">Hidden endpoints:</div>
		<div class="setting">
			<List
				bind:items={settings.hiddenEndpoints}
				placeholder={'/api/v1/example'}
			/>
		</div>
		<div class="export-csv">
			<button class="export-csv-btn" on:click={exportCSV}>
				Export CSV
			</button>
		</div>
	</div>
</div>

<style scoped>
	.background {
		background: rgba(0, 0, 0, 0.6);
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
		width: 35vw;
		/* min-height: 30vh; */
		border: 1px solid #2e2e2e;
		color: var(--faded-text);
		z-index: 20;
		position: absolute;
		cursor: default !important;
		pointer-events: bounding-box;
		padding: 30px 50px 50px;
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
	.text-white {
		color: white;
	}

	.setting-filters {
		margin-bottom: 2em;
		margin-top: 5px;
		text-align: left;
		font-size: 0.9em;
		color: #707070;
	}
	.setting-filter {
		margin: 2px 0;
	}

	input {
		margin-bottom: 2em;
	}

	#checkbox {
		height: 15px;
		width: 15px;
		cursor: pointer;
	}

	.export-csv {
		position: absolute;
		text-align: right;
		/* margin-top: 30px; */
		right: 50px;
		top: 2.4em;
	}
	.export-csv-btn {
		background: var(--background);
		color: var(--dim-text);
		border: 1px solid #2e2e2e;
		padding: 5px 12px;
		cursor: pointer;
		border-radius: 3px;
	}
	.export-csv-btn:hover {
		background: var(--highlight);
		color: var(--background);
	}

	@media screen and (max-width: 1400px) {
		.container {
			width: 60vw;
		}
	}

	@media screen and (max-width: 800px) {
		.container {
			width: 70vw;
			padding: 1.5em 2em;
		}
		.export-csv {
			top: 1.5em;
			right: 2em;
		}
	}
</style>
