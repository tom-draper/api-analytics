<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { type Filter } from '$lib/filter';

	let { filter = $bindable() }: { filter: Filter } = $props();
</script>

<div class="thin-scroll flex max-h-[200px] flex-col overflow-y-auto">
	{#if filter}
		{#each Object.keys(filter.hostnames) as hostname}
			<div class="flex items-center border-b border-solid border-[var(--border)] text-[13px] last:border-b-0">
				<Checkbox
					bind:checked={filter.hostnames[hostname]}
					label={hostname}
					color="var(--highlight)"
				/>
			</div>
		{/each}
	{/if}
</div>
