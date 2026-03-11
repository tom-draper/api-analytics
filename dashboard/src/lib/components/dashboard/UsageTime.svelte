<script lang="ts">
	import { renderPlot } from '$lib/plotly';
	import { setParam } from '$lib/params';
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

	function generatePlot(hourlyBuckets: number[], selectedHour: number | null) {
		renderPlot(plotDiv, bars(hourlyBuckets, selectedHour), getPlotLayout());
	}

	function selectHour(hour: number) {
		if (untrack(() => targetHour) === hour) {
			targetHour = null;
			setParam('hour', null);
		} else {
			targetHour = hour;
			setParam('hour', String(hour));
		}
	}

	let { hourlyBuckets, targetHour = $bindable<number | null>(null) }: {
		hourlyBuckets: number[];
		targetHour: number | null;
	} = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (!plotDiv || !hourlyBuckets) return;

		generatePlot(hourlyBuckets, untrack(() => targetHour));

		const el = plotDiv as any;
		el.removeAllListeners?.('plotly_click');
		el.on?.('plotly_click', (data: any) => {
			const theta = data.points[0]?.theta as string;
			if (theta !== undefined) selectHour(parseInt(theta));
		});
	});

	$effect(() => {
		const h = targetHour;
		const div = untrack(() => plotDiv);
		const buckets = untrack(() => hourlyBuckets);
		if (!div || !buckets) return;
		generatePlot(buckets, h);
	});
</script>

<div class="card">
	<div class="card-title">Usage time</div>
	<div id="plotly">
		<div id="plotDiv" bind:this={plotDiv}>
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
