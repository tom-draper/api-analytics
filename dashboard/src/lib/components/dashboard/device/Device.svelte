<script lang="ts">
	import Client from './Client.svelte';

	type Tab = 'client' | 'os' | 'device';

	// Track the active tab
	let activeBtn = $state<Tab>('client');
	let osLoaded = $state(false);
	let deviceLoaded = $state(false);
	let OperatingSystem = $state<any>(undefined);
	let DeviceType = $state<any>(undefined);

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

	let { data, userAgents }: { data: RequestsData; userAgents: { [id: string]: string } } = $props();
</script>

<div class="card">
	<div class="card-title">
		Device
		<div class="toggle">
			<button class:active={activeBtn === 'client'} onclick={() => setBtn('client')}>
				Client
			</button>
			<button class:active={activeBtn === 'os'} onclick={() => setBtn('os')}> OS </button>
			<button class:active={activeBtn === 'device'} onclick={() => setBtn('device')}>
				Device
			</button>
		</div>
	</div>

	<div class="client" class:display={activeBtn === 'client'}>
		<Client {data} {userAgents} />
	</div>

	{#if activeBtn === 'os' && OperatingSystem}
		<div>
			<OperatingSystem {data} {userAgents} />
		</div>
	{/if}

	{#if activeBtn === 'device' && DeviceType}
		<div>
			<DeviceType {data} {userAgents} />
		</div>
	{/if}
</div>

<style scoped>
	.card {
		margin: 2em 0 2em 1em;
		width: 430px;
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
	.client {
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
