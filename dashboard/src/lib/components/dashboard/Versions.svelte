<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { renderDonut, type PlotlyDiv } from '$lib/plotly';
	import { toggleParam } from '$lib/params';
	import { untrack } from 'svelte';

	let {
		versionCount,
		hasMultiple,
		targetVersion = $bindable<string | null>(null)
	}: {
		versionCount: { [v: string]: number };
		hasMultiple: boolean;
		targetVersion: string | null;
	} = $props();

	let plotDiv = $state<PlotlyDiv | undefined>(undefined);
	const colorMap = new Map<string, string>();

	function buildData(versions: string[], counts: number[]) {
		for (const v of versions) {
			if (!colorMap.has(v)) colorMap.set(v, graphColors[colorMap.size % graphColors.length]);
		}
		return [
			{
				values: counts,
				labels: versions,
				type: 'pie',
				hole: 0.6,
				marker: { colors: versions.map((v) => colorMap.get(v)!) },
				pull: versions.map((v) => (targetVersion === v ? 0.08 : 0))
			}
		];
	}

	function selectVersion(label: string) {
		toggleParam(
			'version',
			label,
			untrack(() => targetVersion),
			(v) => (targetVersion = v)
		);
	}

	$effect(() => {
		if (!plotDiv || (!hasMultiple && targetVersion === null)) return;

		const versions = Object.keys(versionCount);
		const counts = Object.values(versionCount);
		renderDonut(plotDiv, buildData(versions, counts), undefined, selectVersion);
	});
</script>

<div class="card flex-1" class:hidden={!hasMultiple && targetVersion === null}>
	<div class="card-title">Version</div>
	<div class="plot-wrapper">
		<div class="plot-div mr-[20px]" bind:this={plotDiv}>
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
	.plot-div {
		padding-right: 20px;
	}
	@media screen and (max-width: 1070px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
