<script lang="ts">
	import { ColumnIndex, methodMap } from '$lib/consts';
	import { statusBad, statusError, statusSuccess } from '$lib/status';

	const pageSize = 32;
	let pageNumber = 1;
	let page: Page;

	type Nullable<T extends any[]> = { [K in keyof T]: T[K] | null };
	type Page = Nullable<RequestsData[number]>[];

	$: if (data) {
		page = getPage(data, pageNumber);
	}

	function prevPage() {
		if (pageNumber > 1) {
			pageNumber--;
		}
	}

	function nextPage() {
		if (pageNumber < Math.ceil(data.requests.length / pageSize)) {
			pageNumber++;
		}
	}

	function getPage(data: DashboardData, pageNumber: number) {
		const totalRequests = data.requests.length;

		// Calculate start and end indices for slicing in reverse
		const startIdx = totalRequests - pageNumber * pageSize;
		const endIdx = startIdx + pageSize;

		// Ensure we don't go out of bounds
		const page: Page = data.requests.slice(Math.max(0, startIdx), Math.max(0, endIdx)).reverse();

		if (page.length < pageSize) {
			const length = page.length;
			for (let i = 0; i < pageSize - length; i++) {
				const row: Page[number] = [null, null, null, null, null, null, null, null, null, null];
				page.push(row);
			}
		}

		return page;
	}

	export let data: DashboardData;
</script>

<div>
	<table class="w-full text-left text-[0.75em] text-[var(--dim-text)]">
		<thead>
			<tr class="text-[var(--faint-text)]">
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
				{#each page as request}
					<tr
						class="text-[1em]"
						class:success-bg={request[ColumnIndex.Status] &&
							statusSuccess(request[ColumnIndex.Status])}
						class:success-border={request[ColumnIndex.Status] &&
							statusSuccess(request[ColumnIndex.Status])}
						class:warn-bg={request[ColumnIndex.Status] && statusBad(request[ColumnIndex.Status])}
						class:warn-border={request[ColumnIndex.Status] &&
							statusBad(request[ColumnIndex.Status])}
						class:error-bg={request[ColumnIndex.Status] && statusError(request[ColumnIndex.Status])}
						class:error-border={request[ColumnIndex.Status] &&
							statusError(request[ColumnIndex.Status])}
					>
						<td class="!pr-0">
							<div class="grid place-items-center">
								<div
									class="h-2 w-2 rounded-sm"
									class:bg-[var(--highlight)]={request[ColumnIndex.Status] &&
										statusSuccess(request[ColumnIndex.Status])}
									class:bg-[var(--red)]={request[ColumnIndex.Status] &&
										statusError(request[ColumnIndex.Status])}
									class:bg-[rgb(235,235,129)]={request[ColumnIndex.Status] &&
										statusBad(request[ColumnIndex.Status])}
								></div>
							</div>
						</td>
						<td class="text-[var(--faint-text)]"
							>{request[ColumnIndex.CreatedAt]
								? request[ColumnIndex.CreatedAt].toLocaleString()
								: null}</td
						>
						<td
							class:text-[var(--highlight)]={request[ColumnIndex.Status] &&
								statusSuccess(request[ColumnIndex.Status])}
							class:text-[var(--red)]={request[ColumnIndex.Status] &&
								statusError(request[ColumnIndex.Status])}
							class:text-[rgb(235,235,129)]={request[ColumnIndex.Status] &&
								statusBad(request[ColumnIndex.Status])}>{request[ColumnIndex.Status]}</td
						>
						<td class="text-[var(--faint-text)]">{request[ColumnIndex.Hostname]}</td>
						<td class="text-[var(--faint-text)]">{request[ColumnIndex.Path]}</td>
						<td
							>{request[ColumnIndex.Method] !== null
								? methodMap[request[ColumnIndex.Method]]
								: null}</td
						>
						<td>{request[ColumnIndex.IPAddress]}</td>
						<td>{request[ColumnIndex.UserID]}</td>
						<td>{request[ColumnIndex.ResponseTime]}</td>
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
	<div class="grid px-2 text-xs text-[var(--dim-text)]">
		<div class="ml-auto flex gap-2">
			<button class="px-2 py-2 hover:text-[#ededed]" onclick={prevPage}>Prev</button>
			<button class="px-2 py-2 hover:text-[#ededed]" onclick={nextPage}>Next</button>
		</div>
	</div>
</div>

<style scoped>
	table {
		height: inherit;
	}
	tr {
		border-top: 1px solid #2e2e2e;
		border-left: 1px solid transparent;
		border-right: 1px solid transparent;
		border-radius: 4px;
		padding-left: 1em;
	}
	th {
		border: none;
		font-weight: 600;
	}
	.success-border:hover {
		border: 1px solid rgba(63, 207, 142, 0.5) !important;
	}
	.warn-border:hover {
		border: 1px solid rgba(235, 235, 129, 0.5) !important;
	}
	.error-border:hover {
		border: 1px solid rgba(228, 97, 97, 0.5) !important;
	}
	th {
		padding: 0.2em 1em;
	}
	td {
		padding: 1px 1em;
	}
	.success-bg,
	.warn-bg,
	.error-bg {
		cursor: pointer;
	}
	.success-bg {
		background: radial-gradient(rgba(63, 207, 142, 0.03), rgba(63, 207, 142, 0.05));
	}
	.warn-bg {
		background: radial-gradient(rgba(235, 235, 129, 0.14), rgba(235, 235, 129, 0.18));
	}
	.error-bg {
		background: radial-gradient(rgba(228, 97, 97, 0.14), rgba(228, 97, 97, 0.18));
	}
	.white-green {
		color: #bee7c5;
	}
	.white-red {
		color: #ffc1c1;
	}
	.white-yellow {
		color: #c0c0c0;
	}
</style>
