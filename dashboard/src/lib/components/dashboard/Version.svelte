<script lang="ts">
	import { ColumnIndex, graphColors } from '$lib/consts';

	function getVersionCount(data: RequestsData): { versions: Set<string>; count: ValueCount } {
		const versions = new Set<string>();
		const count: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const match = data[i][ColumnIndex.Path].match(/[^a-z0-9](v\d)[^a-z0-9]/i);
			if (match) {
				const v = match[1];
				versions.add(v);
				count[v] = (count[v] ?? 0) + 1;
			}
		}
		return { versions, count };
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

	function pieChart(versionCount: ValueCount) {
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

	function getPlotData(versionCount: ValueCount) {
		return {
			data: pieChart(versionCount),
			layout: getLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function genPlot(versionCount: ValueCount) {
		const plotData = getPlotData(versionCount);
		//@ts-ignore
		new Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	let { data }: { data: RequestsData } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);
	const versionData = $derived(data ? getVersionCount(data) : undefined);

	$effect(() => {
		if (plotDiv && versionData && versionData.versions.size > 1) {
			genPlot(versionData.count);
			window?.dispatchEvent(new Event('resize'));
		}
	});
</script>

<div class="card flex-1" class:hidden={versionData === undefined || versionData.versions.size <= 1}>
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
