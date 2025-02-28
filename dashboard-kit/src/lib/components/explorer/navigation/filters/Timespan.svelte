<script lang="ts">
	import { ColumnIndex } from '$lib/consts';
	import { daysBetween, toDay } from '$lib/date';
	import { type Filter } from '$lib/filter';
	import RangeSlider from 'svelte-range-slider-pips';

	export let filter: Filter, data: DashboardData;

	let values: [number, number];

	let minDate: Date = new Date();
	let maxDate: Date = new Date();

	$: if (data) {
		minDate = data.requests[0][ColumnIndex.CreatedAt];
		maxDate = data.requests[data.requests.length - 1][ColumnIndex.CreatedAt];

		values = [filter.timespan[0], filter.timespan[1]];
	}
</script>

{#if filter && minDate && maxDate}
	<div class="px-2 pb-3 pt-2">
		<RangeSlider
			min={minDate.getTime()}
			max={maxDate.getTime()}
			bind:values
			pips={false}
			first={true}
			pushy={true}
			on:stop={(e) => {
				filter.timespan[e.detail.activeHandle] = e.detail.value;
			}}
		/>
		<div class="flex flex-col justify-between">
			<div class="grid place-items-center pb-1 text-center text-[13px]">
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
			<div class="text-center text-[13px]">
				{daysBetween(new Date(values[0]), new Date(values[1]))} days
			</div>
		</div>
	</div>
{/if}

<style>
	:root {
		--range-slider-color: var(--red);
		--range-slider: var(--highlight); /* slider main background color */
		--range-handle-inactive: var(--highlight);
		--range-handle: var(--highlight);
		--range-handle-focus: var(--highlight);
		--range-handle-border: var(--highlight);
		--range-range-inactive: var(--red); /* inactive range bar background color */
		--range-range: var(--highlight); /* active range background color */
		--range-float-inactive: var(--red); /* inactive floating label background color */
		--range-float: var(--range-handle-focus); /* floating label background color */
		--range-float-text: white; /* text color on floating label */
	}
	:global(.rangeHandle) {
		cursor: pointer;
	}
</style>
