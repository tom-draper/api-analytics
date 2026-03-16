<script lang="ts">
	import Graph from './Graph.svelte';
	import Table from './Table.svelte';

	let { filteredRequests, totalCount }: { filteredRequests: RequestsData; totalCount: number } = $props();
</script>

<div class="ml-[20em] flex w-full min-w-0 flex-col text-[var(--faded-text)]">
	<div class="relative">
		{#if filteredRequests.length > 0}
			<Graph data={filteredRequests} />
		{:else}
			<div class="h-[50px]"></div>
		{/if}
		<div class="pointer-events-none absolute inset-0 flex items-start pt-[14px] px-4 text-[12px] text-[var(--dim-text)]">
			{#if filteredRequests && totalCount > 0}
				{#if totalCount !== filteredRequests.length}
					Showing {filteredRequests.length.toLocaleString()} of {totalCount.toLocaleString()} requests
				{:else}
					{totalCount.toLocaleString()} requests
				{/if}
			{/if}
		</div>
	</div>

	<div class="min-h-[70vh]">
		<Table data={filteredRequests} />
	</div>
</div>
