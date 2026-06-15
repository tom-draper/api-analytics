<script lang="ts">
	import { type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { activityLayout, bucketRange } from '$lib/plotly';
	import PlotChart from '../PlotChart.svelte';

	function bars(buckets: ActivityBucket[], period: Period) {
		const dates = buckets.map((b) => new Date(b.date));
		const timeRanges = dates.map((d) => bucketRange(d, period));
		const users = buckets.map((b) => b.userCount);
		const requests = buckets.map((b) => b.requestCount - b.userCount);
		const requestsText = buckets.map((b) => `${b.requestCount} requests`);
		const usersText = buckets.map((b) => `${b.requestCount} requests from ${b.userCount} users`);

		return [
			{
				x: dates,
				y: users,
				text: usersText,
				customdata: timeRanges,
				textposition: 'none',
				type: 'bar',
				marker: { color: '#3fcf8e' },
				hovertemplate: `<b>%{text}</b><br>%{customdata}<extra></extra>`,
				showlegend: false
			},
			{
				x: dates,
				y: requests,
				text: requestsText,
				customdata: timeRanges,
				textposition: 'none',
				type: 'bar',
				marker: { color: '#228458' },
				hovertemplate: `<b>%{text}</b><br>%{customdata}<extra></extra>`,
				showlegend: false
			}
		];
	}

	let { activityBuckets, period }: { activityBuckets: ActivityBucket[]; period: Period } = $props();

	const data = $derived(bars(activityBuckets, period));
	const layout = $derived(activityLayout(period, 'Requests', 'stack'));
</script>

<PlotChart {data} {layout} />
