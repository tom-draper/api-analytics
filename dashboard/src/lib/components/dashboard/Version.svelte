<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { renderPlot, donutLayout } from '$lib/plotly';
	import { setParam } from '$lib/params';
	import { untrack } from 'svelte';

	let { versionCount, hasMultiple, targetVersion = $bindable<string | null>(null) }: {
		versionCount: { [v: string]: number };
		hasMultiple: boolean;
		targetVersion: string | null;
	} = $props();

	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	function buildData(versions: string[], counts: number[]) {
		return [{
			values: counts,
			labels: versions,
			type: 'pie',
			hole: 0.6,
			marker: { colors: graphColors },
			pull: versions.map((v) => (targetVersion === v ? 0.08 : 0)),
		}];
	}

	function selectVersion(label: string) {
		const current = untrack(() => targetVersion);
		if (current === label) {
			targetVersion = null;
			setParam('version', null);
		} else {
			targetVersion = label;
			setParam('version', label);
		}
	}

	$effect(() => {
		if (!plotDiv || (!hasMultiple && targetVersion === null)) return;

		const versions = Object.keys(versionCount);
		const counts = Object.values(versionCount);
		renderPlot(plotDiv, buildData(versions, counts), donutLayout());
		window?.dispatchEvent(new Event('resize'));

		const el = plotDiv as any;
		el.removeAllListeners?.('plotly_click');
		el.removeAllListeners?.('plotly_legendclick');

		el.on?.('plotly_click', (data: any) => {
			const label = data.points[0]?.label;
			if (label) selectVersion(label);
		});

		el.on?.('plotly_legendclick', (data: any) => {
			const label = data.data?.[0]?.labels?.[data.expandedIndex];
			if (label) selectVersion(label);
			return false;
		});
	});
</script>

<div class="card flex-1" class:hidden={!hasMultiple && targetVersion === null}>
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
	@media screen and (max-width: 1070px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
