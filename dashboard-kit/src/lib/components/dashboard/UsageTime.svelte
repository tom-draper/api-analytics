<script lang="ts">
	import { ColumnIndex } from '$lib/consts';

	function getLayout() {
		return {
			font: { size: 12 },
			paper_bgcolor: 'transparent',
			height: 500,
			margin: { r: 35, l: 70, t: 20, b: 50, pad: 0 },
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
			// @ts-ignore
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

	function buildPlotData(data: RequestsData) {
		return {
			data: bars(data),
			layout: getLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot(data: RequestsData) {
		const plotData = buildPlotData(data);
		//@ts-ignore
		new Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	let plotDiv: HTMLDivElement;

	$: if (plotDiv && data) {
		genPlot(data);
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
