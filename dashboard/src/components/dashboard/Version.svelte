<script lang="ts">
	import { ColumnIndex, graphColors } from '../../lib/consts';
	import { onMount } from 'svelte';
	import { Chart } from 'chart.js/auto';

	function getVersions() {
		const versions = new Set<string>();
		for (let i = 0; i < data.length; i++) {
			const match = data[i][ColumnIndex.Path].match(
				/[^a-z0-9](v\d)[^a-z0-9]/i,
			);
			if (match) {
				versions.add(match[1]);
			}
		}
		return versions;
	}

	function getChartData() {
		const versionCount: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const match = data[i][ColumnIndex.Path].match(
				/[^a-z0-9](v\d)[^a-z0-9]/i,
			);
			if (!match) {
				continue;
			}
			const version = match[1];
			if (version in versionCount) {
				versionCount[version]++;
			} else {
				versionCount[version] = 1;
			}
		}

		const dataPoints = Object.entries(versionCount).sort(
			(a, b) => b[1] - a[1],
		);

		const versions = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [browser, count] of dataPoints) {
			versions[i] = browser;
			counts[i] = count;
			i++;
		}

		return {
			labels: versions,
			datasets: [
				{
					label: 'Version',
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
		chart = new Chart(ctx, {
			type: 'doughnut',
			data: data,
			options: {
				maintainAspectRatio: false,
				borderWidth: 0,
				layout: {
					padding: {
						right: 10,
					},
				},
				plugins: {
					legend: {
						position: 'right',
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
	let versions: Set<string> = new Set();

	onMount(() => {
		versions = getVersions();
		if (versions.size > 1) {
			genPlot();
		}
	});

	$: if (data) {
		updatePlot();
	}

	export let data: RequestsData;
</script>

<div class="card" class:hidden={versions.size <= 1}>
	<div class="card-title">Version</div>
	<div id="plotly">
		<canvas bind:this={chartCanvas} id="chart"></canvas>
	</div>
</div>

<style scoped>
	.card {
		margin: 2em 0 2em 0;
		padding-bottom: 1em;
		flex: 1;
	}
	#chart {
		height: 180px !important;
		width: 100% !important;
	}
	.hidden {
		visibility: hidden;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
