<script lang="ts">
	let {
		buckets,
		lo,
		hi
	}: {
		buckets: { center: number; count: number }[];
		lo: number;
		hi: number;
	} = $props();

	const maxCount = $derived(Math.max(...buckets.map((b) => b.count), 1));
</script>

<div class="flex h-8 items-end gap-[1px] px-2 pb-0.5">
	{#each buckets as bucket}
		{@const inRange = bucket.center >= lo && bucket.center <= hi}
		{@const h = ((bucket.count / maxCount) * 100).toFixed(1)}
		<div
			class="min-w-0 flex-1 rounded-[1px] transition-colors duration-150"
			style="height: {h}%; background: {inRange
				? 'rgba(var(--highlight-rgb), 0.55)'
				: 'rgba(var(--highlight-rgb), 0.1)'}"
		></div>
	{/each}
</div>
