<script lang="ts">
	import { onMount } from 'svelte';
	import { graphColors } from '../../../lib/consts';
	import { ColumnIndex } from '../../../lib/consts';

	function getBrowser(userAgent: string): string {
		if (userAgent == null) {
			return 'Unknown';
		} else if (userAgent.match(/curl\//)) {
			return 'Curl';
		} else if (userAgent.match(/PostmanRuntime\//)) {
			return 'Postman';
		} else if (userAgent.match(/insomnia\//)) {
			return 'Insomnia';
		} else if (userAgent.match(/python-requests\//)) {
			return 'Python requests';
		} else if (userAgent.match(/node-fetch\//)) {
			return 'Nodejs fetch';
		} else if (userAgent.match(/Seamonkey\//)) {
			return 'Seamonkey';
		} else if (userAgent.match(/Firefox\//)) {
			return 'Firefox';
		} else if (userAgent.match(/Chrome\//)) {
			return 'Chrome';
		} else if (userAgent.match(/Chromium\//)) {
			return 'Chromium';
		} else if (userAgent.match(/aiohttp\//)) {
			return 'aiohttp';
		} else if (userAgent.match(/Python\//)) {
			return 'Python';
		} else if (userAgent.match(/[Gg]o-http-client\//)) {
			return 'Go http';
		} else if (userAgent.match(/Java\//)) {
			return 'axios';
		} else if (userAgent.match(/axios\//)) {
			return 'Dart';
		} else if (userAgent.match(/Dart\//)) {
			return 'okhttp';
		} else if (userAgent.match(/Uptime-Kuma\//)) {
			return 'Uptime Kuma';
		} else if (userAgent.match(/undici\//)) {
			return 'undici';
		} else if (userAgent.match(/Lush\//)) {
			return 'Lush';
		} else if (userAgent.match(/Zabbix/)) {
			return 'Zabbix';
		} else if (userAgent.match(/GuzzleHttp\//)) {
			return 'Guzzle';
		} else if (userAgent.match(/Better Uptime/)) {
			return 'Uptime';
		} else if (userAgent.match(/github-camo/)) {
			return 'GitHub Camo';
		} else if (userAgent.match(/Ruby/)) {
			return 'Ruby';
		} else if (userAgent.match(/node/)) {
			return 'node';
		} else if (userAgent.match(/Java\//)) {
			return 'Java';
		} else if (userAgent.match(/Next\.js/)) {
			return 'Next.js';
		} else if (userAgent.match(/Vercel Edge Functions/)) {
			return 'Vercel Edge Functions';
		} else if (userAgent.match(/OpenAI Image Downloader/)) {
			return 'OpenAI Image Downloader';
		} else if (userAgent.match(/OpenAI\//)) {
			return 'OpenAI';
		} else if (userAgent.match(/TsunamiSecurityScanner/)) {
			return 'Tsunami Security Scanner';
		} else if (userAgent.match(/iOS\//)) {
			return 'iOS';
		} else if (userAgent.match(/Safari\//)) {
			return 'Safari';
		} else if (userAgent.match(/Edg\//)) {
			return 'Edge';
		} else if (userAgent.match(/OPR\//) || userAgent.match(/Opera\//)) {
			return 'Opera';
		} else if (userAgent.match(/; MSIE /) || userAgent.match(/Trident\//)) {
			return 'Internet Explorer';
		} else {
			return 'Other';
		}
	}

	function clientPlotLayout() {
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
		const clientCount = {};
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const client = getBrowser(userAgent);
			clientCount[client] |= 0;
			clientCount[client]++;
		}

		const clients = [];
		const count = [];
		for (const browser in clientCount) {
			clients.push(browser);
			count.push(clientCount[browser]);
		}
		return [
			{
				values: count,
				labels: clients,
				type: 'pie',
				marker: {
					colors: graphColors,
				},
			},
		];
	}

	function browserPlotData() {
		return {
			data: pieChart(),
			layout: clientPlotLayout(),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
		};
	}

	function genPlot() {
		const plotData = browserPlotData();
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
