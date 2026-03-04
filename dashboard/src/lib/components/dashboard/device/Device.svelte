<script lang="ts">
	import Client from './Client.svelte';
	import OperatingSystem from './OperatingSystem.svelte';
	import DeviceType from './DeviceType.svelte';

	type Tab = 'client' | 'os' | 'device';

	let activeBtn = $state<Tab>('client');

	function setBtn(target: Tab) {
		activeBtn = target;
		window.dispatchEvent(new Event('resize'));
	}

	let { uaIdCount, userAgents }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
	} = $props();
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

	<div class="tab" class:tab-active={activeBtn === 'client'}>
		<Client {uaIdCount} {userAgents} />
	</div>
	<div class="tab" class:tab-active={activeBtn === 'os'}>
		<OperatingSystem {uaIdCount} {userAgents} />
	</div>
	<div class="tab" class:tab-active={activeBtn === 'device'}>
		<DeviceType {uaIdCount} {userAgents} />
	</div>
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
	.tab {
		display: none;
	}
	.tab-active {
		display: block;
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
