<script lang="ts">
	import { ColumnIndex } from '$lib/consts';

	/* Parameter `arr` assumed sorted. */
	function quantile(arr: number[], q: number) {
		const pos = (arr.length - 1) * q;
		const base = Math.floor(pos);
		const rest = pos - base;
		if (arr[base + 1] !== undefined) {
			return arr[base] + rest * (arr[base + 1] - arr[base]);
		} else if (arr[base] !== undefined) {
			return arr[base];
		}
		return 0;
	}

	function calcualteMetrics(data: RequestsData) {
		const responseTimes: number[] = new Array(data.length);
		for (let i = 0; i < data.length; i++) {
			responseTimes[i] = data[i][ColumnIndex.ResponseTime];
		}
		responseTimes.sort((a, b) => a - b);
		LQ = quantile(responseTimes, 0.25);
		median = quantile(responseTimes, 0.5);
		UQ = quantile(responseTimes, 0.75);
	}

	function getPlotLayout(range: [number, number]) {
		return {
			title: false,
			autosize: true,
			margin: { r: 0, l: 0, t: 5, b: 0, pad: 10 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 50,
			yaxis: {
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
				visible: false,
			},
			xaxis: {
				range: range,
				showgrid: false,
				fixedrange: true,
				visible: false,
			},
			dragmode: false,
		};
	}

	function bars(data: RequestsData) {
		const responseTimesFreq: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const responseTime =
				Math.round(data[i][ColumnIndex.ResponseTime]) || 0;
			if (responseTime in responseTimesFreq) {
				responseTimesFreq[responseTime]++;
			} else {
				responseTimesFreq[responseTime] = 1;
			}
		}

		const responseTimes: number[] = [];
		const counts: number[] = [];
		const times = Object.keys(responseTimesFreq).map(Number);
		if (times.length > 0) {
			const minResponseTime = Math.min(...times);
			const maxResponseTime = Math.max(...times);

			// Split into two lists
			for (let i = 0; i < maxResponseTime - minResponseTime + 1; i++) {
				responseTimes.push(minResponseTime + i);
				counts.push(responseTimesFreq[minResponseTime + i] || 0);
			}
		}

		return [
			{
				x: responseTimes,
				y: counts,
				type: 'bar',
				marker: { color: '#505050' },
				hovertemplate: `<b>%{y} requests</b><br>%{x:.1f}ms</b> elapsed<extra></extra>`,
				showlegend: false,
			},
		];
	}

	function getPlotData(data: RequestsData) {
		const b = bars(data);
		return {
			data: b,
			layout: getPlotLayout([b[0].x[0], b[0].x[b[0].x.length - 1]]),
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
		const b = bars(data);
		Plotly.react(
			plotDiv,
			bars(data),
			getPlotLayout([b[0].x[0], b[0].x[b[0].x.length - 1]]),
		)
	}

	let median: number;
	let LQ: number;
	let UQ: number;
	let plotDiv: HTMLDivElement;

	$: if (data) {
		calcualteMetrics(data);
	}

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData;
</script>

<div class="card">
	<div class="card-title">
		Response times <span class="milliseconds">(ms)</span>
	</div>
	{#if LQ !== undefined && median !== undefined && UQ !== undefined}
		<div class="values">
			<div class="value lower-quartile">{LQ.toFixed(1)}</div>
			<div class="value median">{median.toFixed(1)}</div>
			<div class="value upper-quartile">{UQ.toFixed(1)}</div>
		</div>
	{/if}
	<div class="labels">
		<div class="label">LQ</div>
		<div class="label">Median</div>
		<div class="label">UQ</div>
	</div>
	<div class="distribution">
		<div id="plotly">
			<div id="plotDiv" bind:this={plotDiv}>
				<!-- Plotly chart will be drawn inside this DIV -->
			</div>
		</div>
	</div>
</div>

<style scoped>
	.card {
		overflow: hidden;
	}
	.values {
		display: flex;
		color: var(--highlight);
		font-size: 1.8em;
		font-weight: 700;
	}
	.values,
	.labels {
		margin: 0 0.5rem;
	}
	.value {
		flex: 1;
		font-size: 1.1em;
		padding: 20px 20px 4px;
	}
	.labels {
		display: flex;
		font-size: 0.8em;
		color: var(--dim-text);
	}
	.label {
		flex: 1;
	}

	.milliseconds {
		color: var(--dim-text);
		font-size: 0.8em;
		margin-left: 4px;
	}

	.median {
		font-size: 1em;
	}
	.upper-quartile,
	.lower-quartile {
		font-size: 1em;
		padding-bottom: 0;
	}

	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
