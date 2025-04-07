<script lang="ts">
	import { ColumnIndex } from '$lib/consts';

	function getPlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 50,
			barmode: 'stack',
			yaxis: {
				title: { text: 'Requests' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
				visible: false
			},
			xaxis: {
				title: { text: 'Date' },
				fixedrange: true,
				visible: false
			},
			dragmode: false
		};
	}

	function bars(data: RequestsData) {
		const requestFreq: Record<number, number> = {};

		for (let i = 0; i < data.length; i++) {
			const date = new Date(data[i][ColumnIndex.CreatedAt]);
			date.setHours(0, 0, 0, 0);
			const time = date.getTime();

			if (requestFreq[time]) {
				requestFreq[time] += 1;
			} else {
				requestFreq[time] = 1;
			}
		}

		const requestFreqArr = Object.entries(requestFreq).sort((a, b) => a[0] - b[0]);
		const dates = requestFreqArr.map((value) => new Date(parseInt(value[0])));
		const requests = requestFreqArr.map((value) => value[1]);
		const requestsText = requests.map((count) => `${count} requests`);

		return [
			{
				x: dates,
				y: requests,
				text: requestsText,
				type: 'bar',
				marker: { color: '#3fcf8e' },
				hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: bars(data),
			layout: getPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
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
		Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	function refreshPlot(data: RequestsData) {
		Plotly.react(plotDiv, bars(data), getPlotLayout());
	}

	let plotDiv: HTMLDivElement;

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData;
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>

<style scoped>
	#plotDiv {
		overflow-x: auto;
		height: 50px;
	}
</style>
