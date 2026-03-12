<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { type Filter } from '$lib/filter';

	let { filter = $bindable() }: { filter: Filter } = $props();
</script>

<div class="flex flex-col">
	{#if filter}
		{#each Object.keys(filter.hostnames) as hostname}
			<div class="flex items-center border-b border-solid border-[var(--border)] text-[14px]">
				<Checkbox
					bind:checked={filter.hostnames[hostname]}
					label={hostname}
					color="var(--highlight)"
				/>
			</div>
		{/each}
	{/if}
</div>
