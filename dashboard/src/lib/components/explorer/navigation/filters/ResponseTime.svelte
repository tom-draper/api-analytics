<script lang="ts">
	import { type Filter } from '$lib/filter';
	import RangeSlider from 'svelte-range-slider-pips';

	let {
		filter = $bindable(),
		rtBounds
	}: {
		filter: Filter;
		rtBounds: [number, number] | null;
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

<style>
	:global(.rangeSlider) {
		font-size: 10px !important;
	}
	:global(.rangeHandle) {
		cursor: pointer;
	}
</style>

{#if filter && rtBounds}
	<div class="px-2 pb-3 pt-2">
		<RangeSlider
			min={rtBounds[0]}
			max={rtBounds[1]}
			bind:values
			pips={false}
			first={true}
			pushy={true}
			onstop={(e) => {
				const next: [number, number] = [filter.responseTime[0], filter.responseTime[1]];
				next[e.detail.activeHandle] = e.detail.value;
				filter.responseTime = next;
			}}
		/>
		<div class="flex flex-col justify-between">
			<div class="grid place-items-center pb-1 text-center text-[13px] text-[var(--faint-text)]">
				<div class="flex w-full px-1">
					<div class="flex-grow text-left">{values[0]} ms</div>
					<div class="flex-grow text-right">{values[1]} ms</div>
				</div>
			</div>
		</div>
	</div>
{/if}
