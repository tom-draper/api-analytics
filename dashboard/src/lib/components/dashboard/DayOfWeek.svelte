<script lang="ts">
	import { setParam } from '$lib/params';

	// getDay() order: 0=Sun, 1=Mon, ..., 6=Sat
	// Display order: Mon–Sun
	const dayLabels = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
	const dayOrder = [1, 2, 3, 4, 5, 6, 0];

	let { weekdayBuckets, targetWeekday = $bindable<number | null>(null) }: {
		weekdayBuckets: number[];
		targetWeekday: number | null;
	} = $props();

	const bars = $derived.by(() => {
		let max = 0;
		for (let i = 0; i < weekdayBuckets.length; i++) {
			if (weekdayBuckets[i] > max) max = weekdayBuckets[i];
		}
		return dayOrder
			.map((dayIdx, i) => ({
				label: dayLabels[i],
				dayIdx,
				count: weekdayBuckets[dayIdx],
				height: max > 0 ? weekdayBuckets[dayIdx] / max : 0,
			}))
			.filter((bar) => bar.count > 0);
	});

	function selectDay(dayIdx: number) {
		if (targetWeekday === dayIdx) {
			targetWeekday = null;
			setParam('weekday', null);
		} else {
			targetWeekday = dayIdx;
			setParam('weekday', String(dayIdx));
		}
	}
</script>

<div class="card">
	<div class="card-title">Day of week</div>
	<div class="bars">
		{#each bars as bar}
			<div class="bar-container">
				<button
					aria-label={bar.label}
					class="bar"
					class:selected={targetWeekday === bar.dayIdx}
					title="{bar.label}: {bar.count.toLocaleString()} requests"
					onclick={() => selectDay(bar.dayIdx)}
				>
					<div class="bar-inner" style="height: {bar.height * 100}%"></div>
				</button>
				<div class="label">{bar.label}</div>
			</div>
		{/each}
	</div>
</div>

<style scoped>
	.card {
		width: 100%;
		margin-top: 2em;
		position: relative;
	}
	.bars {
		height: 160px;
		display: flex;
		padding: 1.5em 2em 1em 2em;
	}
	.bar-container {
		flex: 1;
		margin: 0 5px;
		display: flex;
		flex-direction: column;
	}
	.bar {
		position: relative;
		flex-grow: 1;
		border-radius: 3px 3px 0 0;
		cursor: pointer;
	}
	.bar:hover {
		background: var(--fade-down);
	}
	.selected {
		background: var(--fade-down);
	}
	.bar-inner {
		position: absolute;
		bottom: 0;
		width: 100%;
		background: var(--highlight);
		border-radius: var(--radius-sm);
	}
	.label {
		padding-top: 8px;
		font-size: 0.8em;
		color: var(--dim-text);
		text-align: center;
	}
</style>
