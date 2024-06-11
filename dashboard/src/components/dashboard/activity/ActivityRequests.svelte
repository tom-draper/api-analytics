<script lang="ts">
	import { onMount } from 'svelte';
	import { periodToDays } from '../../../lib/period';
	import type { Period } from '../../../lib/settings';
	import { initFreqMap } from '../../../lib/activity';
	import { ColumnIndex } from '../../../lib/consts';
	import { Chart } from 'chart.js/auto';

	function getChartData(data: RequestsData, period: Period) {
		const requestFreq = initFreqMap(period, () => ({
			count: 0,
		}));
		const userFreq = initFreqMap(period, () => new Set());

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
			const ipAddress = data[i][ColumnIndex.IPAddress];
			const time = date.getTime();
			if (userFreq.has(time)) {
				userFreq.get(time).add(ipAddress);
			} else {
				userFreq.set(time, new Set());
			}

			if (requestFreq.has(time)) {
				requestFreq.get(time).count++;
			} else {
				requestFreq.set(time, { count: 1 });
			}
		}

		// Combine date and frequency count into (x, y) tuples for sorting
		const requestFreqArr = new Array(requestFreq.size);
		let i = 0;
		for (const [time, requestsCount] of requestFreq.entries()) {
			const userCount = userFreq.has(time) ? userFreq.get(time).size : 0;
			requestFreqArr[i] = {
				date: time,
				requestCount: requestsCount.count,
				userCount: userCount,
			};
			i++;
		}
		// Sort by date
		requestFreqArr.sort((a, b) => {
			return a.date - b.date;
		});

		// Split into two lists
		const dates = new Array(requestFreqArr.length);
		const requests = new Array(requestFreqArr.length);
		const requestsText = new Array(requestFreqArr.length);
		const users = new Array(requestFreqArr.length);
		const usersText = new Array(requestFreqArr.length);
		for (let i = 0; i < requestFreqArr.length; i++) {
			dates[i] = new Date(requestFreqArr[i].date);
			// Subtract users due to bar stacking
			requests[i] =
				requestFreqArr[i].requestCount - requestFreqArr[i].userCount;

			// Keep actual requests count for hover text
			requestsText[i] = `${requestFreqArr[i].requestCount} requests`;
			users[i] = requestFreqArr[i].userCount;
			usersText[i] =
				`${requestFreqArr[i].userCount} users from ${requestFreqArr[i].requestCount} requests`;
		}

		return {
			labels: dates,
			datasets: [
				{
					label: 'Users',
					data: users,
					backgroundColor: '#3fcf8e',
					borderWidth: 0,
				},
				{
					label: 'Requests',
					data: requests,
					backgroundColor: '#228458',
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
						stacked: true,
						grid: {
							display: false,
						},
						border: {
							display: false,
						},
						beginAtZero: true,
						title: {
							display: true,
							text: 'Requests',
							color: '#505050',
						},
						ticks: {
							color: '#505050',
						},
					},
					x: {
						stacked: true,
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
								const datasetName = context.dataset.label;

								if (datasetName === 'Requests') {
									const usersDataset =
										context.chart.data.datasets.filter(
											(x) => x.label === 'Users',
										);
									const dataIndex = context.dataIndex;
									const usersCount =
										usersDataset[0].data[dataIndex];
									return `${usersCount + context.raw} requests`;
								} else {
									return `${context.raw} users`;
								}
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
