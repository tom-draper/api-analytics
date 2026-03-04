<script lang="ts">
	function getPlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
			hovermode: false,
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 60,
			yaxis: {
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
				dragmode: false
			},
			xaxis: {
				visible: false,
				dragmode: false
			},
			dragmode: false
		};
	}

	function lines(buckets: number[]) {
		const n = buckets.length;
		return [
			{
				x: [...Array(n).keys()],
				y: buckets,
				type: 'lines',
				marker: { color: 'transparent' },
				showlegend: false,
				line: { shape: 'spline', smoothing: 1, color: '#3FCF8E30' },
				fill: 'tozeroy',
				fillcolor: '#3fcf8e15'
			}
		];
	}

	function generatePlot(buckets: number[]) {
		if (plotDiv.data) {
			Plotly.react(plotDiv, lines(buckets), getPlotLayout());
		} else {
			const plotData = {
				data: lines(buckets),
				layout: getPlotLayout(),
				config: { responsive: true, showSendToCloud: false, displayModeBar: false }
			};
			Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
		}
	}

	let { rate, buckets }: { rate: number | null; buckets: number[] } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && buckets) generatePlot(buckets);
	});
</script>

<div class="card">
	<div class="card-title">Success rate</div>
	<div
		class="value"
		class:red={rate !== null && rate <= 70}
		class:yellow={rate !== null && rate > 70 && rate < 90}
		class:green={rate === null || rate > 90}
	>
		{rate !== null ? `${rate.toFixed(1)}%` : 'N/A'}
	</div>
	<div id="plotly">
		<div id="plotDiv" bind:this={plotDiv}>
			<!-- Plotly chart will be drawn inside this DIV -->
		</div>
	</div>
</div>

<style scoped>
	.card {
		width: calc(215px - 1em);
		margin: 0 0 0 1em;
		position: relative;
		overflow: hidden;
	}
	.value {
		padding: 0.55em;
		text-align: center;
		font-size: 1.8em;
		font-weight: 700;
		color: var(--yellow);
		position: inherit;
		z-index: 2;
	}

	.red {
		color: var(--red);
	}
	.yellow {
		color: var(--yellow);
	}
	.green {
		color: var(--highlight);
	}
	#plotly {
		position: absolute;
		width: 110%;
		bottom: 0;
		overflow: hidden;
		margin: 0 -5%;
		z-index: 0;
	}
	@media screen and (max-width: 1030px) {
		.card {
			width: auto;
			flex: 1;
		}
	}
</style>
