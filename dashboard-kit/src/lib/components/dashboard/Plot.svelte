<script lang="ts">
	import { onMount } from 'svelte';
	import Plotly from 'plotly.js-dist'; // Use 'plotly.js' if you're importing modular builds.

	export let data = [];
	export let layout = {};
	export let config = {};

	let plotDiv;

	onMount(() => {
		// Create the plot
		Plotly.newPlot(plotDiv, data, layout, config);

		return () => {
			// Cleanup when the component is destroyed
			Plotly.purge(plotDiv);
		};
	});
</script>

<div bind:this={plotDiv} style="width: 100%; height: 100%;"></div>

<style>
	/* Optional: Style the container */
	div {
		display: block;
	}
</style>
