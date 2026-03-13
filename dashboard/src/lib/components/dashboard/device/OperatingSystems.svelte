<script lang="ts">
	import { cachedFunction } from '$lib/cache';
	import { matchCandidate } from '$lib/candidates';
	import { osCandidates } from '$lib/device';
	import DonutChart from './DonutChart.svelte';

	const getter = cachedFunction((ua: string | null) => matchCandidate(ua, osCandidates));

	let { uaIdCount, userAgents, targetOS = $bindable<string | null>(null) }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
		targetOS: string | null;
	} = $props();
</script>

<DonutChart {uaIdCount} {userAgents} {getter} paramKey="os" bind:target={targetOS} />
