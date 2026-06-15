<script lang="ts">
	import { type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { activityLayout, bucketRange } from '$lib/plotly';
	import PlotChart from '../PlotChart.svelte';

	function bars(buckets: ActivityBucket[], period: Period) {
		const dates = buckets.map((b) => new Date(b.date));
		return [
			{
				x: dates,
				y: buckets.map((b) => b.avgResponseTime),
				customdata: dates.map((d) => bucketRange(d, period)),
				type: 'bar',
				marker: { color: '#707070' },
				hovertemplate: `<b>%{y:.1f}ms average</b><br>%{customdata}<extra></extra>`,
				showlegend: false
			}
		];
	}

	let { activityBuckets, period }: { activityBuckets: ActivityBucket[]; period: Period } = $props();

	const data = $derived(bars(activityBuckets, period));
	const layout = $derived(activityLayout(period, 'Response time (ms)'));
</script>

<PlotChart {data} {layout} />
