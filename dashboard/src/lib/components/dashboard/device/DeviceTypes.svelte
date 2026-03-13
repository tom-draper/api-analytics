<script lang="ts">
	import { cachedFunction } from '$lib/cache';
	import { matchCandidate } from '$lib/candidates';
	import { deviceCandidates } from '$lib/device';
	import DonutChart from './DonutChart.svelte';

	const getter = cachedFunction((ua: string | null) => matchCandidate(ua, deviceCandidates));

	let { uaIdCount, userAgents, targetDeviceType = $bindable<string | null>(null) }: {
		uaIdCount: { [id: number]: number };
		userAgents: UserAgents;
		targetDeviceType: string | null;
	} = $props();
</script>

<DonutChart {uaIdCount} {userAgents} {getter} paramKey="deviceType" bind:target={targetDeviceType} />
