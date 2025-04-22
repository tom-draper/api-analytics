<script lang="ts">
	import { cachedFunction } from '$lib/cache';
	import {
		type Candidate,
		maintainCandidates,
	} from '$lib/candidates';
	import { ColumnIndex, graphColors } from '$lib/consts';

	const osCandidates: Candidate[] = [
		{ name: 'Windows 3.11', regex: /Win16/, matches: 0 },
		{
			name: 'Windows 95',
			regex: /(Windows 95)|(Win95)|(Windows_95)/,
			matches: 0,
		},
		{ name: 'Windows 98', regex: /(Windows 98)|(Win98)/, matches: 0 },
		{
			name: 'Windows 2000',
			regex: /(Windows NT 5.0)|(Windows 2000)/,
			matches: 0,
		},
		{
			name: 'Windows XP',
			regex: /(Windows NT 5.1)|(Windows XP)/,
			matches: 0,
		},
		{ name: 'Windows Server 2003', regex: /(Windows NT 5.2)/, matches: 0 },
		{ name: 'Windows Vista', regex: /(Windows NT 6.0)/, matches: 0 },
		{ name: 'Windows 7', regex: /(Windows NT 6.1)/, matches: 0 },
		{ name: 'Windows 8', regex: /(Windows NT 6.2)/, matches: 0 },
		{ name: 'Windows 10/11', regex: /(Windows NT 10.0)/, matches: 0 },
		{
			name: 'Windows NT 4.0',
			regex: /(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/,
			matches: 0,
		},
		{ name: 'Windows ME', regex: /Windows ME/, matches: 0 },
		{ name: 'OpenBSD', regex: /OpenBSD/, matches: 0 },
		{ name: 'SunOS', regex: /SunOS/, matches: 0 },
		{ name: 'Android', regex: /Android/, matches: 0 },
		{ name: 'Linux', regex: /(Linux)|(X11)/, matches: 0 },
		{ name: 'MacOS', regex: /(Mac_PowerPC)|(Macintosh)/, matches: 0 },
		{ name: 'QNX', regex: /QNX/, matches: 0 },
		{ name: 'iOS', regex: /iPhone OS/, matches: 0 },
		{ name: 'BeOS', regex: /BeOS/, matches: 0 },
		{ name: 'OS/2', regex: /OS\/2/, matches: 0 },
		{
			name: 'Search Bot',
			regex: /(APIs-Google)|(AdsBot)|(nuhk)|(Googlebot)|(Storebot)|(Google-Site-Verification)|(Mediapartners)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/,
			matches: 0,
		},
	];

	function getOS(userAgent: string | null): string {
		if (!userAgent) {
			return 'Unknown';
		}

		for (let i = 0; i < osCandidates.length; i++) {
			const candidate = osCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
				// Ensure osCandidates remains sorted by matches desc for future hits
				maintainCandidates(i, osCandidates);
				return candidate.name;
			}
		}

		return 'Other';
	}

	function getPlotLayout() {
		const monthAgo = new Date();
		monthAgo.setDate(monthAgo.getDate() - 30);
		const tomorrow = new Date();
		tomorrow.setDate(tomorrow.getDate() + 1);
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

	function donut(data: RequestsData) {
		const osCount: ValueCount = {};
		const osGetter = cachedFunction(getOS);
		for (let i = 0; i < data.length; i++) {
			const userAgent = userAgents[data[i][ColumnIndex.UserAgent]] || '';
			const os = osGetter(userAgent);
			if (os in osCount) {
				osCount[os]++;
			} else {
				osCount[os] = 1;
			}
		}

		const dataPoints = Object.entries(osCount).sort((a, b) => b[1] - a[1]);

		const oss = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [os, count] of dataPoints) {
			oss[i] = os;
			counts[i] = count;
			i++;
		}

		return [
			{
				values: counts,
				labels: oss,
				type: 'pie',
				hole: 0.6,
				marker: {
					colors: graphColors,
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
