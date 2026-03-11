<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { bucketRange } from '$lib/plotly';

	type Cell = { value: number; date: number };

	function snapNow(period: Period): number {
		const d = new Date();
		if (period === '24 hours') {
			d.setSeconds(0, 0);
			d.setMinutes(Math.floor(d.getMinutes() / 5) * 5);
		} else if (period === 'week') {
			d.setSeconds(0, 0);
			d.setMinutes(0);
		} else {
			d.setHours(0, 0, 0, 0);
		}
		return d.getTime();
	}

	function getSuccessRateArr(buckets: ActivityBucket[], period: Period, firstRequestDate: Date | null): Cell[] {
		const bucketMap = new Map<number, number>(buckets.map((b) => [b.date, b.successRate]));
		const snap = snapNow(period);

		if (period === '24 hours') {
			const step = 5 * 60 * 1000;
			return Array.from({ length: 288 }, (_, i) => {
				const date = snap - (287 - i) * step;
				return { value: bucketMap.get(date) ?? 0, date };
			});
		} else if (period === 'week') {
			const step = 60 * 60 * 1000;
			return Array.from({ length: 168 }, (_, i) => {
				const date = snap - (167 - i) * step;
				return { value: bucketMap.get(date) ?? 0, date };
			});
		} else {
			let days = period === 'all time'
				? (firstRequestDate ? Math.floor((Date.now() - firstRequestDate.getTime()) / 86_400_000) : 0)
				: (periodToDays(period) ?? 0);
			days = Math.min(days, 500);
			const base = new Date();
			base.setHours(0, 0, 0, 0);
			const baseDay = base.getDate();
			return Array.from({ length: days }, (_, i) => {
				const d = new Date(base);
				d.setDate(baseDay - (days - 1 - i));
				const date = d.getTime();
				return { value: bucketMap.get(date) ?? 0, date };
			});
		}
	}

	let { activityBuckets, period, firstRequestDate }: {
		activityBuckets: ActivityBucket[];
		period: Period;
		firstRequestDate: Date | null;
	} = $props();

	const successRate = $derived(getSuccessRateArr(activityBuckets, period, firstRequestDate));
</script>

<div class="success-rate-container">
	{#if successRate != undefined}
		<div class="success-rate-title">Success rate</div>
		<div class="errors">
			{#each successRate as { value, date }}
				<div
					class="error level-{Math.floor(value * 10)}"
					title={value > 0 ? `${bucketRange(new Date(date), period)}\n${(value * 100).toFixed(1)}% success` : bucketRange(new Date(date), period)}
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
		background: #e46161;
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
		background: #ebeb81;
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
		background: #3fcf8e;
	}
</style>
