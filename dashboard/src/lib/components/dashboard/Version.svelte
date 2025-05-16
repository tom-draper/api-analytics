<script lang="ts">
	import { ColumnIndex, graphColors } from '$lib/consts';

	function getVersions(data: RequestsData) {
		const versions = new Set<string>();
		for (let i = 0; i < data.length; i++) {
			const match = data[i][ColumnIndex.Path].match(/\/(v\d)[^a-z0-9]/i);
			if (match) {
				versions.add(match[1]);
			}
		}
		return versions;
	}

	function build(data: RequestsData) {
		versions = getVersions(data);

		if (versions.size > 1) {
			genPlot(data);
			window?.dispatchEvent(new Event('resize'));
		}
	}

	function getLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 30, l: 30, t: 25, b: 25, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 196,
			yaxis: {
				title: { text: 'Requests' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true
			},
			xaxis: {
				visible: false
			},
			dragmode: false
		};
	}

	function pieChart(data: RequestsData) {
		const versionCount: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const match = data[i][ColumnIndex.Path].match(/[^a-z0-9](v\d)[^a-z0-9]/i);
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

		const versions = Object.keys(versionCount);
		const count = Object.values(versionCount);

		return [
			{
				values: count,
				labels: versions,
				type: 'pie',
				hole: 0.6,
				marker: {
					colors: graphColors
				}
			}
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: pieChart(data),
			layout: getLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot(data: RequestsData) {
		const plotData = getPlotData(data);
		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let versions: Set<string>;
	let plotDiv: HTMLDivElement;

	$: if (plotDiv && data) {
		build(data);
	}

	export let data: RequestsData;
</script>

<div class="card flex-1" class:hidden={versions === undefined || versions.size <= 1}>
	<div class="card-title">Version</div>
	<div id="plotly">
		<div id="plotDiv" class="mr-[20px]" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</div>

<style scoped>
	.card {
		margin: 2em 0 2em 0;
		/* padding-bottom: 1em; */
		flex: 1;
	}
	.hidden {
		display: none;
	}
	#plotDiv {
		padding-right: 20px;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
