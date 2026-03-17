<script lang="ts">
	import { daysBetween, toDay } from '$lib/date';
	import { type Filter } from '$lib/filter';
	import RangeSlider from './RangeSlider.svelte';
	import DistributionChart from './DistributionChart.svelte';

	let {
		filter = $bindable(),
		timespanBounds,
		timespanBuckets = []
	}: {
		filter: Filter;
		timespanBounds: [number, number] | null;
		timespanBuckets?: { center: number; count: number }[];
	} = $props();

	let values = $state<[number, number]>([0, 0]);

	$effect(() => {
		if (filter && timespanBounds) {
			values = [filter.timespan[0], filter.timespan[1]];
		}
	});
</script>

{#if filter && timespanBounds}
	{#if timespanBuckets.length > 0}
		<DistributionChart buckets={timespanBuckets} lo={filter.timespan[0]} hi={filter.timespan[1]} />
	{/if}
	<RangeSlider
		min={timespanBounds[0]}
		max={timespanBounds[1]}
		bind:values
		onstop={(handle, value) => { filter.timespan[handle] = value; }}
	/>
	<div class="flex flex-col justify-between px-3 pb-1">
		<div class="flex text-[13px] text-[var(--faint-text)]">
			<div class="flex-grow">{toDay(new Date(values[0])).toLocaleDateString()}</div>
			<div class="px-2">to</div>
			<div class="flex-grow text-right">{toDay(new Date(values[1])).toLocaleDateString()}</div>
		</div>
		<div class="text-center text-[13px] text-[var(--dim-text)]">
			{daysBetween(new Date(values[0]), new Date(values[1]))} days
		</div>
	</div>
{/if}
