<script lang="ts">
	import { ColumnIndex } from '../../../lib/consts';
	import { cachedFunction } from '../../../lib/cache';
	import {
		type Candidate,
		maintainCandidates,
	} from '../../../lib/candidates';

	const deviceCandidates: Candidate[] = [
		{ name: 'iPhone', regex: /iPhone/, matches: 0 },
		{ name: 'Android', regex: /Android/, matches: 0 },
		{ name: 'Samsung', regex: /Tizen\//, matches: 0 },
		{ name: 'Mac', regex: /Macintosh/, matches: 0 },
		{ name: 'Windows', regex: /Windows/, matches: 0 },
	];

	function getDevice(userAgent: string | null): string {
		if (userAgent === null) {
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

	function devicePlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 10, b: 20, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 180,
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

	function pieChart(data: RequestsData) {
		const deviceCount: ValueCount = {};
		const deviceGetter = cachedFunction(getDevice);
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
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
				marker: {
					colors: colors,
				},
			},
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: pieChart(data),
			layout: devicePlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot(data: RequestsData) {
		const plotData = getPlotData(data);
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

	export let data: RequestsData, getUserAgent: (id: number) => string;
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>

<style>
	#plotDiv {
		margin-right: 20px;
	}
</style>
