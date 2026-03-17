<script lang="ts">
	import { type Filter } from '$lib/filter';
	import RangeSlider from './RangeSlider.svelte';
	import DistributionChart from './DistributionChart.svelte';

	let {
		filter = $bindable(),
		rtBounds,
		rtBuckets = []
	}: {
		filter: Filter;
		rtBounds: [number, number] | null;
		rtBuckets?: { center: number; count: number }[];
	} = $props();

	let values = $state<[number, number]>([0, 0]);

	$effect(() => {
		if (filter && rtBounds) {
			values = [
				filter.responseTime[0] === 0 ? rtBounds[0] : filter.responseTime[0],
				filter.responseTime[1] === Infinity ? rtBounds[1] : filter.responseTime[1]
			];
		}
	});
</script>

{#if filter && rtBounds}
	{#if rtBuckets.length > 0}
		<DistributionChart buckets={rtBuckets} lo={filter.responseTime[0]} hi={filter.responseTime[1]} />
	{/if}
	<RangeSlider
		min={rtBounds[0]}
		max={rtBounds[1]}
		bind:values
		onstop={(handle, value) => { filter.responseTime[handle] = value; }}
	/>
	<div class="flex items-center justify-center gap-1 px-3 pb-3 text-[13px]">
		<span class="text-[var(--faint-text)]">{Math.round(values[0])} ms</span>
		<span class="text-[var(--muted-text)]">–</span>
		<span class="text-[var(--faint-text)]">{Math.round(values[1])} ms</span>
	</div>
{/if}
