<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex, graphColors } from '../../../lib/consts';

	function getOS(userAgent: string): string {
		if (userAgent === null) {
			return 'Unknown';
		} else if (userAgent.match(/Win16/)) {
			return 'Windows 3.11';
		} else if (userAgent.match(/(Windows 95)|(Win95)|(Windows_95)/)) {
			return 'Windows 95';
		} else if (userAgent.match(/(Windows 98)|(Win98)/)) {
			return 'Windows 98';
		} else if (userAgent.match(/(Windows NT 5.0)|(Windows 2000)/)) {
			return 'Windows 2000';
		} else if (userAgent.match(/(Windows NT 5.1)|(Windows XP)/)) {
			return 'Windows XP';
		} else if (userAgent.match(/(Windows NT 5.2)/)) {
			return 'Windows Server 2003';
		} else if (userAgent.match(/(Windows NT 6.0)/)) {
			return 'Windows Vista';
		} else if (userAgent.match(/(Windows NT 6.1)/)) {
			return 'Windows 7';
		} else if (userAgent.match(/(Windows NT 6.2)/)) {
			return 'Windows 8';
		} else if (userAgent.match(/(Windows NT 10.0)/)) {
			return 'Windows 10/11';
		} else if (
			userAgent.match(/(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/)
		) {
			return 'Windows NT 4.0';
		} else if (userAgent.match(/Windows ME/)) {
			return 'Windows ME';
		} else if (userAgent.match(/OpenBSD/)) {
			return 'OpenBSE';
		} else if (userAgent.match(/SunOS/)) {
			return 'SunOS';
		} else if (userAgent.match(/Android/)) {
			return 'Android';
		} else if (userAgent.match(/(Linux)|(X11)/)) {
			return 'Linux';
		} else if (userAgent.match(/(Mac_PowerPC)|(Macintosh)/)) {
			return 'MacOS';
		} else if (userAgent.match(/QNX/)) {
			return 'QNX';
		} else if (userAgent.match(/iPhone OS/)) {
			return 'iOS';
		} else if (userAgent.match(/BeOS/)) {
			return 'BeOS';
		} else if (userAgent.match(/OS\/2/)) {
			return 'OS/2';
		} else if (
			userAgent.match(
				/(APIs-Google)|(AdsBot)|(nuhk)|(Googlebot)|(Storebot)|(Google-Site-Verification)|(Mediapartners)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/,
			)
		) {
			return 'Search Bot';
		} else {
			return 'Unknown';
		}
	}

	function osPlotLayout() {
		const monthAgo = new Date();
		monthAgo.setDate(monthAgo.getDate() - 30);
		const tomorrow = new Date();
		tomorrow.setDate(tomorrow.getDate() + 1);
		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 20, b: 20, pad: 0 },
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

	function pieChart() {
		const osCount: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const os = getOS(userAgent);
			if (os in osCount) {
				osCount[os]++
			} else {
				osCount[os] = 1
			}
		}

		const oss = new Array(Object.keys(osCount).length);
		const counts = new Array(Object.keys(osCount).length);
		let i = 0;
		for (const [os, count] of Object.entries(osCount)) {
			oss[i] = os;
			counts[i] = count;
			i++;
		}
		return [
			{
				values: counts,
				labels: oss,
				type: 'pie',
				marker: {
					colors: graphColors,
				},
			},
		];
	}

	function osPlotData() {
		return {
			data: pieChart(),
			layout: osPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot() {
		const plotData = osPlotData();
		//@ts-ignore
		new Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	let plotDiv: HTMLDivElement;
	let mounted = false;
	onMount(() => {
		mounted = true;
	});

	$: data && mounted && genPlot();

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
