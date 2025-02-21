<script lang="ts">
	import Client from './Client.svelte';

	type Tab = 'client' | 'os' | 'device';

	// Track the active tab
	let activeBtn: Tab = 'client';

	// Track whether the other components should be loaded
	let osLoaded = false;
	let deviceLoaded = false;

	// Variables to hold the dynamically imported components
	let OperatingSystem: any;
	let DeviceType: any;

	// Function to load components dynamically
	async function loadComponent(tab: Tab) {
		if (tab === 'os' && !osLoaded) {
			const { default: importedOperatingSystem } = await import('./OperatingSystem.svelte');
			OperatingSystem = importedOperatingSystem;
			osLoaded = true;
		} else if (tab === 'device' && !deviceLoaded) {
			const { default: importedDeviceType } = await import('./DeviceType.svelte');
			DeviceType = importedDeviceType;
			deviceLoaded = true;
		}
	}

	// Function to set the active button
	function setBtn(target: Tab) {
		activeBtn = target;
		// Load the component when the tab is clicked
		loadComponent(target);
		// Resize window to trigger new plot resize to match current card size
		// window.dispatchEvent(new Event('resize'));
	}

	export let data: RequestsData;
	export let userAgents: { [id: string]: string };
</script>

<div class="card">
	<div class="card-title">
		Device
		<div class="toggle">
			<button
				class:active={activeBtn === 'client'}
				on:click={() => setBtn('client')}>
				Client
			</button>
			<button
				class:active={activeBtn === 'os'}
				on:click={() => setBtn('os')}>
				OS
			</button>
			<button
				class:active={activeBtn === 'device'}
				on:click={() => setBtn('device')}>
				Device
			</button>
		</div>
	</div>

	<div class="client" class:display={activeBtn === 'client'}>
		<Client {data} {userAgents} />
	</div>

	{#if activeBtn === 'os' && OperatingSystem}
		<div>
			<svelte:component this={OperatingSystem} {data} {userAgents} />
		</div>
	{/if}

	{#if activeBtn === 'device' && DeviceType}
		<div>
			<svelte:component this={DeviceType} {data} {userAgents} />
		</div>
	{/if}
</div>

<style scoped>
	.card {
		margin: 2em 0 2em 1em;
		width: 420px;
	}
	.card-title {
		display: flex;
	}
	.toggle {
		margin-left: auto;
	}
	.toggle > .active {
		background: var(--highlight);
	}
	.os,
	.client,
	.device {
		display: none;
	}
	.display {
		display: initial;
	}
	.toggle > button {
		font-size: 0.85em;
		color: #000;
		border: none;
		border-radius: 4px;
		background: rgb(68, 68, 68);
		cursor: pointer;
		padding: 0 6px;
		margin-left: 5px;
	}
	.toggle > button:hover {
		background: rgb(88, 88, 88);
	}
	.toggle > .active:hover {
		background: var(--highlight);
	}
	@media screen and (max-width: 1600px) {
		.card {
			margin: 0 0 2em;
			width: 100%;
		}
	}
</style>
