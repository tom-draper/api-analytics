<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';
	import type { ActivityBucket } from '$lib/aggregate';

	function getPlotLayout(period: Period) {
		const days = periodToDays(period);
		let periodAgo: Date | null = null;
		if (days !== null) {
			periodAgo = new Date();
			periodAgo.setDate(periodAgo.getDate() - days);
		}
		const now = new Date();

		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 20, b: 20, pad: 10 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 159,
			yaxis: {
				title: { text: 'Response time (ms)' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true
			},
			xaxis: {
				title: { text: 'Date' },
				showgrid: false,
				fixedrange: true,
				range: [periodAgo, now],
				visible: false
			},
			dragmode: false
		};
	}

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

	function generatePlot(buckets: ActivityBucket[], period: Period) {
		const b = bars(buckets);
		const layout = getPlotLayout(period);
		const config = { responsive: true, showSendToCloud: false, displayModeBar: false };
		if (plotDiv.data) {
			Plotly.react(plotDiv, b, layout);
		} else {
			Plotly.newPlot(plotDiv, b, layout, config);
		}
	}

	let { activityBuckets, period }: { activityBuckets: ActivityBucket[]; period: Period } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv) {
			generatePlot(activityBuckets, period);
		}
	});
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
