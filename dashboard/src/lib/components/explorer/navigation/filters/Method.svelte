<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { methodMap } from '$lib/consts';
	import { type Filter } from '$lib/filter';

	export let filter: Filter;
</script>

<div class="flex flex-col">
	{#if filter}
		{#each Object.keys(filter.methods) as method}
			<div class="flex items-center border-b border-solid border-[#2e2e2e] text-[14px]">
				<Checkbox
					bind:checked={filter.methods[method]}
					label={methodMap[parseInt(method)]}
					color="var(--highlight)"
				/>
			</div>
		{/each}
	{/if}
</div>
