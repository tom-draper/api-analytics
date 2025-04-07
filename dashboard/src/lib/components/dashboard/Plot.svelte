<script lang="ts">
	import { onMount } from 'svelte';
	import { plotly, loadPlotly } from '$lib/plotly';

	export let data = [];
	export let layout = {};
	export let config = {};

	let plotDiv: HTMLDivElement;

	onMount(async () => {
		await loadPlotly();
		$plotly && $plotly.newPlot(plotDiv, data, layout, config);

		// Create the plot
		// Plotly.newPlot(plotDiv, data, layout, config);

		return () => {
			// Cleanup when the component is destroyed
			$plotly.purge(plotDiv);
		};
	});

	$: {
		if ($plotly && plotDiv) {
			$plotly.react(plotDiv, data, layout, config);
		}
	}
</script>

<div bind:this={plotDiv} style="width: 100%; height: 100%;"></div>

<style>
	/* Optional: Style the container */
	div {
		display: block;
	}
</style>
