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

	function pieChart() {
		const clientCount: ValueCount = {};
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const client = getBrowser(userAgent);
			if (client in clientCount) {
				clientCount[client]++;
			} else {
				clientCount[client] = 1;
			}
		}

		const dataPoints = Object.entries(clientCount).sort(
			(a, b) => b[1] - a[1],
		);

		const clients = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [client, count] of dataPoints) {
			clients[i] = client;
			counts[i] = count;
			i++;
		}

		return {
			labels: clients,
			datasets: [
				{
					label: 'Device Type',
					data: counts,
					backgroundColor: graphColors,
					hoverOffset: 4,
				},
			],
		};
	}

	function genPlot() {
		const data = pieChart();

		let ctx = chartCanvas.getContext('2d');
		let chart = new Chart(ctx, {
			type: 'doughnut',
			data: data,
			options: {
				maintainAspectRatio: false,
				borderWidth: 0,
				plugins: {
					legend: {
						position: 'right',
						// reverse: true,
					},
				},
			},
		});
	}

	let chartCanvas: HTMLCanvasElement;
	let mounted = false;
	onMount(() => {
		mounted = true;
	});

	$: data && mounted && genPlot();
	export let data: RequestsData, getUserAgent: (id: number) => string;
</script>

<div id="plotly">
	<canvas bind:this={chartCanvas} id="chart"></canvas>
</div>

<style>
	#chart {
		height: 180px !important;
		width: 100% !important;
	}
</style>
