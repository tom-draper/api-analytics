<script lang="ts">
	import Client from './Client.svelte';
	import OperatingSystem from './OperatingSystem.svelte';
	import DeviceType from './DeviceType.svelte';

	function setBtn(target: 'client' | 'os' | 'device') {
		activeBtn = target;
		// Resize window to trigger new plot resize to match current card size
		window.dispatchEvent(new Event('resize'));
	}

	let activeBtn: 'client' | 'os' | 'device' = 'client';

	export let data: RequestsData, getUserAgent: (id: number) => string;
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
		<Client {data} {getUserAgent} />
	</div>
	<div class="os" class:display={activeBtn === 'os'}>
		<OperatingSystem {data} {getUserAgent} />
	</div>
	<div class="device" class:display={activeBtn === 'device'}>
		<DeviceType {data} {getUserAgent} />
	</div>
</div>

<style>
	.card {
		margin: 2em 0 2em 1em;
		padding-bottom: 1em;
		width: 420px;
	}
	.card-title {
		display: flex;
	}
	.toggle {
		margin-left: auto;
	}
	.active {
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
	button {
		border: none;
		border-radius: 4px;
		background: rgb(68, 68, 68);
		cursor: pointer;
		padding: 2px 6px;
		margin-left: 5px;
	}
	button:hover {
		background: rgb(88, 88, 88);
	}
	@media screen and (max-width: 1600px) {
		.card {
			margin: 0 0 2em;
			width: 100%;
		}
	}
</style>
