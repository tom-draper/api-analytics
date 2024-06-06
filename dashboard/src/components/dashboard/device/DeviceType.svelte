<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex, graphColors } from '../../../lib/consts';
	import { cachedFunction } from '../../../lib/cache';
	import { Chart } from 'chart.js/auto';

	const deviceCandidates = [
		{name: 'iPhone', regex: /iPhone/, matches: 0},
		{name: 'Android', regex: /Android/, matches: 0},
		{name: 'Samsung', regex: /Tizen\//, matches: 0},
		{name: 'Mac', regex: /Macintosh/, matches: 0},
		{name: 'Windows', regex: /Windows/, matches: 0},
	]

	function getDevice(userAgent: string): string {
		if (userAgent === null) {
			return 'Unknown';
		}

		for (let i = 0; i < deviceCandidates.length; i++) {
			const candidate = deviceCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
				// Ensure deviceCandidates remains sorted by matches desc for future hits
				maintainCandidates(i, deviceCandidates);
				return candidate.name;
			}
		}

		return 'Other';
	}

	function maintainCandidates(indexUpdated: number, candidates: {name: string, regex: string, matches: number}[]) {
		let j = indexUpdated;
    	while (j > 0 && count > candidates[j - 1].matches) {
        	j--
    	}
    	if (j < indexUpdated) {
        	[candidates[indexUpdated], candidates[j]] = [candidates[j], candidates[indexUpdated]]
    	}
	}

	function getChartData() {
		const deviceCount: ValueCount = {};
		const deviceGetter = cachedFunction(getDevice);
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const device = deviceGetter(userAgent);
			if (device in deviceCount) {
				deviceCount[device]++;
			} else {
				deviceCount[device] = 1;
			}
		}

		const dataPoints = Object.entries(clientCount).sort(
			(a, b) => b[1] - a[1],
		);

		const devices = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [browser, count] of dataPoints) {
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
		const data = getChartData();

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
