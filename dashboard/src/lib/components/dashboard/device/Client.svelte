<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { cachedFunction } from '$lib/cache';
	import { matchCandidate } from '$lib/candidates';
	import { clientCandidates } from '$lib/device';
	import { renderPlot, donutLayout, buildDonutData } from '$lib/plotly';
	import { setParam } from '$lib/params';
	import { untrack } from 'svelte';

	function getClient(userAgent: string | null): string {
		return matchCandidate(userAgent, clientCandidates);
	}

	const clientGetter = cachedFunction(getClient);

	function selectLabel(label: string) {
		const current = untrack(() => targetClient);
		if (current === label) {
			targetClient = null;
			setParam('client', null);
		} else {
			targetClient = label;
			setParam('client', label);
		}
	}

	let { uaIdCount, userAgents, targetClient = $bindable<string | null>(null) }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
		targetClient: string | null;
	} = $props();

	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (!plotDiv || !uaIdCount) return;

		renderPlot(plotDiv, buildDonutData(uaIdCount, userAgents, clientGetter, graphColors, targetClient), donutLayout(411));
		window?.dispatchEvent(new Event('resize'));

		const el = plotDiv as any;
		el.removeAllListeners?.('plotly_click');
		el.removeAllListeners?.('plotly_legendclick');

		el.on?.('plotly_click', (data: any) => {
			const label = data.points[0]?.label;
			if (label) selectLabel(label);
		});
		el.on?.('plotly_legendclick', (data: any) => {
			const label = data.data?.[0]?.labels?.[data.expandedIndex];
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
