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
		{#each Object.keys(filter.hostnames) as hostname}
			<div class="flex items-center border-b border-solid border-[var(--border)] text-[13px] last:border-b-0">
				<Checkbox
					bind:checked={filter.hostnames[hostname]}
					label={hostname}
					color="var(--highlight)"
					proportion={counts && max > 0 ? (counts[hostname] ?? 0) / max : undefined}
				/>
			</div>
		{/each}
	{/if}
</div>
