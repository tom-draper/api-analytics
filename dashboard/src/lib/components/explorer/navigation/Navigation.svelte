<script lang="ts">
	import { type Filter } from '$lib/filter';
	import Status from './filters/Status.svelte';
	import Timespan from './filters/Timespan.svelte';
	import Method from './filters/Method.svelte';
	import Hostname from './filters/Hostname.svelte';
	import Location from './filters/Location.svelte';
	import Referrer from './filters/Referrer.svelte';
	import ResponseTime from './filters/ResponseTime.svelte';

	type Bucket = { center: number; count: number };
	type FilterBounds = { timespan: [number, number]; rt: [number, number]; timespanBuckets: Bucket[]; rtBuckets: Bucket[] };
	type FilterCounts = {
		status: { success: number; redirect: number; client: number; server: number };
		methods: Record<number, number>;
		hostnames: Record<string, number>;
		locations: Record<string, number>;
		referrers: Record<string, number>;
	};

	let {
		filter = $bindable(),
		filtersActive,
		filterBounds,
		counts,
		resetFilter
	}: {
		filter: Filter;
		filtersActive: boolean;
		filterBounds: FilterBounds | null;
		counts: FilterCounts | null;
		resetFilter: () => void;
	} = $props();
</script>

<nav
	class="thin-scroll fixed top-[52px] flex h-[calc(100vh-52px)] w-[20em] flex-col overflow-y-auto border-r border-[var(--border)] bg-[var(--light-background)]"
>
	<div class="flex-1 p-3">
		<div class="mb-4 flex items-center justify-between px-1 text-left">
			<span class="text-[13px] font-semibold text-[var(--faded-text)]">Filters</span>
			<button
				class="flex cursor-pointer items-center gap-1 rounded border px-2 py-0.5 text-[11px] transition-opacity"
				class:border-[var(--border)]={filtersActive}
				class:border-transparent={!filtersActive}
				class:text-[var(--faint-text)]={filtersActive}
				class:text-transparent={!filtersActive}
				class:pointer-events-none={!filtersActive}
				tabindex={filtersActive ? 0 : -1}
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
		</div>

		<div class="mb-4">
			<div class="section-label">Timespan</div>
			{#if filterBounds}
				<Timespan bind:filter timespanBounds={filterBounds.timespan} timespanBuckets={filterBounds.timespanBuckets} />
			{/if}
		</div>

		<div class="mb-4">
			<div class="section-label">Status</div>
			<div class="rounded border border-[var(--border)]">
				<Status bind:filter counts={counts?.status} />
			</div>
		</div>

		{#if filter && Object.keys(filter.methods).length > 0}
			<div class="mb-4">
				<div class="section-label">Method</div>
				<div class="rounded border border-[var(--border)]">
					<Method bind:filter counts={counts?.methods} />
				</div>
			</div>
		{/if}

		{#if filter && Object.keys(filter.hostnames).length > 1}
			<div class="mb-4">
				<div class="section-label">Hostname</div>
				<div class="rounded border border-[var(--border)]">
					<Hostname bind:filter counts={counts?.hostnames} />
				</div>
			</div>
		{/if}

		{#if filter && Object.keys(filter.locations).length > 0}
			<div class="mb-4">
				<div class="section-label">Location</div>
				<div class="rounded border border-[var(--border)]">
					<Location bind:filter counts={counts?.locations} />
				</div>
			</div>
		{/if}

		{#if filter && Object.keys(filter.referrers).length > 0}
			<div class="mb-4">
				<div class="section-label">Referrer</div>
				<div class="rounded border border-[var(--border)]">
					<Referrer bind:filter counts={counts?.referrers} />
				</div>
			</div>
		{/if}

		<div class="mb-2">
			<div class="section-label">Response Time</div>
			{#if filterBounds}
				<ResponseTime bind:filter rtBounds={filterBounds.rt} rtBuckets={filterBounds.rtBuckets} />
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
