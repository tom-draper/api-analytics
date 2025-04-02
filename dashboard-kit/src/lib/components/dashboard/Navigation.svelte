<script lang="ts">
	import { page } from '$app/stores';
	import { page as p } from '$app/state';
	import { replaceState } from '$app/navigation';
	import type { DashboardSettings } from '$lib/settings';
	import formatUUID from '$lib/uuid';
	import Dropdown from './Dropdown.svelte';
	import type { Period } from '$lib/period';

	const timePeriods: Period[] = ['24 hours', 'Week', 'Month', '6 months', 'Year', 'All time'];

	const userID = formatUUID($page.params.uuid);

	function setPeriodParam(period: Period) {
		p.url.searchParams.set('period', period.toLocaleLowerCase().replace(' ', ''))
		replaceState(p.url, p.state)
	}

	function setHostnameParam(hostname: string | null) {
		if (hostname === null) {
			p.url.searchParams.delete('hostname')
		} else {
			p.url.searchParams.set('hostname', hostname);
		}
		replaceState(p.url, p.state)
	}

	$: hostname = settings.hostname;

	$: {
		setHostnameParam(hostname ?? null)
	}

	let dropdownOpen: boolean = false;

	export let settings: DashboardSettings, showSettings: boolean, hostnames: string[];
</script>

<nav class="button-nav text-sm">
	<!-- <a class="info" href={userID ? `/explorer/${userID}` : '/explorer'}>
		<div class="info-content">
			Try the Log Explorer <svg
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
					d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"
				/>
			</svg>
		</div>
	</a> -->
	<div class="donate">
		<a target="_blank" href="https://www.buymeacoffee.com/tomdraper" class="donate-link">Donate</a>
	</div>
	<button
		class="settings"
		on:click={() => {
			showSettings = true;
		}}
	>
		<img class="settings-icon" src="/images/icons/cog.png" alt="" />
	</button>
	<div class="dropdown-container" class:no-display={hostnames.length <= 1}>
		<Dropdown
			options={hostnames.slice(0, 25)}
			bind:selected={settings.hostname}
			bind:open={dropdownOpen}
			defaultOption={'All hostnames'}
		/>
	</div>
	<div class="nav-btn time-period">
		{#each timePeriods as period}
			<button
				class="time-period-btn"
				class:time-period-btn-active={settings.period === period}
				on:click={() => {
					settings.period = period;
					dropdownOpen = false;
					setPeriodParam(period);
				}}
			>
				{period}
			</button>
		{/each}
	</div>
</nav>

<style scoped>
	.button-nav {
		margin: 2.5em 2rem 0;
		display: flex;
	}
	.info {
		background: var(--background);
		padding: 2px 12px;
		color: var(--dim-text);
		cursor: pointer;
		border: 1px solid #2e2e2e;
		border-radius: 4px;
		font-weight: 400;
	}
	.info-content {
		display: flex;
		align-items: center;
		cursor: pointer;
	}
	.info-content > svg {
		margin-left: 0.6em;
		width: 16px;
		transition: transform 0.15s ease;
	}
	.info:hover .info-content > svg {
		transform: translateX(2px);
	}
	.info:hover {
		background: #161616;
	}
	.time-period {
		display: flex;
		border: 1px solid #2e2e2e;
		border-radius: 4px;
		overflow: hidden;
	}
	.time-period-btn {
		background: var(--background);
		padding: 4px 12px;
		border: none;
		color: var(--dim-text);
		cursor: pointer;
	}
	.time-period-btn:hover {
		background: #161616;
	}
	.time-period-btn-active:hover {
		background: var(--highlight);
		color: black;
	}
	.time-period-btn-active {
		background: var(--highlight);
		color: black;
	}

	@keyframes gradient-shift {
		0%,
		100% {
			background-position: 0% 50%; /* Start at the left */
		}
		50% {
			background-position: 100% 50%; /* End at the right */
		}
	}

	.settings {
		background: transparent;
		outline: none;
		border: none;
		margin-right: 10px;
		cursor: pointer;
		text-align: right;
	}
	.donate {
		margin-left: auto;
		display: grid;
		place-items: center;
		margin-right: 1em;
	}

	.dropdown-container {
		margin-right: 10px;
	}

	.donate-link {
		color: rgb(73, 73, 73);
		color: rgb(82, 82, 82);
		color: #464646;
		transition: 0.1s;
	}
	.donate-link:hover {
		color: var(--highlight);
	}
	.settings-icon {
		width: 20px;
		height: 20px;
		filter: contrast(0.45);
		margin-top: 2px;
		transition: 0.1s;
	}
	.settings-icon:hover {
		filter: contrast(0.01);
	}

	@media screen and (max-width: 800px) {
	}

	@media screen and (max-width: 1300px) {
		.button-nav {
			margin: 2.5em 3rem 0;
		}
		.info {
			display: none;
		}
	}

	@media screen and (max-width: 820px) {
		.button-nav {
			flex-direction: column;
		}
		.dropdown-container {
			margin-left: auto;
			margin-right: 0;
			margin: -2em 0 0 auto;
		}
		.time-period {
			margin-top: 15px;
		}
		.time-period-btn {
			flex: 1;
		}
		.settings {
			width: 30px;
			height: 30px;
			margin-left: 0.5em;
			margin-top: 0;
			margin-right: auto;
		}
		.donate {
			display: none;
		}
	}
	@media screen and (max-width: 660px) {
		.time-period {
			right: 1em;
		}
		.button-nav {
			margin: 2em 1rem 0;
		}
		.time-period-btn {
			padding: 3px 0;
		}
	}
	@media screen and (max-width: 500px) {
		.time-period-btn {
			flex-grow: 1;
			flex: auto;
		}
	}
	@media screen and (max-width: 450px) {
		.button-nav {
			margin: 2.5em 1em 0;
		}
	}
</style>
