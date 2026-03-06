<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';

	function daysAgo(date: Date): number {
		return Math.floor((Date.now() - date.getTime()) / (24 * 60 * 60 * 1000));
	}

	function hoursAgoTime(time: number): number {
		return Math.floor((Date.now() - time) / (60 * 60 * 1000));
	}

	function daysAgoTime(time: number): number {
		return Math.floor((Date.now() - time) / (24 * 60 * 60 * 1000));
	}

	function getSuccessRateArr(buckets: ActivityBucket[], period: Period, firstRequestDate: Date | null): number[] {
		if (period === '24 hours') {
			const count = 288;
			const successArr = new Array(count).fill(0);
			for (const bucket of buckets) {
				const idx = Math.floor((Date.now() - bucket.date) / (5 * 60 * 1000));
				if (idx >= 0 && idx < count) {
					successArr[successArr.length - 1 - idx] = bucket.successRate;
				}
			}
			return successArr;
		} else if (period === 'week') {
			const count = 24 * 7;
			const successArr = new Array(count).fill(0);
			for (const bucket of buckets) {
				const idx = hoursAgoTime(bucket.date);
				if (idx >= 0 && idx < count) {
					successArr[successArr.length - 1 - idx] = bucket.successRate;
				}
			}
			return successArr;
		} else {
			let days = period === 'all time'
				? (firstRequestDate ? daysAgo(firstRequestDate) : 0)
				: (periodToDays(period) ?? 0);
			days = Math.min(days, 500);
			const successArr = new Array(days).fill(0);
			for (const bucket of buckets) {
				const idx = daysAgoTime(bucket.date);
				if (idx >= 0 && idx < days) {
					successArr[successArr.length - 1 - idx] = bucket.successRate;
				}
			}
			return successArr;
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
			{#each successRate as value}
				<div
					class="error level-{Math.floor(value * 10)}"
					title={value >= 0 ? `Success rate: ${(value * 100).toFixed(1)}%` : 'No requests'}
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
