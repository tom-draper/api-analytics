<script lang="ts">
	import { ColumnIndex, methodMap } from '$lib/consts';
	import { statusBad, statusError, statusRedirect, statusSuccess } from '$lib/status';

	type Nullable<T extends any[]> = { [K in keyof T]: T[K] | null };
	type Page = Nullable<RequestsData[number]>[];

	let { data }: { data: RequestsData } = $props();

	const ROW_HEIGHT = 36;

	let containerEl = $state<HTMLDivElement | undefined>(undefined);
	let theadEl = $state<HTMLElement | undefined>(undefined);
	let paginationEl = $state<HTMLDivElement | undefined>(undefined);
	let containerHeight = $state(0);
	let theadHeight = $state(0);
	let paginationHeight = $state(0);

	const pageSize = $derived(
		containerHeight > 0
			? Math.max(1, Math.floor((containerHeight - theadHeight - paginationHeight) / ROW_HEIGHT))
			: 20
	);

	function observeHeight(el: HTMLElement, set: (h: number) => void) {
		const ro = new ResizeObserver(() => set(el.clientHeight));
		ro.observe(el);
		return () => ro.disconnect();
	}

	$effect(() => { if (containerEl) return observeHeight(containerEl, h => (containerHeight = h)); });
	$effect(() => { if (theadEl) return observeHeight(theadEl, h => (theadHeight = h)); });
	$effect(() => { if (paginationEl) return observeHeight(paginationEl, h => (paginationHeight = h)); });

	// Response time stats for proportional colouring
	const rtStats = $derived.by(() => {
		if (!data || data.length === 0) return { mean: 0, stddev: 1 };
		let sum = 0, count = 0;
		for (const row of data) {
			const rt = row[ColumnIndex.ResponseTime];
			if (rt != null) { sum += rt as number; count++; }
		}
		if (count === 0) return { mean: 0, stddev: 1 };
		const mean = sum / count;
		let variance = 0;
		for (const row of data) {
			const rt = row[ColumnIndex.ResponseTime];
			if (rt != null) variance += ((rt as number) - mean) ** 2;
		}
		return { mean, stddev: Math.sqrt(variance / count) || 1 };
	});

	// Interpolate dim-text (#707070) → red (#e46161) based on z-score (capped at 3σ)
	function rtColor(rt: number | null): string {
		if (rt == null) return '';
		const t = Math.min(1, Math.max(0, (rt - rtStats.mean) / (3 * rtStats.stddev)));
		if (t === 0) return '';
		const r = Math.round(112 + t * (228 - 112));
		const g = Math.round(112 + t * (97 - 112));
		const b = Math.round(112 + t * (97 - 112));
		return `color: rgb(${r}, ${g}, ${b})`;
	}

	let pageNumber = $state(1);
	const page = $derived(data ? getPage(data, pageNumber) : undefined);

	$effect(() => {
		data;
		pageNumber = 1;
	});

	function prevPage() {
		if (pageNumber > 1) pageNumber--;
	}

	function nextPage() {
		if (pageNumber < Math.ceil(data.length / pageSize)) pageNumber++;
	}

	function getPage(data: RequestsData, pageNumber: number) {
		const total = data.length;
		const startIdx = total - pageNumber * pageSize;
		const endIdx = startIdx + pageSize;
		const page: Page = data.slice(Math.max(0, startIdx), Math.max(0, endIdx)).reverse();
		if (page.length < pageSize) {
			const length = page.length;
			for (let i = 0; i < pageSize - length; i++) {
				page.push([null, null, null, null, null, null, null, null, null, null]);
			}
		}
		return page;
	}
</script>

<div bind:this={containerEl} class="flex h-full flex-col">
	<table class="w-full flex flex-col text-[13px] text-[var(--dim-text)]">
		<thead bind:this={theadEl}>
			<tr class="flex w-full text-[var(--faint-text)]">
				<th class="w-6 flex-none"></th>
				<th class="w-40 flex-none text-left">Timestamp</th>
				<th class="w-14 flex-none text-left">Status</th>
				<th class="w-16 flex-none text-left">Method</th>
				<th class="min-w-0 flex-1 text-left">Hostname</th>
				<th class="min-w-0 flex-[2] text-left">Path</th>
				<th class="w-28 flex-none text-left">IP Address</th>
				<th class="w-36 flex-none text-left">User ID</th>
				<th class="w-20 flex-none text-left">Time (ms)</th>
			</tr>
		</thead>
		<tbody class="flex flex-col">
			{#if page}
				{#each page as request, i}
					<tr
						class="flex w-full"
						class:success-bg={request[ColumnIndex.Status] && statusSuccess(request[ColumnIndex.Status])}
						class:success-border={request[ColumnIndex.Status] && statusSuccess(request[ColumnIndex.Status])}
						class:redirect-bg={request[ColumnIndex.Status] && statusRedirect(request[ColumnIndex.Status])}
						class:redirect-border={request[ColumnIndex.Status] && statusRedirect(request[ColumnIndex.Status])}
						class:warn-bg={request[ColumnIndex.Status] && statusBad(request[ColumnIndex.Status])}
						class:warn-border={request[ColumnIndex.Status] && statusBad(request[ColumnIndex.Status])}
						class:error-bg={request[ColumnIndex.Status] && statusError(request[ColumnIndex.Status])}
						class:error-border={request[ColumnIndex.Status] && statusError(request[ColumnIndex.Status])}
						class:bottom-row={i === page.length - 1}
					>
						<td class="w-6 flex-none flex items-center !px-0">
							<div
								class="mx-auto h-1.5 w-1.5 rounded-sm"
								class:bg-[var(--highlight)]={request[ColumnIndex.Status] && statusSuccess(request[ColumnIndex.Status])}
								class:bg-[var(--redirect-color)]={request[ColumnIndex.Status] && statusRedirect(request[ColumnIndex.Status])}
								class:bg-[var(--red)]={request[ColumnIndex.Status] && statusError(request[ColumnIndex.Status])}
								class:bg-[var(--yellow)]={request[ColumnIndex.Status] && statusBad(request[ColumnIndex.Status])}
							></div>
						</td>
						<td class="w-40 flex-none truncate text-[var(--faint-text)]">
							{request[ColumnIndex.CreatedAt] ? request[ColumnIndex.CreatedAt].toLocaleString() : ''}
						</td>
						<td
							class="w-14 flex-none"
							class:text-[var(--highlight)]={request[ColumnIndex.Status] && statusSuccess(request[ColumnIndex.Status])}
							class:text-[var(--redirect-color)]={request[ColumnIndex.Status] && statusRedirect(request[ColumnIndex.Status])}
							class:text-[var(--red)]={request[ColumnIndex.Status] && statusError(request[ColumnIndex.Status])}
							class:text-[var(--yellow)]={request[ColumnIndex.Status] && statusBad(request[ColumnIndex.Status])}
						>{request[ColumnIndex.Status]}</td>
						<td class="w-16 flex-none">{request[ColumnIndex.Method] !== null ? methodMap[request[ColumnIndex.Method]] : ''}</td>
						<td class="min-w-0 flex-1 truncate text-[var(--faint-text)]">{request[ColumnIndex.Hostname] ?? ''}</td>
						<td class="min-w-0 flex-[2] truncate text-[var(--faint-text)]">{request[ColumnIndex.Path] ?? ''}</td>
						<td class="w-28 flex-none truncate">{request[ColumnIndex.IPAddress] ?? ''}</td>
						<td class="w-36 flex-none truncate text-[var(--muted-text)]">{request[ColumnIndex.UserID] ?? ''}</td>
						<td class="w-20 flex-none" style={rtColor(request[ColumnIndex.ResponseTime] as number | null)}>{request[ColumnIndex.ResponseTime] ?? ''}</td>
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>

	<div bind:this={paginationEl} class="flex items-center justify-end gap-2 px-3 py-1 text-[12px] text-[var(--dim-text)]">
		{#if data && data.length}
			<span class="px-1">Page {pageNumber} of {Math.ceil(data.length / pageSize).toLocaleString()}</span>
			<button
				class="cursor-pointer p-1.5 hover:text-[var(--faded-text)] disabled:opacity-30 disabled:cursor-default"
				onclick={prevPage}
				aria-label="Previous page"
				disabled={pageNumber === 1}
			>
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
					<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
				</svg>
			</button>
			<button
				class="cursor-pointer p-1.5 hover:text-[var(--faded-text)] disabled:opacity-30 disabled:cursor-default"
				onclick={nextPage}
				aria-label="Next page"
				disabled={pageNumber === Math.ceil(data.length / pageSize)}
			>
				<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
					<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
				</svg>
			</button>
		{/if}
	</div>
</div>

<style scoped>
	table {
		display: flex;
		flex-direction: column;
	}
	tr {
		border-top: 1px solid var(--border);
		border-left: 1px solid transparent;
		border-right: 1px solid transparent;
		border-radius: var(--radius-md);
	}
	tbody tr {
		display: flex;
		height: 36px;
		overflow: hidden;
		align-items: center;
	}
	.bottom-row {
		border-bottom: 1px solid var(--background);
	}
	th {
		border: none;
		font-weight: 500;
		padding: 0.2em 0.6em;
		text-align: left;
	}
	td {
		padding: 0 0.6em;
		text-align: left;
	}
	.success-border:hover {
		outline: 1px solid rgba(var(--highlight-rgb), 0.5);
		outline-offset: -1px;
	}
	.warn-border:hover {
		outline: 1px solid rgba(var(--yellow-rgb), 0.5);
		outline-offset: -1px;
	}
	.error-border:hover {
		outline: 1px solid rgba(var(--red-rgb), 0.5);
		outline-offset: -1px;
	}
	.redirect-border:hover {
		outline: 1px solid rgba(var(--redirect-color-rgb), 0.5);
		outline-offset: -1px;
	}
	.success-bg,
	.redirect-bg,
	.warn-bg,
	.error-bg {
		cursor: pointer;
	}
	.success-bg {
		background: radial-gradient(rgba(var(--highlight-rgb), 0.03), rgba(var(--highlight-rgb), 0.05));
	}
	.redirect-bg {
		background: radial-gradient(rgba(var(--redirect-color-rgb), 0.06), rgba(var(--redirect-color-rgb), 0.1));
	}
	.warn-bg {
		background: radial-gradient(rgba(var(--yellow-rgb), 0.14), rgba(var(--yellow-rgb), 0.18));
	}
	.error-bg {
		background: radial-gradient(rgba(var(--red-rgb), 0.14), rgba(var(--red-rgb), 0.18));
	}
</style>
