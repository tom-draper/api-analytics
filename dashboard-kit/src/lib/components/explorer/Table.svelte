<script lang="ts">
	import { ColumnIndex, methodMap } from '$lib/consts';
	import { statusBad, statusError, statusSuccess } from '$lib/status';
	import { onMount } from 'svelte';

	onMount(() => {
		console.log(data);
	});

	export let data: DashboardData;
</script>

<table class="w-full text-left text-sm text-[var(--dim-text)]">
	<thead>
		<tr>
			<th></th>
			<th>Timestamp</th>
			<th>Status</th>
			<th>Hostname</th>
			<th>Path</th>
			<th>Method</th>
			<th>IP Address</th>
			<th>User ID</th>
			<th>Response Time (ms)</th>
		</tr>
	</thead>
	<tbody>
		{#if data}
			{#each data.requests.slice(0, 36) as request}
				<tr
					class:bg-[rgba(63,207,142,0.03)]={statusSuccess(request[ColumnIndex.Status])}
					class:bg-[rgba(228,97,97,0.12)]={statusError(request[ColumnIndex.Status])}
					class:bg-[rgba(235,235,129,0.12)]={statusBad(request[ColumnIndex.Status])}
				>
					<td>
						<div
							class="h-3 w-3 rounded"
							class:bg-[var(--highlight)]={statusSuccess(request[ColumnIndex.Status])}
							class:bg-[var(--red)]={statusError(request[ColumnIndex.Status])}
							class:bg-[rgb(235,235,129)]={statusBad(request[ColumnIndex.Status])}
						></div>
					</td>
					<td class="text-[var(--faint-text)]">{request[ColumnIndex.CreatedAt].toLocaleString()}</td
					>
					<td
						class:text-[var(--highlight)]={statusSuccess(request[ColumnIndex.Status])}
						class:text-[var(--red)]={statusError(request[ColumnIndex.Status])}
						class:text-[rgb(235,235,129)]={statusBad(request[ColumnIndex.Status])}
						>{request[ColumnIndex.Status]}</td
					>
					<td>{request[ColumnIndex.Hostname]}</td>
					<td>{request[ColumnIndex.Path]}</td>
					<td>{methodMap[request[ColumnIndex.Method]]}</td>
					<td>{request[ColumnIndex.IPAddress]}</td>
					<td>{request[ColumnIndex.UserID]}</td>
					<td>{request[ColumnIndex.ResponseTime]}</td>
				</tr>
			{/each}
		{/if}
	</tbody>
</table>

<style scoped>
	tr {
		border-top: 1px solid #2e2e2e;
		border-left: 1px solid transparent;
		border-right: 1px solid transparent;
		border-radius: 4px;
		cursor: pointer;
	}
    th {
        border: none;
    }
	tr:hover {
		border: 1px solid white !important;
	}
	th {
		padding: 0.2em 1em;
	}
	td {
		padding: 0 1em;
	}
</style>
