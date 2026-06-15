<script lang="ts">
	import { type Period } from '$lib/period';
	import { countPerHour } from '$lib/utils';
	import MetricCard from './MetricCard.svelte';

	let { buckets, count, prevCount, firstDate, lastDate, period }: {
		buckets: number[];
		count: number;
		prevCount: number;
		firstDate: Date | null;
		lastDate: Date | null;
		period: Period;
	} = $props();

	const perHour = $derived(count <= 1 ? count : countPerHour(count, period, firstDate, lastDate));
</script>

<MetricCard title="Requests" {count} {prevCount} {buckets} {perHour} side="right" />
