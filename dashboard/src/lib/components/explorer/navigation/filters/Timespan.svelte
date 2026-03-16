<script lang="ts">
	import { daysBetween, toDay } from '$lib/date';
	import { type Filter } from '$lib/filter';
	import RangeSlider from 'svelte-range-slider-pips';

	let {
		filter = $bindable(),
		timespanBounds
	}: {
		filter: Filter;
		timespanBounds: [number, number] | null;
	} = $props();

	let values = $state<[number, number]>([0, 0]);

	$effect(() => {
		if (filter && timespanBounds) {
			values = [filter.timespan[0], filter.timespan[1]];
		}
	});
</script>

{#if filter && timespanBounds}
	<div class="px-2 pb-1 pt-2">
		<RangeSlider
			min={timespanBounds[0]}
			max={timespanBounds[1]}
			bind:values
			pips={false}
			first={true}
			pushy={true}
			onstop={(e) => {
				filter.timespan[e.detail.activeHandle] = e.detail.value;
			}}
		/>
		<div class="flex flex-col justify-between">
			<div class="grid place-items-center pb-1 text-center text-[13px] text-[var(--faint-text)]">
				<div class="flex">
					<div class="flex-grow">
						{toDay(new Date(values[0])).toLocaleDateString()}
					</div>
					<div class="px-2">to</div>
					<div class="flex-grow">
						{toDay(new Date(values[1])).toLocaleDateString()}
					</div>
				</div>
			</div>
			<div class="text-center text-[13px] text-[var(--dim-text)]">
				{daysBetween(new Date(values[0]), new Date(values[1]))} days
			</div>
		</div>
	</div>
{/if}

<style>
	:root {
		--range-slider-color: var(--red);
		--range-slider: var(--highlight);
		--range-handle-inactive: var(--highlight);
		--range-handle: var(--highlight);
		--range-handle-focus: var(--highlight);
		--range-handle-border: var(--highlight);
		--range-range-inactive: var(--red);
		--range-range: var(--highlight);
		--range-float-inactive: var(--red);
		--range-float: var(--range-handle-focus);
		--range-float-text: white;
	}
	:global(.rangeSlider) {
		font-size: 10px !important;
	}
	:global(.rangeHandle) {
		cursor: pointer;
	}
</style>
