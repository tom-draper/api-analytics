<script lang="ts">
	import { page as p } from '$app/state';
	import { replaceState } from '$app/navigation';
	import type { DashboardSettings } from '$lib/settings';
	import Dropdown from './Dropdown.svelte';
	import type { Period } from '$lib/period';
	import { onMount } from 'svelte';

	const timePeriods: Period[] = ['24 hours', 'week', 'month', '6 months', 'year', 'all time'];

	const donateMessages = [
		'Donate',
		'Support',
		'Contribute',
		'Contribute today',
		'Support the project',
		'Keep API Analytics running',
		'Keep the project going',
		'Keep the project alive',
		'Donate to API Analytics',
		'Support API Analytics',
		'Give back',
		'Support us',
		'Buy us a coffee',
		'Give us a tip',
		'Make a donation',
		'Support development',
		'Help with server costs'
	];

	const timePeriodsDisplay: Record<Period, string> = {
		'24 hours': '24 hours',
		week: 'Week',
		month: 'Month',
		'6 months': '6 months',
		year: 'Year',
		'all time': 'All time'
	}

	function setPeriodParam(period: Period) {
		if (period === "week") {
			// Week is default period, so avoid param to keep simple
			p.url.searchParams.delete('period');
		} else {
			p.url.searchParams.set('period', period.toLocaleLowerCase().replace(' ', '-'));
			replaceState(p.url, p.state);
		}
	}

	function setHostnameParam(hostname: string | null) {
		if (hostname === null) {
			p.url.searchParams.delete('hostname');
		} else {
			p.url.searchParams.set('hostname', hostname);
		}
		replaceState(p.url, p.state);
	}

	let donateMessage: string;
	onMount(() => {
		donateMessage = donateMessages[Math.floor(Math.random() * donateMessages.length)];
	});

	$: hostname = settings.hostname;

	$: {
		setHostnameParam(hostname ?? null);
	}

	let dropdownOpen: boolean = false;

	export let settings: DashboardSettings, showSettings: boolean, hostnames: string[];
</script>

<nav class="button-nav flex text-sm">
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
	<div class="donate ml-auto grid">
		<a
			target="_blank"
			href="https://www.buymeacoffee.com/tomdraper"
			class="donate-link text-[#464646]">{donateMessage}</a
		>
	</div>
	<button
		class="settings"
		on:click={() => {
			showSettings = true;
		}}
	>
		<img class="settings-icon" src="/images/icons/cog.png" alt="" />
	</button>
	<div class="dropdown-container mr-[10px]" class:no-display={hostnames.length <= 1}>
		<Dropdown
			options={hostnames.slice(0, 25)}
			bind:selected={settings.hostname}
			bind:open={dropdownOpen}
			defaultOption={'All hostnames'}
		/>
	</div>
	<div class="nav-btn time-period flex overflow-hidden rounded-[4px] border border-[#2e2e2e]">
		{#each timePeriods as period}
			<button
				class="time-period-btn cursor-pointer border-none bg-[var(--background)] px-[12px] py-[4px] text-[var(--dim-text)]"
				class:time-period-btn-active={settings.period === period}
				on:click={() => {
					settings.period = period;
					dropdownOpen = false;
					setPeriodParam(period);
				}}
			>
				{timePeriodsDisplay[period]}
			</button>
		{/each}
	</div>
</nav>

<style scoped>
	.button-nav {
		margin: 2.5em 2rem 0;
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
		/* margin-left: auto; */
		display: grid;
		place-items: center;
		margin-right: 1rem;
	}

	.donate-link {
		transition: 0.1s;
	}
	.donate-link:hover {
		color: var(--highlight);
	}
	.settings-icon {
		width: 20px;
		height: 20px;
		filter: contrast(0.45);
		transition: 0.1s;
		max-width: 20px;
	}
	.settings-icon:hover {
		filter: contrast(0.01);
	}

	@media screen and (max-width: 1300px) {
		.button-nav {
			margin: 2.5em 3rem 0;
		}
	}

	@media screen and (max-width: 900px) {
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
