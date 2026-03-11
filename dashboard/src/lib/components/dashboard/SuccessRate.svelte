<script lang="ts">
	import { renderPlot, sparklineData, sparklineLayout } from '$lib/plotly';

	let { rate, buckets }: { rate: number | null; buckets: number[] } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && buckets) renderPlot(plotDiv, sparklineData(buckets), sparklineLayout());
	});
</script>

<div class="card">
	<div class="card-title">Success rate</div>
	<div
		class="value"
		class:red={rate !== null && rate <= 70}
		class:yellow={rate !== null && rate > 70 && rate < 90}
		class:green={rate === null || rate > 90}
	>
		{rate !== null ? `${rate.toFixed(1)}%` : 'N/A'}
	</div>
	<div id="plotly">
		<div id="plotDiv" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</div>

<style scoped>
	.card {
		width: calc(215px - 1em);
		margin: 0 0 0 1em;
		position: relative;
		overflow: hidden;
	}
	.value {
		padding: 0.55em;
		text-align: center;
		font-size: 1.8em;
		font-weight: 700;
		color: var(--yellow);
		position: inherit;
		z-index: 2;
	}

	.red {
		color: var(--red);
	}
	.yellow {
		color: var(--yellow);
	}
	.green {
		color: var(--highlight);
	}
	#plotly {
		position: absolute;
		width: 110%;
		bottom: 0;
		overflow: hidden;
		margin: 0 -5%;
		z-index: 0;
	}
	@media screen and (max-width: 1070px) {
		.card {
			width: auto;
			flex: 1;
		}
	}
</style>
