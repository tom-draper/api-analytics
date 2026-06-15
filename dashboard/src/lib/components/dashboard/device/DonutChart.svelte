<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { renderDonut, buildDonutData, type PlotlyDiv } from '$lib/plotly';
	import { toggleParam } from '$lib/params';
	import { untrack } from 'svelte';

	let { uaIdCount, userAgents, getter, paramKey, target = $bindable<string | null>(null) }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
		getter: (userAgent: string | null) => string;
		paramKey: string;
		target: string | null;
	} = $props();

	let plotDiv = $state<PlotlyDiv | undefined>(undefined);
	const colorMap = new Map<string, string>();

	function selectLabel(label: string) {
		toggleParam(paramKey, label, untrack(() => target), (v) => (target = v));
	}

	$effect(() => {
		if (!plotDiv || !uaIdCount) return;

		renderDonut(plotDiv, buildDonutData(uaIdCount, userAgents, getter, graphColors, target, colorMap), 411, selectLabel);
	});
</script>

<div class="plot-wrapper">
	<div class="plot-div" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>

<style scoped>
	.plot-div {
		padding-right: 20px;
		overflow-x: auto;
	}
</style>
