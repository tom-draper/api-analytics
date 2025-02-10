<script lang="ts">
	import Checkbox from '$components/explorer/navigation/Checkbox.svelte';
	import { type Filter } from '$lib/filter';

	export let filter: Filter;
</script>

<div class="flex flex-col">
	<div class="flex items-center border-b border-solid border-[#2e2e2e]">
		{#if filter}
			<Checkbox bind:checked={filter.status.success} label="Success" color="var(--highlight)" />
		{/if}
	</div>
	<div class="flex items-center border-b border-solid border-[#2e2e2e]">
		{#if filter}
			<Checkbox bind:checked={filter.status.redirect} label="Redirect" color="var(--blue)" />
		{/if}
	</div>
	<div class="flex items-center border-b border-solid border-[#2e2e2e]">
		{#if filter}
			<Checkbox bind:checked={filter.status.client} label="Client error" color="var(--yellow)" />
		{/if}
	</div>
	<div class="flex items-center">
		{#if filter}
			<Checkbox bind:checked={filter.status.server} label="Server error" color="var(--red)" />
		{/if}
	</div>
</div>
