<script lang="ts">
	import { graphColors } from '$lib/consts';

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

	function pieChart(versionCount: { [v: string]: number }) {
		const versions = Object.keys(versionCount);
		const counts = Object.values(versionCount);
		return [
			{
				values: counts,
				labels: versions,
				type: 'pie',
				hole: 0.6,
				marker: { colors: graphColors }
			}
		];
	}

	function genPlot(versionCount: { [v: string]: number }) {
		const data = pieChart(versionCount);
		const layout = getLayout();
		const config = { responsive: true, showSendToCloud: false, displayModeBar: false };
		//@ts-ignore
		new Plotly.newPlot(plotDiv, data, layout, config);
	}

	let { versionCount, hasMultiple }: { versionCount: { [v: string]: number }; hasMultiple: boolean } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && hasMultiple) {
			genPlot(versionCount);
			window?.dispatchEvent(new Event('resize'));
		}
	});
</script>

<div class="card flex-1" class:hidden={!hasMultiple}>
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
