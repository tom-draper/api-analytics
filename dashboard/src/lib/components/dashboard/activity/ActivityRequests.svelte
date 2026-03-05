<script lang="ts">
	import { type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';
	import { renderPlot, activityLayout } from '$lib/plotly';

	function bars(buckets: ActivityBucket[]) {
		const dates = buckets.map((b) => new Date(b.date));
		const users = buckets.map((b) => b.userCount);
		const requests = buckets.map((b) => b.requestCount - b.userCount);
		const requestsText = buckets.map((b) => `${b.requestCount} requests`);
		const usersText = buckets.map(
			(b) => `${b.requestCount} requests from ${b.userCount} users`
		);

		return [
			{
				x: dates,
				y: users,
				text: usersText,
				textposition: 'none',
				type: 'bar',
				marker: { color: '#3fcf8e' },
				hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
				showlegend: false
			},
			{
				x: dates,
				y: requests,
				text: requestsText,
				textposition: 'none',
				type: 'bar',
				marker: { color: '#228458' },
				hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	let { activityBuckets, period }: { activityBuckets: ActivityBucket[]; period: Period } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && activityBuckets) {
			renderPlot(plotDiv, bars(activityBuckets), activityLayout(period, 'Requests', 'stack'));
		}
	});
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
