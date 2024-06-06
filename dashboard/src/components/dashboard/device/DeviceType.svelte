<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex, graphColors } from '../../../lib/consts';
	import { Chart } from 'chart.js/auto';

	function getDevice(userAgent: string): string {
		if (userAgent === null) {
			return 'Unknown';
		} else if (userAgent.match(/iPhone/)) {
			return 'iPhone';
		} else if (userAgent.match(/Android/)) {
			return 'Android';
		} else if (userAgent.match(/Tizen\//)) {
			return 'Samsung';
		} else if (userAgent.match(/Macintosh/)) {
			return 'Mac';
		} else if (userAgent.match(/Windows/)) {
			return 'PC';
		} else {
			return 'Other';
		}
	}

	const colors = [
		'#3FCF8E', // Green
		'#E46161', // Red
		'#EBEB81', // Yellow
	];

	function pieChart() {
		const deviceCount: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const device = getDevice(userAgent);
			if (device in deviceCount) {
				deviceCount[device]++;
			} else {
				deviceCount[device] = 1;
			}
		}

		const devices = new Array(Object.keys(deviceCount).length);
		const counts = new Array(Object.keys(deviceCount).length);
		let i = 0;
		for (const [browser, count] of Object.entries(deviceCount)) {
			devices[i] = browser;
			counts[i] = count;
			i++;
		}

		return {
			labels: devices,
			datasets: [
				{
					label: 'Device Type',
					data: counts,
					backgroundColor: graphColors,
					hoverOffset: 4,
				},
			],
		};
	}

	function genPlot() {
		const data = pieChart();

		let ctx = chartCanvas.getContext('2d');
		let chart = new Chart(ctx, {
			type: 'doughnut',
			data: data,
			options: {
				maintainAspectRatio: false,
				borderWidth: 0,
				plugins: {
					legend: {
						position: 'right',
					},
				},
			},
		});
	}

	let chartCanvas: HTMLCanvasElement;
	let mounted = false;
	onMount(() => {
		mounted = true;
	});

	$: data && mounted && genPlot();

	export let data: RequestsData, getUserAgent: (id: number) => string;
</script>

<div id="plotly">
	<canvas bind:this={chartCanvas} id="chart"></canvas>
</div>

<style>
	#chart {
		height: 180px !important;
		width: 100% !important;
	}
</style>
