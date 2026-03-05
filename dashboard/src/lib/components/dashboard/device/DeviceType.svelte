<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { cachedFunction } from '$lib/cache';
	import { type Candidate, matchCandidate } from '$lib/candidates';
	import { renderPlot, donutLayout, buildDonutData } from '$lib/plotly';

	const deviceCandidates: Candidate[] = [
		{ name: 'iPhone', regex: /iPhone/, matches: 0 },
		{ name: 'Android', regex: /Android/, matches: 0 },
		{ name: 'Samsung', regex: /Tizen\//, matches: 0 },
		{ name: 'Mac', regex: /Macintosh/, matches: 0 },
		{ name: 'Windows', regex: /Windows/, matches: 0 }
	];

	function getDevice(userAgent: string | null): string {
		return matchCandidate(userAgent, deviceCandidates);
	}

	const deviceGetter = cachedFunction(getDevice);

	let { uaIdCount, userAgents }: { uaIdCount: { [id: number]: number }; userAgents: UserAgents } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && uaIdCount) renderPlot(plotDiv, buildDonutData(uaIdCount, userAgents, deviceGetter, graphColors), donutLayout(411));
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
