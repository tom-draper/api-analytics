<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { type Filter } from '$lib/filter';

	let {
		filter = $bindable(),
		counts
	}: {
		filter: Filter;
		counts?: Record<string, number>;
	} = $props();

	const max = $derived(counts ? Math.max(...Object.values(counts), 0) : 0);
</script>

<div class="thin-scroll flex max-h-[200px] flex-col overflow-y-auto">
	{#if filter}
		{#each Object.keys(filter.referrers) as referrer}
			<div class="flex items-center border-b border-solid border-[var(--border)] text-[13px] last:border-b-0">
				<Checkbox
					bind:checked={filter.referrers[referrer]}
					label={referrer}
					color="var(--highlight)"
					proportion={counts && max > 0 ? (counts[referrer] ?? 0) / max : undefined}
				/>
			</div>
		{/each}
	{/if}
</div>
