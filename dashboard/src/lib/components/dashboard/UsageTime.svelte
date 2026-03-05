<script lang="ts">
	import { renderPlot } from '$lib/plotly';

	function getPlotLayout() {
		return {
			font: { size: 12 },
			paper_bgcolor: 'transparent',
			height: 500,
			margin: { r: 50, l: 50, t: 20, b: 50, pad: 0 },
			polar: {
				bargap: 0,
				bgcolor: 'transparent',
				angularaxis: { direction: 'clockwise', showgrid: false },
				radialaxis: { gridcolor: '#303030' }
			}
		};
	}

	function bars(hourlyBuckets: number[]) {
		// Shift to 12 onwards to make barpolar like clock face
		const dates = Array.from({ length: 24 }, (_, i) => i.toString() + ':00');
		const shiftedDates = dates.slice(12).concat(...dates.slice(0, 12));
		const shiftedBuckets = hourlyBuckets.slice(12).concat(...hourlyBuckets.slice(0, 12));

		return [
			{
				r: shiftedBuckets,
				theta: shiftedDates,
				marker: { color: '#3fcf8e' },
				type: 'barpolar',
				hovertemplate: `<b>%{r}</b> requests at <b>%{theta}</b><extra></extra>`
			}
		];
	}

	function generatePlot(hourlyBuckets: number[]) {
		renderPlot(plotDiv, bars(hourlyBuckets), getPlotLayout());
	}

	let { hourlyBuckets }: { hourlyBuckets: number[] } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && hourlyBuckets) generatePlot(hourlyBuckets);
	});
</script>

<div class="card">
	<div class="card-title">Usage time</div>
	<div id="plotly">
		<div id="plotDiv" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</div>

<style scoped>
	.card {
		width: 100%;
		margin: 0;
	}
</style>
