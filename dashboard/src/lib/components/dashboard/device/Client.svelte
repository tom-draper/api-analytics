<script lang="ts">
	import { cachedFunction } from '$lib/cache';
	import { matchCandidate } from '$lib/candidates';
	import { clientCandidates } from '$lib/device';
	import DonutChart from './DonutChart.svelte';

	const getter = cachedFunction((ua: string | null) => matchCandidate(ua, clientCandidates));

	let { uaIdCount, userAgents, targetClient = $bindable<string | null>(null) }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
		targetClient: string | null;
	} = $props();
</script>

<DonutChart {uaIdCount} {userAgents} {getter} paramKey="client" bind:target={targetClient} />
