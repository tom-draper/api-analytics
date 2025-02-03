<script lang="ts">
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

		for (let i = 0; i < data.length; i++) {
			const idx = Math.min(Math.floor(i / (data.length / n)), n - 1);

			if (
				data[i][ColumnIndex.Status] >= 200 &&
				data[i][ColumnIndex.Status] <= 299
			) {
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

	function getPlotData(data: RequestsData) {
		return {
			data: lines(data),
			layout: getPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
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
		Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	function refreshPlot(data: RequestsData) {
		Plotly.react(
			plotDiv,
			lines(data),
			getPlotLayout(),
		)
	}

	function getSuccessRate(data: RequestsData) {
		if (!data || data.length === 0) {
			return null;
		}

		let successfulRequests = 0;
		for (let i = 0; i < data.length; i++) {
			if (
				data[i][ColumnIndex.Status] >= 200 &&
				data[i][ColumnIndex.Status] <= 299
			) {
				successfulRequests++;
			}
		}

		return (successfulRequests / data.length) * 100;
	}

	let plotDiv: HTMLDivElement;
	let successRate: number | null;

	$: if (data) {
		successRate = getSuccessRate(data);
	}

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData;
</script>

<div class="card">
	<div class="card-title">Success rate</div>
	<div
		class="value"
		class:red={successRate !== null && successRate <= 75}
		class:yellow={successRate !== null && successRate > 75 && successRate < 90}
		class:green={successRate === null || successRate > 90}
	>
		{successRate ? `${successRate.toFixed(1)}%` : 'N/A'}
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
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
		}
	}
</style>
