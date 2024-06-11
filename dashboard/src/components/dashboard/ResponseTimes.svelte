<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex } from '../../lib/consts';
	import { Chart } from 'chart.js/auto';

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

	function setQuartiles(data: RequestsData) {
		const responseTimes: number[] = new Array(data.length);
		for (let i = 0; i < data.length; i++) {
			responseTimes[i] = data[i][ColumnIndex.ResponseTime];
		}
		responseTimes.sort((a, b) => a - b);
		quartiles = {
			lq: quantile(responseTimes, 0.25),
			median: quantile(responseTimes, 0.5),
			uq: quantile(responseTimes, 0.75),
		};
	}

	function getChartData(data: RequestsData) {
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

		return {
			labels: responseTimes,
			datasets: [
				{
					label: 'Requests',
					data: counts,
					backgroundColor: '#707070',
					borderColor: '#707070',
					borderWidth: 1,
				},
			],
		};
	}

	function genPlot(data: RequestsData) {
		const chartData = getChartData(data);

		const ctx = chartCanvas.getContext('2d');
		chart = new Chart(ctx, {
			type: 'bar',
			data: chartData,
			options: {
				maintainAspectRatio: false,
				layout: {
					padding: {
						left: 10,
						right: 10,
					},
				},
				scales: {
					y: {
						grid: {
							display: false,
						},
						border: {
							display: false,
						},
						beginAtZero: true,
						ticks: {
							display: false,
						},
					},
					x: {
						grid: {
							display: false,
						},
						border: {
							display: false,
						},
						ticks: {
							display: false,
						},
						beginAtZero: true,
					},
				},
				plugins: {
					legend: {
						display: false,
					},
					tooltip: {
						callbacks: {
							title: () => null,
							label: function (context) {
								return `${context.label}ms: ${context.formattedValue} requests`;
							},
						},
					},
				},
			},
		});
	}

	function updatePlot(data: RequestsData) {
		if (chart === null) {
			return;
		}
		chart.data = getChartData(data);
		chart.update();
	}

	type Quartiles = {
		lq: number;
		median: number;
		uq: number;
	};

	let chart: Chart<'bar'> | null = null;
	let chartCanvas: HTMLCanvasElement;
	let quartiles: Quartiles;
	onMount(() => {
		setQuartiles(data);
		genPlot(data);
	});

	$: if (data) {
		setQuartiles(data);
		updatePlot(data);
	}

	export let data: RequestsData;
</script>

<div class="card">
	<div class="card-title">
		Response times <span class="milliseconds">(ms)</span>
	</div>
	{#if quartiles !== undefined}
		<div class="values">
			<div class="value lower-quartile">{quartiles.lq.toFixed(1)}</div>
			<div class="value median">{quartiles.median.toFixed(1)}</div>
			<div class="value upper-quartile">{quartiles.uq.toFixed(1)}</div>
		</div>
	{/if}
	<div class="labels">
		<div class="label">LQ</div>
		<div class="label">Median</div>
		<div class="label">UQ</div>
	</div>
	<div class="distribution">
		<div id="plotly">
			<canvas bind:this={chartCanvas} id="chart"></canvas>
		</div>
	</div>
</div>

<style scoped>
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
	#chart {
		height: 50px !important;
		width: 100% !important;
	}

	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
