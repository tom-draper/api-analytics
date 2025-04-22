<script lang="ts">
	import { ColumnIndex } from '$lib/consts';
	import { cachedFunction } from '$lib/cache';
	import {
		type Candidate,
		maintainCandidates,
	} from '$lib/candidates';

	const deviceCandidates: Candidate[] = [
		{ name: 'iPhone', regex: /iPhone/, matches: 0 },
		{ name: 'Android', regex: /Android/, matches: 0 },
		{ name: 'Samsung', regex: /Tizen\//, matches: 0 },
		{ name: 'Mac', regex: /Macintosh/, matches: 0 },
		{ name: 'Windows', regex: /Windows/, matches: 0 },
	];

	function getDevice(userAgent: string | null): string {
		if (!userAgent) {
			return 'Unknown';
		}

		for (let i = 0; i < deviceCandidates.length; i++) {
			const candidate = deviceCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
				// Ensure deviceCandidates remains sorted by matches desc for future hits
				maintainCandidates(i, deviceCandidates);
				return candidate.name;
			}
		}

		return 'Other';
	}

	function getPlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 30, l: 30, t: 30, b: 25, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 196,
			width: 411,
			yaxis: {
				title: { text: 'Requests' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
			},
			xaxis: {
				visible: false,
			},
			dragmode: false,
		};
	}

	const colors = [
		'#3FCF8E', // Green
		'#E46161', // Red
		'#EBEB81', // Yellow
	];

	function donut(data: RequestsData) {
		const deviceCount: ValueCount = {};
		const deviceGetter = cachedFunction(getDevice);
		for (let i = 0; i < data.length; i++) {
			const userAgent = userAgents[data[i][ColumnIndex.UserAgent]] || '';
			const device = deviceGetter(userAgent);
			if (device in deviceCount) {
				deviceCount[device]++;
			} else {
				deviceCount[device] = 1;
			}
		}

		const dataPoints = Object.entries(deviceCount).sort(
			(a, b) => b[1] - a[1],
		);

		const devices = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [browser, count] of dataPoints) {
			devices[i] = browser;
			counts[i] = count;
			i++;
		}

		return [
			{
				values: counts,
				labels: devices,
				type: 'pie',
				hole: 0.6,
				marker: {
					colors: colors,
				},
			},
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: donut(data),
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
			donut(data),
			getPlotLayout(),
		)
	}

	let plotDiv: HTMLDivElement;

	$: if (plotDiv && data) {
		generatePlot(data);
	}

	export let data: RequestsData, userAgents: { [id: string]: string };
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>

<style scoped>
	#plotDiv {
		padding-right: 20px;
		overflow-x: auto;
	}
</style>
