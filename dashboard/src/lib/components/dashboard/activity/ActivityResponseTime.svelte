<script lang="ts">
	import { type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { renderPlot, activityLayout, bucketRange } from '$lib/plotly';

	function bars(buckets: ActivityBucket[], period: Period) {
		const dates = buckets.map((b) => new Date(b.date));
		return [
			{
				x: dates,
				y: buckets.map((b) => b.avgResponseTime),
				customdata: dates.map((d) => bucketRange(d, period)),
				type: 'bar',
				marker: { color: 'var(--dim-text)' },
				hovertemplate: `<b>%{y:.1f}ms average</b><br>%{customdata}<extra></extra>`,
				showlegend: false
			}
		];
	}

	let { activityBuckets, period }: { activityBuckets: ActivityBucket[]; period: Period } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv) {
			renderPlot(plotDiv, bars(activityBuckets, period), activityLayout(period, 'Response time (ms)'));
		}
	});
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
