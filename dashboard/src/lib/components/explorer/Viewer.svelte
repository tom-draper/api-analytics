<script lang="ts">
	import Graph from './Graph.svelte';
	import Search from './Search.svelte';
	import Table from './Table.svelte';

	let { filteredRequests, totalCount }: { filteredRequests: RequestsData; totalCount: number } = $props();
</script>

<div class="ml-[20em] w-[100%] text-[var(--faded-text)]">
	<Search />

	<div class="mx-4 h-[16px] text-left text-xs text-[var(--dim-text)]">
		{#if filteredRequests && totalCount && totalCount !== filteredRequests.length}
			Showing {filteredRequests.length.toLocaleString()} out of {totalCount.toLocaleString()} rows
		{/if}
	</div>

	<div>
		{#if filteredRequests.length > 0}
			<Graph data={filteredRequests} />
		{/if}
	</div>

	<div class="min-h-[70vh]">
		<Table data={filteredRequests} />
	</div>
</div>
