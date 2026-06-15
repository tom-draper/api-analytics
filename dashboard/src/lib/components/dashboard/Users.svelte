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

	const perHour = $derived(count === 0 ? 0 : countPerHour(count, period, firstDate, lastDate));
</script>

<MetricCard
	title="Users"
	{count}
	{prevCount}
	{buckets}
	{perHour}
	side="left"
	tooltip="Based on IP address"
/>
