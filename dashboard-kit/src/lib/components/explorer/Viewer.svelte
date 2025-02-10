<script lang="ts">
	import Graph from './Graph.svelte';
	import Search from './Search.svelte';
	import Table from './Table.svelte';

	export let data: DashboardData, filteredData: RequestsData;
</script>

<div class="ml-[20em] w-[100%] text-[var(--faded-text)]">
	<Search />

	<div class="mx-4 h-[16px] text-left text-xs text-[var(--dim-text)]">
		{#if data && filteredData && data.requests.length !== filteredData.length}
			Showing {filteredData.length.toLocaleString()} out of {data.requests.length.toLocaleString()} rows
		{/if}
	</div>

	<div>
		{#if data}
			<Graph data={filteredData} />
		{/if}
	</div>

	<div class="min-h-[70vh]">
		<Table data={filteredData} />
	</div>
</div>
