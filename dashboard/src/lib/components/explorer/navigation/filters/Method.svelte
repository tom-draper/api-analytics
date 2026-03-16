<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { methodMap } from '$lib/consts';
	import { type Filter } from '$lib/filter';

	let { filter = $bindable() }: { filter: Filter } = $props();
</script>

<div class="flex flex-col">
	{#if filter}
		{#each Object.keys(filter.methods) as method}
			<div class="flex items-center border-b border-solid border-[var(--border)] text-[14px] last:border-b-0">
				<Checkbox
					bind:checked={filter.methods[method]}
					label={methodMap[parseInt(method)]}
					color="var(--highlight)"
				/>
			</div>
		{/each}
	{/if}
</div>
