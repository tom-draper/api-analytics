<script lang="ts">
	import { renderPlot, type PlotlyDiv } from '$lib/plotly';
	import { toggleParam } from '$lib/params';
	import { untrack } from 'svelte';

	function getPlotLayout() {
		return {
			font: { size: 12 },
			paper_bgcolor: 'transparent',
			height: 500,
			margin: { r: 50, l: 50, t: 20, b: 50, pad: 0 },
			polar: {
				bargap: 0,
				bgcolor: 'transparent',
				angularaxis: { direction: 'clockwise', showgrid: false },
				radialaxis: { gridcolor: '#303030' }
			}
		};
	}

	function bars(hourlyBuckets: number[], selectedHour: number | null) {
		// Shift to 12 onwards to make barpolar like clock face
		const dates = Array.from({ length: 24 }, (_, i) => i.toString() + ':00');
		const shiftedDates = dates.slice(12).concat(...dates.slice(0, 12));
		const shiftedBuckets = hourlyBuckets.slice(12).concat(...hourlyBuckets.slice(0, 12));

		const colors = shiftedDates.map((label) => {
			if (selectedHour === null) return '#3fcf8e';
			return parseInt(label) === selectedHour ? '#3fcf8e' : '#3fcf8e30';
		});

		return [
			{
				r: shiftedBuckets,
				theta: shiftedDates,
				marker: { color: colors },
				type: 'barpolar',
				hovertemplate: `<b>%{r}</b> requests at <b>%{theta}</b><extra></extra>`
			}
		];
	}

	function generatePlot(div: PlotlyDiv, hourlyBuckets: number[], selectedHour: number | null) {
		renderPlot(div, bars(hourlyBuckets, selectedHour), getPlotLayout());
	}

	function selectHour(hour: number) {
		toggleParam(
			'hour',
			hour,
			untrack(() => targetHour),
			(v) => (targetHour = v)
		);
	}

	let {
		hourlyBuckets,
		targetHour = $bindable<number | null>(null)
	}: {
		hourlyBuckets: number[];
		targetHour: number | null;
	} = $props();
	let plotDiv = $state<PlotlyDiv | undefined>(undefined);

	// Re-renders on both hourlyBuckets and targetHour changes (targetHour drives
	// the highlighted slice). Re-attaching the click handler is idempotent thanks
	// to removeAllListeners, so a single effect covers both cases.
	$effect(() => {
		if (!plotDiv || !hourlyBuckets) return;

		generatePlot(plotDiv, hourlyBuckets, targetHour);

		plotDiv.removeAllListeners?.('plotly_click');
		plotDiv.on?.('plotly_click', (data: any) => {
			const theta = data.points[0]?.theta as string;
			if (theta !== undefined) selectHour(parseInt(theta));
		});
	});
</script>

<div class="card">
	<div class="card-title">Usage time</div>
	<div class="plot-wrapper">
		<div class="plot-div" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</div>

<style scoped>
	.card {
		width: 100%;
		margin: 0;
	}
</style>
