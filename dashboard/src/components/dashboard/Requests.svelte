<script lang="ts">
	import { periodToDays } from '../../lib/period';
	import type { Period } from '../../lib/settings';
	import { ColumnIndex } from '../../lib/consts';

	function requestsPlotLayout() {
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
				dragmode: false,
			},
			xaxis: {
				visible: false,
				dragmode: false,
			},
			dragmode: false,
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
				const idx = Math.floor(diff / (range / n));
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
				fillcolor: '#3fcf8e15',
			},
		];
	}

	function requestsPlotData(data: RequestsData) {
		return {
			data: lines(data),
			layout: requestsPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot(data: RequestsData) {
		const plotData = requestsPlotData(data);
		//@ts-ignore
		new Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	function getPercentageChange(data: RequestsData) {
		let percentageChange: number | null;
		if (prevData.length == 0) {
			percentageChange = null;
		} else {
			percentageChange = (data.length / prevData.length) * 100 - 100;
		}
		return percentageChange;
	}

	function getRequestsPerHour(data: RequestsData) {
		let requestsPerHour: number = 0;
		if (data.length > 0) {
			const days = periodToDays(period);
			if (days != null) {
				requestsPerHour = data.length / (24 * days);
			}
		}
		return requestsPerHour;
	}

	function togglePeriod() {
		perHour = !perHour;
	}

	function build(data: RequestsData) {
		percentageChange = getPercentageChange(data);
		requestsPerHour = getRequestsPerHour(data);
		genPlot(data);
	}

	let plotDiv: HTMLDivElement;
	let requestsPerHour: number;
	let percentageChange: number;
	let perHour = false;

	$: if (plotDiv && data) {
		build(data);
	}

	export let data: RequestsData, prevData: RequestsData, period: Period;
</script>

<button class="card" on:click={togglePeriod}>
	{#if perHour}
		<div class="card-title">
			Requests <span class="per-hour">/ hour</span>
		</div>
		{#if requestsPerHour}
			<div class="value">{requestsPerHour.toFixed(2)}</div>
		{/if}
	{:else}
		{#if percentageChange}
			<div
				class="percentage-change"
				class:positive={percentageChange > 0}
				class:negative={percentageChange < 0}
			>
				{#if percentageChange > 0}
					<img class="arrow" src="../img/icons/green-up.png" alt="" />
				{:else if percentageChange < 0}
					<img class="arrow" src="../img/icons/red-down.png" alt="" />
				{/if}
				{Math.abs(percentageChange).toFixed(1)}%
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
		padding: 20px 10px;
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
