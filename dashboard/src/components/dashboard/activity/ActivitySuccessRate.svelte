<script lang="ts">
	import { onMount } from 'svelte';
	import { periodToDays } from '../../../lib/period';
	import type { Period } from '../../../lib/settings';
	import { ColumnIndex } from '../../../lib/consts';

	function daysAgo(date: Date): number {
		const now = new Date();
		return Math.floor(
			(now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000),
		);
	}

	function daysAgoTime(time: number): number {
		const now = new Date();
		return Math.floor((now.getTime() - time) / (24 * 60 * 60 * 1000));
	}

	function hoursAgo(date: Date): number {
		const now = new Date();
		return Math.floor((now.getTime() - date.getTime()) / (60 * 60 * 1000));
	}

	function hoursAgoTime(time: number): number {
		const now = new Date();
		return Math.floor((now.getTime() - time) / (60 * 60 * 1000));
	}

	type NumberSuccessCounter = Map<
		number,
		{ total: number; successful: number }
	>;

	function setSuccessRate() {
		const success: NumberSuccessCounter = new Map();
		let minDate = new Date(8640000000000000);
		for (let i = 0; i < data.length; i++) {
			const date = new Date(data[i][ColumnIndex.CreatedAt]);
			if (period === '24 hours' || period === 'Week') {
				// Hourly
				date.setMinutes(0, 0, 0);
			} else {
				date.setHours(0, 0, 0, 0);
			}
			const time = date.getTime();
			if (!success.has(time)) {
				success.set(time, { total: 0, successful: 0 });
			}
			if (
				data[i][ColumnIndex.Status] >= 200 &&
				data[i][ColumnIndex.Status] <= 299
			) {
				success.get(time).successful++;
			}
			success.get(time).total++;
			if (date < minDate) {
				minDate = date;
			}
		}

		let successArr: number[];
		if (period === '24 hours' || period === 'Week') {
			// Hourly
			let hours: number;
			if (period === '24 hours') {
				hours = 24;
			} else {
				hours = 24 * 7;
			}

			successArr = new Array(hours).fill(-0.1); // -0.1 -> 0
			for (const time of success.keys()) {
				const idx = hoursAgoTime(time);
				successArr[successArr.length - 1 - idx] =
					success.get(time).successful / success.get(time).total;
			}
		} else {
			let days: number;
			if (period === 'All time') {
				days = daysAgo(minDate);
			} else {
				days = periodToDays(period);
			}

			successArr = new Array(days).fill(-0.1); // -0.1 -> 0
			for (const time of success.keys()) {
				const idx = daysAgoTime(time);
				successArr[successArr.length - 1 - idx] =
					success.get(time).successful / success.get(time).total;
			}
		}

		successRate = successArr;
	}

	function build() {
		setSuccessRate();
	}

	let successRate: number[];

	$: if (data) {
		build();
	}

	export let data: RequestsData, period: Period;
</script>

<div class="success-rate-container">
	{#if successRate != undefined}
		<div class="success-rate-title">Success rate</div>
		<div class="errors">
			{#each successRate as value}
				<div
					class="error level-{Math.floor(value * 10) + 1}"
					title={value >= 0
						? `Success rate: ${(value * 100).toFixed(1)}%`
						: 'No requests'}
				/>
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
		background: rgb(40, 40, 40);
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
