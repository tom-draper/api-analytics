<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { type Filter } from '$lib/filter';

	let {
		filter = $bindable(),
		counts
	}: {
		filter: Filter;
		counts?: { success: number; redirect: number; client: number; server: number };
	} = $props();

	const max = $derived(counts ? Math.max(counts.success, counts.redirect, counts.client, counts.server) : 0);
</script>

<div class="flex flex-col text-[13px]">
	<div class="flex items-center border-b border-solid border-[var(--border)]">
		{#if filter}
			<Checkbox bind:checked={filter.status.success} label="Success" color="var(--highlight)" proportion={counts && max > 0 ? counts.success / max : undefined} />
		{/if}
	</div>
	<div class="flex items-center border-b border-solid border-[var(--border)]">
		{#if filter}
			<Checkbox bind:checked={filter.status.redirect} label="Redirect" color="var(--blue)" proportion={counts && max > 0 ? counts.redirect / max : undefined} />
		{/if}
	</div>
	<div class="flex items-center border-b border-solid border-[var(--border)]">
		{#if filter}
			<Checkbox bind:checked={filter.status.client} label="Client error" color="var(--yellow)" proportion={counts && max > 0 ? counts.client / max : undefined} />
		{/if}
	</div>
	<div class="flex items-center">
		{#if filter}
			<Checkbox bind:checked={filter.status.server} label="Server error" color="var(--red)" proportion={counts && max > 0 ? counts.server / max : undefined} />
		{/if}
	</div>
</div>
