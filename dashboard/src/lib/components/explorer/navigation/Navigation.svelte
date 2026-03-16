<script lang="ts">
	import Lightning from '$components/Lightning.svelte';
	import { type Filter } from '$lib/filter';
	import Expandable from '../Expandable.svelte';
	import Status from './filters/Status.svelte';
	import Timespan from './filters/Timespan.svelte';
	import Method from './filters/Method.svelte';
	import Hostname from './filters/Hostname.svelte';
	import ResponseTime from './filters/ResponseTime.svelte';

	type FilterBounds = { timespan: [number, number]; rt: [number, number] };

	let {
		filter = $bindable(),
		filteredRequests,
		totalCount,
		filterBounds,
		resetFilter
	}: {
		filter: Filter;
		filteredRequests: RequestsData;
		totalCount: number;
		filterBounds: FilterBounds | null;
		resetFilter: () => void;
	} = $props();
</script>

<nav
	class="fixed flex h-full w-[20em] flex-col overflow-y-auto border-r border-[var(--border)] bg-[var(--light-background)]"
>
	<div class="flex-grow p-2">
		<div class="flex px-2 pb-4 pt-2">
			<h2 class="text-left text-[var(--faded-text)]">Filters</h2>
			{#if filter && filteredRequests && totalCount !== filteredRequests.length}
				<button
					class="ml-auto flex items-center rounded border border-[var(--border)] px-2 text-xs text-[var(--faint-text)] hover:text-[var(--faded-text)]"
					onclick={resetFilter}
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="mr-1 size-3"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
					<div>Reset</div>
				</button>
			{/if}
		</div>
		<Expandable title="Timespan" {filter}>
			<Timespan bind:filter timespanBounds={filterBounds?.timespan ?? null} />
		</Expandable>
		<Expandable title="Status" {filter}>
			<Status bind:filter />
		</Expandable>
		<Expandable title="Method" {filter}>
			<Method bind:filter />
		</Expandable>
		<Expandable title="Hostname" {filter}>
			<Hostname bind:filter />
		</Expandable>
		<Expandable title="Response Time" {filter}>
			<ResponseTime bind:filter rtBounds={filterBounds?.rt ?? null} />
		</Expandable>
	</div>

	<div class="my-10 grid place-items-center">
		<div class="h-[28px] text-[var(--highlight)]">
			<Lightning />
		</div>
	</div>
</nav>
