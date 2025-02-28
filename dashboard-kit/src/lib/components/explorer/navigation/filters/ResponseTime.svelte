<script lang="ts">
	import { ColumnIndex } from '$lib/consts';
	import { type Filter } from '$lib/filter';
	import RangeSlider from 'svelte-range-slider-pips';

	export let filter: Filter, data: DashboardData;

	let values: [number, number];

	let minResponseTime: number;
	let maxResponseTime: number;

	function responseTimeRange(data: DashboardData) {
		let min = Infinity;
		let max = 0;
		for (const row of data.requests) {
			const responseTime = row[ColumnIndex.ResponseTime];
			if (responseTime < min) {
				min = responseTime;
			}
			if (responseTime > max) {
				max = responseTime;
			}
		}

		return {min, max}
	}

	$: if (data) {
		const range = responseTimeRange(data)
		minResponseTime = range.min;
		maxResponseTime = range.max;

		values = [filter.responseTime[0], filter.responseTime[1]];
	}
</script>

{#if filter && minResponseTime && maxResponseTime}
	<div class="px-2 pb-3 pt-2">
		<RangeSlider
			min={minResponseTime}
			max={maxResponseTime}
			bind:values
			pips={false}
			first={true}
			pushy={true}
			on:stop={(e) => {
				filter.responseTime[e.detail.activeHandle] = e.detail.value;
			}}
		/>
		<div class="flex flex-col justify-between">
			<div class="grid place-items-center pb-1 text-center text-[13px]">
				<div class="flex w-full px-1">
					<div class="flex-grow text-left">
						{minResponseTime} ms
					</div>
					<div class="flex-grow text-right">
						{maxResponseTime} ms
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
