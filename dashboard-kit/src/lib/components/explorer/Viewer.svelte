<script lang="ts">
	import Search from './Search.svelte';
	import Table from './Table.svelte';

	export let data: DashboardData;
</script>

<div class="ml-[20em] w-[100%] text-[var(--faded-text)]">
	<Search />

	<div class="mx-4 mb-2 text-left text-xs text-[var(--dim-text)]">
		0 out of {data ? data.requests.length : 0} rows filtered
	</div>

	<div class="min-h-[70vh]">
		<Table {data} />
	</div>
</div>
