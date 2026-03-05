<script lang="ts">
	import { cachedFunction } from '$lib/cache';
	import { type Candidate, matchCandidate } from '$lib/candidates';
	import { renderPlot, donutLayout, donutData } from '$lib/plotly';

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
	const colors = ['#3FCF8E', '#E46161', '#EBEB81'];

	function donut(uaIdCount: { [id: number]: number }, userAgents: UserAgents) {
		const deviceCount: ValueCount = {};
		for (const [uaId, count] of Object.entries(uaIdCount)) {
			const userAgent = userAgents[uaId as unknown as number] || '';
			const device = deviceGetter(userAgent);
			deviceCount[device] = (deviceCount[device] ?? 0) + count;
		}

		const dataPoints = Object.entries(deviceCount).sort((a, b) => b[1] - a[1]);
		const devices = dataPoints.map(([d]) => d);
		const counts = dataPoints.map(([, n]) => n);

		return donutData(devices, counts, colors);
	}

	let { uaIdCount, userAgents }: { uaIdCount: { [id: number]: number }; userAgents: UserAgents } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && uaIdCount) renderPlot(plotDiv, donut(uaIdCount, userAgents), donutLayout(411));
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
