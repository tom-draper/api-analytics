<script lang="ts">
	import { periodToDays } from '$lib/period';
	import type { Period } from '$lib/settings';
	import { ColumnIndex } from '$lib/consts';

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

	function lines(data: RequestsData) {
		const n = 5;
		const x = [...Array(n).keys()];
		const y = Array(n).fill(0);

		if (data.length > 0) {
			const start = data[0][ColumnIndex.CreatedAt].getTime();
			const end = data[data.length - 1][ColumnIndex.CreatedAt].getTime();
			const range = end - start;
			for (let i = 0; i < data.length; i++) {
				const time = data[i][ColumnIndex.CreatedAt].getTime();
				const diff = time - start;
				// Make sure idx doesn't exceed n - 1
				const idx = Math.min(Math.floor(diff / (range / n)), n - 1);
				y[idx] += 1;
			}
		}

		return [
			{
				x: x,
				y: y,
				type: 'lines',
				marker: { color: 'transparent' },
				showlegend: false,
				line: { shape: 'spline', smoothing: 1, color: '#3FCF8E30' },
				fill: 'tozeroy',
				fillcolor: '#3fcf8e15'
			}
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: lines(data),
			layout: getPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function generatePlot(data: RequestsData) {
		if (plotDiv.data) {
			refreshPlot(data);
		} else {
			newPlot(data);
		}
	}

	async function newPlot(data: RequestsData) {
		const plotData = getPlotData(data);
		Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	function refreshPlot(data: RequestsData) {
		Plotly.react(plotDiv, lines(data), getPlotLayout());
	}

	function getPercentageChange(data: RequestsData) {
		if (prevData.length == 0) {
			return null;
		}

		return (percentageChange = (data.length / prevData.length) * 100 - 100);
	}

	function getRequestsPerHour(data: RequestsData) {
		if (data.length === 0) {
			return 0;
		}
		if (data.length === 1) {
			return 1;
		}

		let days = periodToDays(period);
		if (days === null) {
			days = daysBetween(
				data[0][ColumnIndex.CreatedAt],
				data[data.length - 1][ColumnIndex.CreatedAt]
			);
		}
		return data.length / (24 * days);
	}

	function daysBetween(date1: Date, date2: Date) {
		const diff = date2.getTime() - date1.getTime();
		return Math.floor(diff / (1000 * 60 * 60 * 24));
	}

	function togglePeriod() {
		perHour = !perHour;
	}

	let plotDiv: HTMLDivElement;
	let requestsPerHour: number;
	let percentageChange: number | null;
	let perHour = false;

	$: if (data) {
		percentageChange = getPercentageChange(data);
		requestsPerHour = getRequestsPerHour(data);
	}

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData, prevData: RequestsData, period: Period;
</script>

<button class="card" on:click={togglePeriod}>
	{#if perHour}
		<div class="card-title">
			Requests <span class="per-hour">/ hour</span>
		</div>
		{#if requestsPerHour !== undefined}
			<div class="value">{requestsPerHour === 0 ? '0' : requestsPerHour.toFixed(2)}</div>
		{/if}
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
		<div class="value">{data.length.toLocaleString()}</div>
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
