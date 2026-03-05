<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { renderPlot, donutLayout, donutData } from '$lib/plotly';

	let { versionCount, hasMultiple }: { versionCount: { [v: string]: number }; hasMultiple: boolean } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && hasMultiple) {
			const versions = Object.keys(versionCount);
			const counts = Object.values(versionCount);
			renderPlot(plotDiv, donutData(versions, counts, graphColors), donutLayout());
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
