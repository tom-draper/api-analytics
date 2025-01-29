<script lang="ts">
	import Client from './Client.svelte';
	import OperatingSystem from './OperatingSystem.svelte';
	import DeviceType from './DeviceType.svelte';

	type Tab = 'client' | 'os' | 'device';

	function setBtn(target: Tab) {
		activeBtn = target;
		// Resize window to trigger new plot resize to match current card size
		window.dispatchEvent(new Event('resize'));
	}

	let activeBtn: Tab = 'client';

	export let data: RequestsData, userAgents: { [id: string]: string };
</script>

<div class="card">
	<div class="card-title">
		Device
		<div class="toggle">
			<button
				class:active={activeBtn === 'client'}
				on:click={() => {
					setBtn('client');
				}}>Client</button
			>
			<button
				class:active={activeBtn === 'os'}
				on:click={() => {
					setBtn('os');
				}}>OS</button
			>
			<button
				class:active={activeBtn === 'device'}
				on:click={() => {
					setBtn('device');
				}}>Device</button
			>
		</div>
	</div>
	<div class="client" class:display={activeBtn === 'client'}>
		<Client {data} {userAgents} />
	</div>
	<div class="os" class:display={activeBtn === 'os'}>
		<OperatingSystem {data} {userAgents} />
	</div>
	<div class="device" class:display={activeBtn === 'device'}>
		<DeviceType {data} {userAgents} />
	</div>
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
