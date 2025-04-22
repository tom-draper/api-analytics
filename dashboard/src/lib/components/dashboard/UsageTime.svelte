<script lang="ts">
	import { ColumnIndex } from '$lib/consts';

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
				radialaxis: { gridcolor: '#303030' },
			},
		};
	}

	function bars(data: RequestsData) {
		const responseTimes = Array(24).fill(0);

		for (let i = 0; i < data.length; i++) {
			const date = data[i][ColumnIndex.CreatedAt];
			const time = date.getHours();
			responseTimes[time]++;
		}

		const requestFreqArr = Array.from({ length: 24 }, (_, i) => ({
			hour: i,
			responseTime: responseTimes[i],
		})).sort((a, b) => {
			return a.hour - b.hour;
		});

		let dates = new Array(requestFreqArr.length);
		let requests = new Array(requestFreqArr.length);
		for (let i = 0; i < requestFreqArr.length; i++) {
			dates[i] = requestFreqArr[i].hour.toString() + ':00';
			requests[i] = requestFreqArr[i].responseTime;
		}

		// Shift to 12 onwards to make barpolar like clock face
		dates = dates.slice(12).concat(...dates.slice(0, 12));
		requests = requests.slice(12).concat(...requests.slice(0, 12));

		return [
			{
				r: requests,
				theta: dates,
				marker: { color: '#3fcf8e' },
				type: 'barpolar',
				hovertemplate: `<b>%{r}</b> requests at <b>%{theta}</b><extra></extra>`,
			},
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: bars(data),
			layout: getPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function generatePlot(data: RequestsData) {
		if (plotDiv.data) {
			refreshPlot(data);
		} else {
			newPlot(data);
		}
	}

	async function newPlot(data: RequestsData) {
		const plotData = getPlotData(data);
		Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	function refreshPlot(data: RequestsData) {
		Plotly.react(
			plotDiv,
			bars(data),
			getPlotLayout(),
		)
	}

	let plotDiv: HTMLDivElement;

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData;
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
