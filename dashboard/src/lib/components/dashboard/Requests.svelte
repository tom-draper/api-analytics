<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';

	function getPlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
			hovermode: false,
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 60,
			yaxis: {
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
				dragmode: false
			},
			xaxis: {
				visible: false,
				dragmode: false
			},
			dragmode: false
		};
	}

	function lines(buckets: number[]) {
		return [
			{
				x: [...Array(buckets.length).keys()],
				y: buckets,
				type: 'lines',
				marker: { color: 'transparent' },
				showlegend: false,
				line: { shape: 'spline', smoothing: 1, color: '#3FCF8E30' },
				fill: 'tozeroy',
				fillcolor: '#3fcf8e15'
			}
		];
	}

	function generatePlot(buckets: number[]) {
		if (plotDiv.data) {
			Plotly.react(plotDiv, lines(buckets), getPlotLayout());
		} else {
			const plotData = {
				data: lines(buckets),
				layout: getPlotLayout(),
				config: { responsive: true, showSendToCloud: false, displayModeBar: false }
			};
			Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
		}
	}

	function getPercentageChange(count: number, prevCount: number) {
		if (prevCount === 0) return null;
		return (count / prevCount) * 100 - 100;
	}

	function getRequestsPerHour(count: number, period: Period, firstDate: Date | null, lastDate: Date | null) {
		if (count <= 1) return count;
		let days = periodToDays(period);
		if (days === null && firstDate && lastDate) {
			const diff = lastDate.getTime() - firstDate.getTime();
			days = Math.floor(diff / (1000 * 60 * 60 * 24));
		}
		if (!days) return count;
		return count / (24 * days);
	}

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
	const requestsPerHour = $derived(getRequestsPerHour(count, period, firstDate, lastDate));

	$effect(() => {
		if (plotDiv && buckets) generatePlot(buckets);
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
