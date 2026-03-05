<script lang="ts">
	import { type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { renderPlot, activityLayout } from '$lib/plotly';

	function bars(buckets: ActivityBucket[]) {
		return [
			{
				x: buckets.map((b) => new Date(b.date)),
				y: buckets.map((b) => b.avgResponseTime),
				type: 'bar',
				marker: { color: '#707070' },
				hovertemplate: `<b>%{y:.1f}ms average</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	let { activityBuckets, period }: { activityBuckets: ActivityBucket[]; period: Period } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv) {
			renderPlot(plotDiv, bars(activityBuckets), activityLayout(period, 'Response time (ms)'));
		}
	});
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
