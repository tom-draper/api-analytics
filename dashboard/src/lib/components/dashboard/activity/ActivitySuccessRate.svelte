<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';
	import { ColumnIndex } from '$lib/consts';
	import { statusSuccessful } from '$lib/status';

	function daysAgo(date: Date): number {
		const now = new Date();
		// Calculate the difference in milliseconds
		const differenceInMilliseconds = now.getTime() - date.getTime();

		// Convert the difference to days
		const millisecondsPerDay = 24 * 60 * 60 * 1000;
		const differenceInDays = Math.floor(differenceInMilliseconds / millisecondsPerDay);

		return differenceInDays;
	}

	function daysAgoTime(time: number): number {
		const now = new Date();
		return Math.floor((now.getTime() - time) / (24 * 60 * 60 * 1000));
	}

	function hoursAgoTime(time: number): number {
		const now = new Date();
		return Math.floor((now.getTime() - time) / (60 * 60 * 1000));
	}

	type NumberSuccessCounter = Map<number, { total: number; successful: number }>;

	function getSuccessRate(data: RequestsData) {
		const success: NumberSuccessCounter = new Map();
		let minDate = new Date(8640000000000000);

		for (const row of data) {
			const date = new Date(row[ColumnIndex.CreatedAt]);
			if (period === '24 hours' || period === 'week') {
				// Hourly
				date.setMinutes(0, 0, 0);
			} else {
				date.setHours(0, 0, 0, 0);
			}

			const time = date.getTime();
			let entry = success.get(time);
			if (!entry) {
				entry = { total: 0, successful: 0 };
				success.set(time, entry);
			}

			entry.total++;
			if (statusSuccessful(row[ColumnIndex.Status])) {
				entry.successful++;
			}

			if (time < minDate.getTime()) {
				minDate = date;
			}
		}

		let successArr: number[];

		if (period === '24 hours' || period === 'week') {
			const hours = period === '24 hours' ? 24 : 24 * 7;
			successArr = new Array(hours).fill(0);

			for (const [time, entry] of success.entries()) {
				const idx = hoursAgoTime(time);
				if (idx >= 0 && idx < hours) {
					successArr[successArr.length - 1 - idx] = entry.successful / entry.total;
				}
			}
		} else {
			let days = period === 'all time' ? daysAgo(minDate) : periodToDays(period);
			if (days === null) {
				throw new Error('Invalid period');
			}
			days = Math.min(days, 500); // Limit to 500 days
			successArr = new Array(days).fill(0);

			for (const [time, entry] of success.entries()) {
				const idx = daysAgoTime(time);
				if (idx >= 0 && idx < days) {
					successArr[successArr.length - 1 - idx] = entry.successful / entry.total;
				}
			}
		}

		return successArr;
	}

	let successRate: number[];

	$: if (data) {
		successRate = getSuccessRate(data);
	}

	export let data: RequestsData, period: Period;
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
