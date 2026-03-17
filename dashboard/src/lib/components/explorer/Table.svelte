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

	type SortDir = 'asc' | 'desc';
	let sortCol = $state<number | null>(null); // null = default newest-first
	let sortDir = $state<SortDir>('asc');

	function toggleSort(col: number) {
		if (sortCol === col) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortCol = col;
			sortDir = 'asc';
		}
		pageNumber = 1;
	}

	const sortedData = $derived.by((): RequestsData => {
		if (!data || sortCol === null) return data;
		const col = sortCol;
		const copy = data.slice();
		copy.sort((a, b) => {
			const av = a[col], bv = b[col];
			if (av == null && bv == null) return 0;
			if (av == null) return 1;
			if (bv == null) return -1;
			let cmp: number;
			if (col === ColumnIndex.Method) {
				cmp = (methodMap[av as number] ?? '').localeCompare(methodMap[bv as number] ?? '');
			} else if (av instanceof Date) {
				cmp = (av as Date).getTime() - (bv as Date).getTime();
			} else if (typeof av === 'number') {
				cmp = (av as number) - (bv as number);
			} else {
				cmp = String(av).localeCompare(String(bv));
			}
			return sortDir === 'asc' ? cmp : -cmp;
		});
		return copy;
	});

	let pageNumber = $state(1);
	const page = $derived(sortedData ? getPage(sortedData, pageNumber) : undefined);

	$effect(() => {
		data;
		pageNumber = 1;
	});

	function prevPage() {
		if (pageNumber > 1) pageNumber--;
	}

	function nextPage() {
		if (pageNumber < Math.ceil(sortedData.length / pageSize)) pageNumber++;
	}

	function getPage(data: RequestsData, pageNumber: number) {
		let page: Page;
		if (sortCol === null) {
			// Default: newest first via reverse-slice
			const total = data.length;
			const startIdx = total - pageNumber * pageSize;
			const endIdx = startIdx + pageSize;
			page = data.slice(Math.max(0, startIdx), Math.max(0, endIdx)).reverse();
		} else {
			// Custom sort: forward pagination
			const startIdx = (pageNumber - 1) * pageSize;
			page = data.slice(startIdx, startIdx + pageSize) as Page;
		}
		if (data.length > 0) {
			while (page.length < pageSize) {
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
				{#each [
					{ label: 'Timestamp',  col: ColumnIndex.CreatedAt,    cls: 'w-40 flex-none' },
					{ label: 'Status',     col: ColumnIndex.Status,        cls: 'w-14 flex-none' },
					{ label: 'Method',     col: ColumnIndex.Method,        cls: 'w-16 flex-none' },
					{ label: 'Hostname',   col: ColumnIndex.Hostname,      cls: 'min-w-0 flex-1' },
					{ label: 'Path',       col: ColumnIndex.Path,          cls: 'min-w-0 flex-[2]' },
					{ label: 'IP Address', col: ColumnIndex.IPAddress,     cls: 'w-28 flex-none' },
					{ label: 'User ID',    col: ColumnIndex.UserID,        cls: 'w-36 flex-none' },
					{ label: 'Time (ms)',  col: ColumnIndex.ResponseTime,  cls: 'w-24 flex-none' },
				] as col}
					<th
						class="{col.cls} cursor-pointer select-none text-left hover:text-[var(--faded-text)]"
						onclick={() => toggleSort(col.col)}
					>
						<span class="inline-flex items-center gap-1">
							{col.label}
							{#if sortCol === col.col}
								<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="size-2.5">
									{#if sortDir === 'asc'}
										<path stroke-linecap="round" stroke-linejoin="round" d="M4.5 10.5 12 3m0 0 7.5 7.5M12 3v18" />
									{:else}
										<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 13.5 12 21m0 0-7.5-7.5M12 21V3" />
									{/if}
								</svg>
							{/if}
						</span>
					</th>
				{/each}
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
						<td class="w-24 flex-none" style={rtColor(request[ColumnIndex.ResponseTime] as number | null)}>{request[ColumnIndex.ResponseTime] ?? ''}</td>
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>

	<div bind:this={paginationEl} class="flex items-center justify-end gap-2 px-3 py-1 text-[12px] text-[var(--dim-text)]">
		{#if data && data.length}
			<span class="px-1">Page {pageNumber} of {Math.ceil(sortedData.length / pageSize).toLocaleString()}</span>
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
				disabled={pageNumber === Math.ceil(sortedData.length / pageSize)}
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
		background: linear-gradient(90deg, rgba(var(--highlight-rgb), 0.09) 0%, rgba(var(--highlight-rgb), 0.03) 45%, transparent);
	}
	.redirect-bg {
		background: linear-gradient(90deg, rgba(var(--redirect-color-rgb), 0.14) 0%, rgba(var(--redirect-color-rgb), 0.04) 45%, transparent);
	}
	.warn-bg {
		background: linear-gradient(90deg, rgba(var(--yellow-rgb), 0.22) 0%, rgba(var(--yellow-rgb), 0.06) 45%, transparent);
	}
	.error-bg {
		background: linear-gradient(90deg, rgba(var(--red-rgb), 0.22) 0%, rgba(var(--red-rgb), 0.06) 45%, transparent);
	}
	.success-bg:hover {
		background: linear-gradient(90deg, rgba(var(--highlight-rgb), 0.18) 0%, rgba(var(--highlight-rgb), 0.06) 45%, transparent);
	}
	.redirect-bg:hover {
		background: linear-gradient(90deg, rgba(var(--redirect-color-rgb), 0.24) 0%, rgba(var(--redirect-color-rgb), 0.08) 45%, transparent);
	}
	.warn-bg:hover {
		background: linear-gradient(90deg, rgba(var(--yellow-rgb), 0.35) 0%, rgba(var(--yellow-rgb), 0.1) 45%, transparent);
	}
	.error-bg:hover {
		background: linear-gradient(90deg, rgba(var(--red-rgb), 0.35) 0%, rgba(var(--red-rgb), 0.1) 45%, transparent);
	}
</style>
