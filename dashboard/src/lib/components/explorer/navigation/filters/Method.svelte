<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { methodMap } from '$lib/consts';
	import { type Filter } from '$lib/filter';

	let {
		filter = $bindable(),
		counts
	}: {
		filter: Filter;
		counts?: Record<number, number>;
	} = $props();

	const max = $derived(counts ? Math.max(...Object.values(counts), 0) : 0);
</script>

<div class="flex flex-col">
	{#if filter}
		{#each Object.keys(filter.methods) as method}
			<div class="flex items-center border-b border-solid border-[var(--border)] text-[13px] last:border-b-0">
				<Checkbox
					bind:checked={filter.methods[method]}
					label={methodMap[parseInt(method)]}
					color="var(--highlight)"
					proportion={counts && max > 0 ? (counts[parseInt(method)] ?? 0) / max : undefined}
				/>
			</div>
		{/each}
	{/if}
</div>
