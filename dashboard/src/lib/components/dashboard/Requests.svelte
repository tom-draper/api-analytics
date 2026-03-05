<script lang="ts">
	import { type Period } from '$lib/period';
	import { renderPlot, sparklineData, sparklineLayout } from '$lib/plotly';
	import { getPercentageChange, countPerHour } from '$lib/utils';

	function togglePeriod() {
		perHour = !perHour;
	}

	let { buckets, count, prevCount, firstDate, lastDate, period }: {
		buckets: number[];
		count: number;
		prevCount: number;
		firstDate: Date | null;
		lastDate: Date | null;
		period: Period;
	} = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);
	let perHour = $state(false);
	const percentageChange = $derived(getPercentageChange(count, prevCount));
	const requestsPerHour = $derived(count <= 1 ? count : countPerHour(count, period, firstDate, lastDate));

	$effect(() => {
		if (plotDiv && buckets) renderPlot(plotDiv, sparklineData(buckets), sparklineLayout());
	});
</script>

<button class="card" onclick={togglePeriod}>
	{#if perHour}
		<div class="card-title">
			Requests <span class="per-hour">/ hour</span>
		</div>
		<div class="value">{requestsPerHour === 0 ? '0' : requestsPerHour.toFixed(2)}</div>
	{:else}
		{#if percentageChange}
			<div
				class="percentage-change flex"
				class:positive={percentageChange > 0}
				class:negative={percentageChange < 0}
			>
				{#if percentageChange > 0}
					<img class="arrow" src="/images/icons/green-up.png" alt="" />
				{:else if percentageChange < 0}
					<img class="arrow" src="/images/icons/red-down.png" alt="" />
				{/if}
				<div>
					{Math.abs(percentageChange).toFixed(1)}%
				</div>
			</div>
		{/if}
		<div class="card-title">Requests</div>
		<div class="value">{count.toLocaleString()}</div>
	{/if}
	<div id="plotly">
		<div id="plotDiv" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</button>

<style scoped>
	.card {
		width: calc(215px - 1em);
		margin: 0 1em 0 0;
		position: relative;
		cursor: pointer;
		padding: 0;
		overflow: hidden;
	}
	.value {
		padding: 0.55em 0.2em;
		font-size: 1.8em;
		font-weight: 700;
		position: inherit;
		z-index: 2;
	}
	.percentage-change {
		position: absolute;
		right: 20px;
		top: 20px;
		font-size: 0.8em;
	}
	.positive {
		color: var(--highlight);
	}
	.negative {
		color: rgb(228, 97, 97);
	}

	.per-hour {
		color: var(--dim-text);
		font-size: 0.8em;
		margin-left: 4px;
	}
	button {
		font-size: unset;
		font-family: unset;
		font-family: 'Noto Sans' !important;
	}
	.arrow {
		height: 11px;
		align-self: center;
		margin-right: 0.25em;
	}
	#plotly {
		position: absolute;
		width: 110%;
		bottom: 0;
		overflow: hidden;
		margin: 0 -5%;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 1em 0 0;
		}
	}
</style>
