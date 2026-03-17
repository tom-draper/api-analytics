<script lang="ts">
	import { ColumnIndex, methodMap } from '$lib/consts';
	import { statusBad, statusError, statusRedirect, statusSuccess } from '$lib/status';

	let {
		request,
		userAgents,
		data,
		onclose
	}: {
		request: RequestsData[number];
		userAgents: Record<number, string>;
		data: RequestsData;
		onclose: () => void;
	} = $props();

	const status = $derived(request[ColumnIndex.Status] as number | null);
	const method = $derived(request[ColumnIndex.Method] as number | null);
	const path = $derived(request[ColumnIndex.Path] as string | null);
	const hostname = $derived(request[ColumnIndex.Hostname] as string | null);
	const ip = $derived(request[ColumnIndex.IPAddress] as string | null);
	const location = $derived(request[ColumnIndex.Location] as string | null);
	const userId = $derived(request[ColumnIndex.UserID] as string | null);
	const referrer = $derived(request[ColumnIndex.Referrer] as string | null);
	const createdAt = $derived(request[ColumnIndex.CreatedAt] as Date | null);
	const rt = $derived(request[ColumnIndex.ResponseTime] as number | null);
	const uaId = $derived(request[ColumnIndex.UserAgent] as number | null);
	const ua = $derived(uaId != null ? (userAgents[uaId] ?? null) : null);

	const statusColor = $derived(
		status == null ? 'var(--faint-text)' :
		statusSuccess(status) ? 'var(--highlight)' :
		statusRedirect(status) ? 'var(--redirect-color)' :
		statusBad(status) ? 'var(--yellow)' :
		statusError(status) ? 'var(--red)' : 'var(--faint-text)'
	);

	const rtStats = $derived.by(() => {
		if (rt == null) return null;
		let count = 0, total = 0, sum = 0;
		const rts: number[] = [];
		for (const row of data) {
			const v = row[ColumnIndex.ResponseTime] as number | null;
			if (v == null) continue;
			total++;
			sum += v;
			rts.push(v);
			if (v <= rt) count++;
		}
		if (total === 0) return null;
		rts.sort((a, b) => a - b);
		const percentile = Math.round((count / total) * 100);
		const median = rts[Math.floor(rts.length * 0.5)];
		const p95 = rts[Math.floor(rts.length * 0.95)];
		const p99 = rts[Math.floor(rts.length * 0.99)];
		const mean = sum / total;
		const diffFromMean = rt - mean;

		const N = 40;
		const min = rts[0];
		const max = rts[rts.length - 1];
		const range = max - min;
		const bucketCounts = new Array<number>(N).fill(0);
		for (const v of rts) {
			bucketCounts[Math.min(N - 1, range === 0 ? 0 : Math.floor(((v - min) / range) * N))]++;
		}
		const activeBucket = range === 0 ? 0 : Math.min(N - 1, Math.floor(((rt - min) / range) * N));
		const buckets = bucketCounts.map((c, i) => ({ count: c, active: i === activeBucket }));
		const maxCount = Math.max(...bucketCounts, 1);

		return { percentile, median, p95, p99, mean, diffFromMean, buckets, maxCount, min, max };
	});

	const percentileColor = $derived(
		!rtStats ? 'var(--highlight)' :
		rtStats.percentile >= 95 ? 'var(--red)' :
		rtStats.percentile >= 75 ? 'var(--yellow)' :
		'var(--highlight)'
	);

	function ordinal(n: number): string {
		const v = n % 100;
		const s = n % 10;
		if (v >= 11 && v <= 13) return `${n}th`;
		if (s === 1) return `${n}st`;
		if (s === 2) return `${n}nd`;
		if (s === 3) return `${n}rd`;
		return `${n}th`;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<aside class="thin-scroll flex w-[300px] flex-none flex-col overflow-y-auto border-l border-[var(--border)] bg-[var(--light-background)]">

	<div class="flex flex-none items-center justify-between border-b border-[var(--border)] px-3 py-2">
		<span class="text-[13px] font-semibold text-[var(--faded-text)]">Request detail</span>
		<button class="cursor-pointer text-[var(--faint-text)] hover:text-[var(--faded-text)]" onclick={onclose}>
			<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4">
				<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
			</svg>
		</button>
	</div>

	<div class="flex flex-col divide-y divide-[var(--border)] text-[13px]">

		<!-- Overview -->
		<div class="flex flex-col gap-1 px-3 py-3">
			<div class="flex items-center gap-2">
				{#if status != null}
					<span class="font-semibold" style="color: {statusColor}">{status}</span>
				{/if}
				{#if method != null}
					<span class="text-[var(--faded-text)]">{methodMap[method]}</span>
				{/if}
			</div>
			{#if createdAt}
				<span class="text-[var(--faint-text)]">{createdAt.toLocaleString()}</span>
			{/if}
			{#if path}
				<span class="break-all font-mono text-[12px] text-[var(--faint-text)] pt-1">{path}</span>
			{/if}
		</div>

		<!-- Network -->
		<div class="flex flex-col px-3 py-3">
			<div class="section-label">Network</div>
			<div class="flex flex-col gap-1.5">
				{#if hostname}
					<div class="kv-row">
						<span class="kv-key">Host</span>
						<span class="kv-val break-all">{hostname}</span>
					</div>
				{/if}
				{#if ip}
					<div class="kv-row">
						<span class="kv-key">IP</span>
						<span class="kv-val">{ip}</span>
					</div>
				{/if}
				{#if location}
					<div class="kv-row">
						<span class="kv-key">Location</span>
						<span class="kv-val">{location}</span>
					</div>
				{/if}
				{#if referrer}
					<div class="kv-row">
						<span class="kv-key">Referrer</span>
						<span class="kv-val break-all">{referrer}</span>
					</div>
				{/if}
			</div>
		</div>

		<!-- Client -->
		{#if userId || ua}
			<div class="flex flex-col px-3 py-3">
				<div class="section-label">Client</div>
				<div class="flex flex-col gap-1.5">
					{#if userId}
						<div class="kv-row">
							<span class="kv-key">User ID</span>
							<span class="kv-val break-all font-mono text-[11px]">{userId}</span>
						</div>
					{/if}
					{#if ua}
						<div class="kv-row">
							<span class="kv-key">Agent</span>
							<span class="kv-val break-all text-[11px]">{ua}</span>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Performance -->
		{#if rt != null}
			<div class="flex flex-col px-3 py-3">
				<div class="section-label">Response time</div>
				<div class="flex flex-col gap-3">
					<div class="flex items-baseline gap-2">
						<span class="font-semibold text-[var(--faded-text)]">{rt}ms</span>
						{#if rtStats}
							<span class="text-[var(--faint-text)]">{ordinal(rtStats.percentile)} percentile</span>
						{/if}
					</div>

					{#if rtStats}
						<div class="flex flex-col gap-1">
							<div class="flex h-10 items-end gap-[2px]">
								{#each rtStats.buckets as bucket}
									<div
										class="min-w-0 flex-1 rounded-[1px]"
										style="height: {((bucket.count / rtStats.maxCount) * 100).toFixed(1)}%; background: {bucket.active ? percentileColor : 'rgba(var(--highlight-rgb), 0.12)'}"
									></div>
								{/each}
							</div>
							<div class="flex justify-between text-[11px] text-[var(--faint-text)]">
								<span>{Math.round(rtStats.min)}ms</span>
								<span>{Math.round(rtStats.max)}ms</span>
							</div>
						</div>

						<div class="flex flex-col gap-1.5 rounded border border-[var(--border)] px-3 py-2 text-[12px]">
							<div class="stat-row">
								<span>Mean</span>
								<span>{Math.round(rtStats.mean)}ms</span>
							</div>
							<div class="stat-row">
								<span>Median</span>
								<span>{Math.round(rtStats.median)}ms</span>
							</div>
							<div class="stat-row">
								<span>P95</span>
								<span>{Math.round(rtStats.p95)}ms</span>
							</div>
							<div class="stat-row">
								<span>P99</span>
								<span>{Math.round(rtStats.p99)}ms</span>
							</div>
							<div class="stat-row border-t border-[var(--border)] pt-1.5">
								<span>vs mean</span>
								<span
									class:text-[var(--highlight)]={rtStats.diffFromMean <= 0}
									class:text-[var(--red)]={rtStats.diffFromMean > 0}
								>{rtStats.diffFromMean > 0 ? '+' : ''}{Math.round(rtStats.diffFromMean)}ms</span>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{/if}

	</div>
</aside>

<style scoped>
	.section-label {
		font-size: 13px;
		font-weight: 500;
		color: var(--faint-text);
		margin-bottom: 8px;
	}
	.kv-row {
		display: flex;
		align-items: flex-start;
		gap: 8px;
	}
	.kv-key {
		width: 52px;
		flex-shrink: 0;
		color: var(--faint-text);
	}
	.kv-val {
		color: var(--faded-text);
	}
	.stat-row {
		display: flex;
		justify-content: space-between;
		color: var(--faint-text);
	}
	.stat-row span:last-child {
		color: var(--faded-text);
	}
</style>
