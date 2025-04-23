<script lang="ts">
	import { ColumnIndex, methodMap } from '$lib/consts';
	import { statusBad, statusError, statusSuccess } from '$lib/status';

	const pageSize = 18;
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
		if (pageNumber < Math.ceil(data.length / pageSize)) {
			pageNumber++;
		}
	}

	function getPage(data: RequestsData, pageNumber: number) {
		const totalRequests = data.length;

		// Calculate start and end indices for slicing in reverse
		const startIdx = totalRequests - pageNumber * pageSize;
		const endIdx = startIdx + pageSize;

		// Ensure we don't go out of bounds
		const page: Page = data.slice(Math.max(0, startIdx), Math.max(0, endIdx)).reverse();

		if (page.length < pageSize) {
			const length = page.length;
			for (let i = 0; i < pageSize - length; i++) {
				const row: Page[number] = [null, null, null, null, null, null, null, null, null, null];
				page.push(row);
			}
		}

		return page;
	}

	export let data: RequestsData;
</script>

<div class="min-h-[inherit] flex flex-col">
	<table class="w-full text-left text-[14px] text-[var(--dim-text)] flex flex-col flex-1">
		<thead>
			<tr class="text-[var(--faint-text)] flex w-full">
				<th class="flex-none w-8"></th>
				<th class="flex-1">Timestamp</th>
				<th class="flex-none w-24 text-left">Status</th>
				<th class="flex-1">Hostname</th>
				<th class="flex-1">Path</th>
				<th class="flex-none w-16 text-left">Method</th>
				<th class="flex-1">IP Address</th>
				<th class="flex-1">User ID</th>
				<th class="flex-1">Response Time (ms)</th>
			</tr>
		</thead>
		<tbody class="flex-1 flex flex-col">
			{#if page}
				{#each page as request, i}
					<tr
						class="text-[14px] flex w-full flex-1"
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
						class:bottom-row={i === page.length - 1}
					>
						<td class="!pr-0 flex-none w-8 flex items-center justify-start">
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
						<td class="text-[var(--faint-text)] flex-1 flex items-center justify-start"
							>{request[ColumnIndex.CreatedAt]
								? request[ColumnIndex.CreatedAt].toLocaleString()
								: null}</td
						>
						<td
							class="flex-none w-24 flex items-center justify-start"
							class:text-[var(--highlight)]={request[ColumnIndex.Status] &&
								statusSuccess(request[ColumnIndex.Status])}
							class:text-[var(--red)]={request[ColumnIndex.Status] &&
								statusError(request[ColumnIndex.Status])}
							class:text-[rgb(235,235,129)]={request[ColumnIndex.Status] &&
								statusBad(request[ColumnIndex.Status])}>{request[ColumnIndex.Status]}</td
						>
						<td class="text-[var(--faint-text)] flex-1 flex items-center justify-start">{request[ColumnIndex.Hostname]}</td>
						<td class="text-[var(--faint-text)] flex-1 flex items-center justify-start">{request[ColumnIndex.Path]}</td>
						<td class="flex-none w-16 flex items-center justify-start"
							>{request[ColumnIndex.Method] !== null
								? methodMap[request[ColumnIndex.Method]]
								: null}</td
						>
						<td class="flex-1 flex items-center justify-start">{request[ColumnIndex.IPAddress]}</td>
						<td class="flex-1 flex items-center justify-start">{request[ColumnIndex.UserID]}</td>
						<td class="flex-1 flex items-center justify-start">{request[ColumnIndex.ResponseTime]}</td>
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
	<div class="grid px-3 text-xs text-[var(--dim-text)]">
		<div class="ml-auto flex gap-2">
			{#if data && data.length}
				<div class="content-center px-1">
					Page {pageNumber} of {data ? Math.ceil(data.length / pageSize).toLocaleString() : 0}
				</div>
				<button
					class="px-1 py-2 hover:text-[#ededed] disabled:text-[var(--dim-text)]"
					onclick={prevPage}
					aria-label="Previous page"
					disabled={pageNumber === 1}
					><svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"
						/>
					</svg>
				</button>
				<button
					class="px-1 py-2 hover:text-[#ededed] disabled:text-[var(--dim-text)]"
					onclick={nextPage}
					aria-label="Next page"
					disabled={pageNumber === (data ? Math.ceil(data.length / pageSize) : 0)}
					><svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="size-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3"
						/>
					</svg>
				</button>
			{/if}
		</div>
	</div>
</div>

<style scoped>
	table {
		min-height: inherit;
		display: flex;
		flex-direction: column;
	}
	tr {
		border-top: 1px solid #2e2e2e;
		border-left: 1px solid transparent;
		border-right: 1px solid transparent;
		border-radius: 4px;
	}
	tbody {
		display: flex;
		flex-direction: column;
		flex: 1;
	}
	tbody tr {
		display: flex;
		flex: 1;
		align-items: center;
	}
	.bottom-row {
		border-bottom: 1px solid var(--background);
	}
	th {
		border: none;
		font-weight: 600;
		padding: 0.2em 1em;
		display: flex;
		align-items: center;
		justify-content: flex-start;
		text-align: left;
	}
	td {
		padding: 1px 1em;
		text-align: left;
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
</style>
