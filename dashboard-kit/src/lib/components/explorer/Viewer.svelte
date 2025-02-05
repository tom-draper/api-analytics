<script lang="ts">
	import Search from "./Search.svelte";
	import Table from "./Table.svelte";

    export let data: DashboardData;
</script>

<div class="w-[100%] ml-[20em] text-[var(--faded-text)]">
    <Search />

    <div class="text-xs text-[var(--dim-text)] text-left mb-2 mx-4">
        0 out of {data ? data.requests.length : 0} rows filtered
    </div>

    <div class="min-h-[70vh]">
        <Table {data} />
    </div>

</div>