<script lang="ts">
	import { renderPlot } from '$lib/plotly';

	function getPlotLayout(range: [number, number]) {
		return {
			title: false,
			autosize: true,
			margin: { r: 0, l: 0, t: 5, b: 0, pad: 10 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 50,
			yaxis: {
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
				visible: false,
			},
			xaxis: {
				range: range,
				showgrid: false,
				fixedrange: true,
				visible: false,
			},
			dragmode: false,
		};
	}

	function bars(freqTimes: number[], freqCounts: number[]) {
		return [
			{
				x: freqTimes,
				y: freqCounts,
				type: 'bar',
				marker: { color: '#505050' },
				hovertemplate: `<b>%{y} requests</b><br>%{x:.1f}ms</b> elapsed<extra></extra>`,
				showlegend: false,
			},
		];
	}

	function generatePlot(freqTimes: number[], freqCounts: number[]) {
		const range: [number, number] = [freqTimes[0] ?? 0, freqTimes[freqTimes.length - 1] ?? 0];
		renderPlot(plotDiv, bars(freqTimes, freqCounts), getPlotLayout(range));
	}

	let { freqTimes, freqCounts, LQ, median, UQ }: {
		freqTimes: number[];
		freqCounts: number[];
		LQ: number;
		median: number;
		UQ: number;
	} = $props();

	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (!plotDiv) return;
		if (freqTimes.length > 0) {
			generatePlot(freqTimes, freqCounts);
		} else if (plotDiv.data) {
			Plotly.purge(plotDiv);
		}
	});
</script>

<div class="card">
	<div class="card-title">
		Response times <span class="milliseconds">(ms)</span>
	</div>
	<div class="values">
		<div class="value lower-quartile">{LQ.toFixed(1)}</div>
		<div class="value median">{median.toFixed(1)}</div>
		<div class="value upper-quartile">{UQ.toFixed(1)}</div>
	</div>
	<div class="labels">
		<div class="label">LQ</div>
		<div class="label">Median</div>
		<div class="label">UQ</div>
	</div>
	<div class="distribution">
		<div id="plotly">
			<div id="plotDiv" bind:this={plotDiv}>
				<!-- Plotly chart will be drawn inside this DIV -->
			</div>
		</div>
	</div>
</div>

<style scoped>
	.card {
		overflow: hidden;
	}
	.values {
		display: flex;
		color: var(--highlight);
		font-size: 1.8em;
		font-weight: 700;
	}
	.values,
	.labels {
		margin: 0 0.5rem;
	}
	.value {
		flex: 1;
		font-size: 1.1em;
		padding: 20px 20px 4px;
	}
	.labels {
		display: flex;
		font-size: 0.8em;
		color: var(--dim-text);
	}
	.label {
		flex: 1;
	}

	#plotly {
		min-height: 50px;
	}

	.milliseconds {
		color: var(--dim-text);
		font-size: 0.8em;
		margin-left: 4px;
	}

	.median {
		font-size: 1em;
	}
	.upper-quartile,
	.lower-quartile {
		font-size: 1em;
		padding-bottom: 0;
	}

	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
			margin: 0 0 2em 0;
		}
	}
</style>
