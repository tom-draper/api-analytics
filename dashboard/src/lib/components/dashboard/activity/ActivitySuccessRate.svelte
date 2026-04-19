<script lang="ts">
	import { type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { bucketRange } from '$lib/plotly';

	type Cell = { value: number; date: number; requestCount: number };

	function getSuccessRateArr(buckets: ActivityBucket[]): Cell[] {
		return buckets.map((b) => ({ value: b.successRate, date: b.date, requestCount: b.requestCount }));
	}

	let { activityBuckets, period }: {
		activityBuckets: ActivityBucket[];
		period: Period;
	} = $props();

	const successRate = $derived(getSuccessRateArr(activityBuckets));
	</script>

	<div class="success-rate-container">
	{#if successRate != undefined}
		<div class="success-rate-title">Success rate</div>
		<div class="errors">
			{#each successRate as { value, date, requestCount }}
				<div
					class="error level-{requestCount === 0 ? 0 : Math.max(1, Math.min(9, Math.floor(value * 10)))}"
					title={requestCount > 0 ? `${bucketRange(new Date(date), period)}\n${(value * 100).toFixed(1)}% success` : bucketRange(new Date(date), period)}
				></div>
			{/each}
		</div>
	{/if}
	</div>

<style>
	.errors {
		display: flex;
		margin-left: 35px;
	}
	.error {
		background: var(--highlight);
		flex: 1;
		height: 40px;
		margin: 0 0.1%;
		border-radius: 1px;
	}
	.success-rate-container {
		text-align: left;
		font-size: 0.9em;
		color: var(--dim-text);
	}
	.success-rate-title {
		margin: 0 0 4px 43px;
	}
	.success-rate-container {
		margin: 1.5em 2.5em 2em;
	}
	.level-0 {
		background: #282828;
	}
	.level-1 {
		background: var(--red);
	}
	.level-2 {
		background: #f18359;
	}
	.level-3 {
		background: #f5a65a;
	}
	.level-4 {
		background: #f3c966;
	}
	.level-5 {
		background: var(--yellow);
	}
	.level-6 {
		background: #c7e57d;
	}
	.level-7 {
		background: #a1df7e;
	}
	.level-8 {
		background: #77d884;
	}
	.level-9 {
		background: var(--highlight);
	}
	.level-10 {
		background: var(--highlight);
	}
</style>
