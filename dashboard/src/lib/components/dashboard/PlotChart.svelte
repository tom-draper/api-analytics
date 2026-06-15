<script lang="ts">
	import { renderPlot, type PlotlyDiv } from '$lib/plotly';

	let { data, layout }: { data: object[]; layout: object } = $props();
	let plotDiv = $state<PlotlyDiv | undefined>(undefined);

	$effect(() => {
		if (plotDiv) renderPlot(plotDiv, data, layout);
	});
</script>

<div class="plot-wrapper">
	<div class="plot-div" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
