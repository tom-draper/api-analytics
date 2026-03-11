<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { renderPlot, donutLayout, buildDonutData } from '$lib/plotly';
	import { setParam } from '$lib/params';
	import { untrack } from 'svelte';

	let { uaIdCount, userAgents, getter, paramKey, target = $bindable<string | null>(null) }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
		getter: (userAgent: string | null) => string;
		paramKey: string;
		target: string | null;
	} = $props();

	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	function selectLabel(label: string) {
		const current = untrack(() => target);
		if (current === label) {
			target = null;
			setParam(paramKey, null);
		} else {
			target = label;
			setParam(paramKey, label);
		}
	}

	$effect(() => {
		if (!plotDiv || !uaIdCount) return;

		renderPlot(plotDiv, buildDonutData(uaIdCount, userAgents, getter, graphColors, target), donutLayout(411));
		window?.dispatchEvent(new Event('resize'));

		const el = plotDiv as any;
		el.removeAllListeners?.('plotly_click');
		el.removeAllListeners?.('plotly_legendclick');

		el.on?.('plotly_click', (data: any) => {
			const label = data.points[0]?.label;
			if (label) selectLabel(label);
		});
		el.on?.('plotly_legendclick', (data: any) => {
			const label = data.node?.querySelector?.('.legendtext')?.textContent?.trim();
			if (label) selectLabel(label);
			return false;
		});
	});
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>

<style scoped>
	#plotDiv {
		padding-right: 20px;
		overflow-x: auto;
	}
</style>
