<script lang="ts">
	import { onMount } from 'svelte';
	import { periodToDays } from '../../../lib/period';
	import type { Period } from '../../../lib/settings';
	import { initFreqMap } from '../../../lib/activity';
	import { ColumnIndex } from '../../../lib/consts';
	import { Chart } from 'chart.js/auto';

	function getChartData(data: RequestsData, period: Period) {
		const responseTimesFreq = initFreqMap(period, () => {
			return { totalResponseTime: 0, count: 0 };
		});

		const days = periodToDays(period);

		for (let i = 0; i < data.length; i++) {
			const date = new Date(data[i][ColumnIndex.CreatedAt]);
			if (days === 1) {
				const minutePeriod = 15;
				date.setMinutes(
					Math.floor(date.getMinutes() / minutePeriod) * minutePeriod,
					0,
					0,
				);
			} else if (days === 7) {
				const minutePeriod = 60;
				date.setMinutes(
					Math.floor(date.getMinutes() / minutePeriod) * minutePeriod,
					0,
					0,
				);
			} else {
				date.setHours(0, 0, 0, 0);
			}
			const time = date.getTime();
			if (responseTimesFreq.has(time)) {
				responseTimesFreq.get(time).totalResponseTime +=
					data[i][ColumnIndex.ResponseTime];
				responseTimesFreq.get(time).count++;
			} else {
				responseTimesFreq.set(time, { totalResponseTime: 1, count: 1 });
			}
		}

		// Combine date and avg response time into (x, y) tuples for sorting
		const responseTimeArr: { date: number; avgResponseTime: number }[] =
			new Array(responseTimesFreq.size);
		let i = 0;
		for (const [time, obj] of responseTimesFreq.entries()) {
			const point = { date: time, avgResponseTime: 0 };
			if (obj.count > 0) {
				point.avgResponseTime = obj.totalResponseTime / obj.count;
			}
			responseTimeArr[i] = point;
			i++;
		}

		// Sort by date
		responseTimeArr.sort((a, b) => {
			//@ts-ignore
			return a.date - b.date;
		});

		// Split into two lists
		const dates: Date[] = new Array(responseTimeArr.length);
		const responseTimes: number[] = new Array(responseTimeArr.length);
		let minAvgResponseTime = Number.POSITIVE_INFINITY;
		for (let i = 0; i < responseTimeArr.length; i++) {
			dates[i] = new Date(responseTimeArr[i].date);
			responseTimes[i] = responseTimeArr[i].avgResponseTime;
			if (responseTimeArr[i].avgResponseTime < minAvgResponseTime) {
				minAvgResponseTime = responseTimeArr[i].avgResponseTime;
			}
		}

		return {
			labels: dates,
			datasets: [
				{
					label: 'Requests',
					data: responseTimes,
					backgroundColor: '#707070',
					borderWidth: 0,
				},
			],
		};
	}

	function genPlot(data: RequestsData, period: Period) {
		const chartData = getChartData(data, period);

		const ctx = chartCanvas.getContext('2d');
		chart = new Chart(ctx, {
			type: 'bar',
			data: chartData,
			options: {
				maintainAspectRatio: false,
				layout: {
					padding: {
						top: 20,
						left: 10,
						right: 40,
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
						title: {
							display: true,
							text: 'Response time (ms)',
							color: '#505050',
						},
						ticks: {
							color: '#505050',
						},
					},
					x: {
						grid: {
							display: false,
						},
						border: {
							color: '#505050',
							width: 1,
						},
						ticks: {
							display: false,
						},
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
								return `${context.formattedValue}ms average`;
							},
							footer: function (context) {
								const date = new Date(
									context[0].label,
								).toLocaleString();
								return date;
							},
						},
					},
				},
			},
		});
	}

	function updatePlot(data: RequestsData, period: Period) {
		if (chart === null) {
			return;
		}
		chart.data = getChartData(data, period);
		chart.update();
	}

	let chart: Chart<'bar'> | null = null;
	let chartCanvas: HTMLCanvasElement;
	onMount(() => {
		genPlot(data, period);
	});

	$: if (data) {
		updatePlot(data, period);
	}

	export let data: RequestsData, period: Period;
</script>

<div id="plotly">
	<canvas bind:this={chartCanvas} id="chart"></canvas>
</div>

<style scoped>
	#chart {
		height: 159px !important;
		width: 100% !important;
	}
</style>
