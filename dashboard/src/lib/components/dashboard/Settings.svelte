<script lang="ts">
	import type { DashboardSettings } from '$lib/settings';
	import { periodDisplay } from '$lib/period';
	import { formatDisplayUserID } from '$lib/user';
	import List from './List.svelte';

	function toggleDisable404() {
		settings.disable404 = !settings.disable404;
	}

	function toggleIgnoreParams() {
		settings.ignoreParams = !settings.ignoreParams;
	}

	function toggleIgnoreBots() {
		settings.ignoreBots = !settings.ignoreBots;
	}

	function hideSettings() {
		show = false;
	}

	function handleClick(e: MouseEvent) {
		e.stopImmediatePropagation();
	}

	let { show = $bindable(false), settings = $bindable(), exportCSV }: { show: boolean; settings: DashboardSettings; exportCSV: () => void } = $props();
	let hiddenEndpoints = $state<Set<string>>(settings.hiddenEndpoints);

	$effect(() => {
		settings.hiddenEndpoints = hiddenEndpoints;
	});
</script>

<div class="background" class:hidden={!show} role="presentation" onclick={hideSettings} onkeydown={(e) => e.key === 'Escape' && hideSettings()}>
	<div class="container" role="dialog" aria-modal="true" aria-label="Settings" tabindex="-1" onclick={handleClick} onkeydown={handleClick}>
		<h2 class="title">Settings</h2>
		<div class="setting mb-2">
			<div class="setting-label">Exclude status 404</div>
			<button
				class="toggle-switch"
				class:on={settings.disable404}
				onclick={toggleDisable404}
				title="Hide requests made to non-existent routes"
				role="switch"
				aria-checked={settings.disable404}
			>
				<span class="toggle-thumb"></span>
			</button>
		</div>
		<div class="setting mb-2">
			<div class="setting-label">Exclude bots and crawlers</div>
			<button
				class="toggle-switch"
				class:on={settings.ignoreBots}
				onclick={toggleIgnoreBots}
				title="Hide requests from bots, crawlers and automated tools"
				role="switch"
				aria-checked={settings.ignoreBots}
			>
				<span class="toggle-thumb"></span>
			</button>
		</div>
		<div class="setting mb-8">
			<div class="setting-label">Exclude URL params</div>
			<button
				class="toggle-switch"
				class:on={settings.ignoreParams}
				onclick={toggleIgnoreParams}
				title="Ignore URL parameters when grouping endpoints"
				role="switch"
				aria-checked={settings.ignoreParams}
			>
				<span class="toggle-thumb"></span>
			</button>
		</div>
		<div class="setting-title">Filters:</div>
		<div class="setting-filters mb-8 mt-1">
			<div class="setting-filter" class:active={settings.hostname}>
				Hostname: <span>{settings.hostname || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.period}>
				Period: <span>{periodDisplay[settings.period]}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetEndpoint.path}>
				Endpoint: <span>{settings.targetEndpoint.path || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetEndpoint.status}>
				Status: <span>{settings.targetEndpoint.status || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetLocation}>
				Location: <span>{settings.targetLocation || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetUser}>
				User: <span>{formatDisplayUserID(settings.targetUser) || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetVersion}>
				Version: <span>{settings.targetVersion || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetClient}>
				Client: <span>{settings.targetClient || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetDeviceType}>
				Device: <span>{settings.targetDeviceType || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetOS}>
				OS: <span>{settings.targetOS || 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetWeekday !== null}>
				Day: <span>{settings.targetWeekday !== null ? ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][settings.targetWeekday] : 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetHour !== null}>
				Hour: <span>{settings.targetHour !== null ? `${settings.targetHour}:00` : 'None'}</span>
			</div>
			<div class="setting-filter" class:active={settings.targetReferrer}>
				Referrer: <span>{settings.targetReferrer || 'None'}</span>
			</div>
		</div>
		<div class="setting-title">Hidden endpoints:</div>
		<div class="setting mb-4">
			<List bind:items={hiddenEndpoints} placeholder={'/api/v1/example'} />
		</div>
		<div class="export-csv">
			<button class="export-csv-btn" onclick={exportCSV} title="Export data to CSV"
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
		align-items: center;
	}
	.setting-label {
		margin-right: 10px;
	}

	/* Toggle switch */
	.toggle-switch {
		position: relative;
		width: 36px;
		height: 20px;
		border-radius: 10px;
		background: #3a3a3a;
		border: none;
		cursor: pointer;
		padding: 0;
		flex-shrink: 0;
		transition: background 0.2s ease;
	}
	.toggle-switch.on {
		background: var(--highlight);
	}
	.toggle-thumb {
		position: absolute;
		top: 3px;
		left: 3px;
		width: 14px;
		height: 14px;
		border-radius: 50%;
		background: #fff;
		transition: transform 0.2s ease;
		pointer-events: none;
	}
	.toggle-switch.on .toggle-thumb {
		transform: translateX(16px);
	}

	.setting-filters {
		text-align: left;
		font-size: 0.9em;
		color: var(--dim-text);
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

	svg {
		width: 22px;
		height: 22px;
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
