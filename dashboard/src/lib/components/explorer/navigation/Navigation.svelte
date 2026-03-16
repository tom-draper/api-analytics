<script lang="ts">
	import { type Filter } from '$lib/filter';
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
	class="fixed top-[52px] flex h-[calc(100vh-52px)] w-[20em] flex-col overflow-y-auto border-r border-[var(--border)] bg-[var(--light-background)]"
>
	<div class="flex-1 p-3">
		<div class="mb-4 flex items-center justify-between px-1 text-left">
			<span class="text-[13px] font-semibold text-[var(--faded-text)]">Filters</span>
			{#if filter && filteredRequests && totalCount !== filteredRequests.length}
				<button
					class="flex items-center gap-1 rounded border border-[var(--border)] px-2 py-0.5 text-[11px] text-[var(--faint-text)] hover:text-[var(--faded-text)]"
					onclick={resetFilter}
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-3"
					>
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
					Reset
				</button>
			{/if}
		</div>

		<div class="mb-4">
			<div class="section-label">Timespan</div>
			{#if filterBounds}
				<Timespan bind:filter timespanBounds={filterBounds.timespan} />
			{/if}
		</div>

		<div class="mb-4">
			<div class="section-label">Status</div>
			<div class="rounded border border-[var(--border)]">
				<Status bind:filter />
			</div>
		</div>

		{#if filter && Object.keys(filter.methods).length > 0}
			<div class="mb-4">
				<div class="section-label">Method</div>
				<div class="rounded border border-[var(--border)]">
					<Method bind:filter />
				</div>
			</div>
		{/if}

		{#if filter && Object.keys(filter.hostnames).length > 1}
			<div class="mb-4">
				<div class="section-label">Hostname</div>
				<div class="rounded border border-[var(--border)]">
					<Hostname bind:filter />
				</div>
			</div>
		{/if}

		<div class="mb-2">
			<div class="section-label">Response Time</div>
			{#if filterBounds}
				<ResponseTime bind:filter rtBounds={filterBounds.rt} />
			{/if}
		</div>
	</div>

</nav>

<style scoped>
	.section-label {
		padding: 0 4px;
		margin-bottom: 6px;
		font-size: 13px;
		font-weight: 500;
		color: var(--faint-text);
		text-align: left;
	}
</style>
