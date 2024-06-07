<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex } from '../../lib/consts';
	import { Chart } from 'chart.js/auto';

	function getChartData() {
		const responseTimes = Array(24).fill(0);

		for (let i = 0; i < data.length; i++) {
			const date = data[i][ColumnIndex.CreatedAt];
			const time = date.getHours();
			responseTimes[time]++;
		}

		const requestFreqArr = responseTimes
			.map((requestCount, i) => ({
				hour: i,
				responseTime: requestCount,
			}))
			.sort((a, b) => {
				return a.hour - b.hour;
			});

		let dates = new Array(requestFreqArr.length);
		let requests = new Array(requestFreqArr.length);
		for (let i = 0; i < requestFreqArr.length; i++) {
			dates[i] = requestFreqArr[i].hour.toString() + ':00';
			requests[i] = requestFreqArr[i].responseTime;
		}

		// Shift to 12 onwards to make barpolar like clock face
		dates = dates.slice(12).concat(...dates.slice(0, 12));
		requests = requests.slice(12).concat(...requests.slice(0, 12));

		return {
			labels: dates,
			datasets: [
				{
					label: 'Usage',
					data: requests,
					backgroundColor: ['#3FCF8E'],
				},
			],
		};
	}

	function genPlot() {
		const data = getChartData();

		let ctx = chartCanvas.getContext('2d');
		chart = new Chart(ctx, {
			type: 'polarArea',
			data: data,
			options: {
				responsive: true,
				maintainAspectRatio: false,
				borderWidth: 0,
				layout: {
					padding: {
						left: 20,
						right: 20,
						bottom: 20,
					},
				},
				scales: {
					r: {
						pointLabels: {
							display: true,
							centerPointLabels: true,
							font: {
								size: 12,
							},
						},
						ticks: {
							display: false, // Hides the number labels
						},
					},
				},
				plugins: {
					legend: {
						display: false,
					},
					title: {
						display: false,
					},
				},
			},
		});
	}

	function updatePlot() {
		if (chart === null) {
			return;
		}
		chart.data = getChartData();
		chart.update();
	}

	let chart: Chart | null = null;
	let chartCanvas: HTMLCanvasElement;
	onMount(() => {
		genPlot();
	});

	$: if (data) {
		updatePlot();
	}

	export let data: RequestsData;
</script>

<div class="card">
	<div class="card-title">Usage time</div>
	<div id="plotly">
		<canvas bind:this={chartCanvas} id="chart"></canvas>
	</div>
</div>

<style scoped>
	.card {
		width: 100%;
		margin: 0;
	}
	#chart {
		height: 500px !important;
		width: 100% !important;
		margin: auto;
	}
</style>
