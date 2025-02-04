<script lang="ts">
	import Table from "./Table.svelte";

    export let data: DashboardData;
</script>

<div class="w-[100%] ml-[20em] text-[var(--faded-text)]">
    <Table {data} />
</div>