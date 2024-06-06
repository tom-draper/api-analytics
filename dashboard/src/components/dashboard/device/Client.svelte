<script lang="ts">
	import { onMount } from 'svelte';
	import { graphColors, ColumnIndex } from '../../../lib/consts';
	import { cachedFunction } from '../../../lib/cache';
	import { Chart } from 'chart.js/auto';

	const clientCandidates = [
		{name: 'Curl', regex: /curl\//, matches: 0},
		{name: 'Postman', regex: /PostmanRuntime\//, matches: 0},
		{name: 'Insomnia', regex: /insomnia\//, matches: 0},
		{name: 'Python requests', regex: /python-requests\//, matches: 0},
		{name: 'Nodejs fetch', regex: /node-fetch\//, matches: 0},
		{name: 'Seamonkey', regex: /Seamonkey\//, matches: 0},
		{name: 'Firefox', regex: /Firefox\//, matches: 0},
		{name: 'Chrome', regex: /Chrome\//, matches: 0},
		{name: 'Chromium', regex: /Chromium\//, matches: 0},
		{name: 'aiohttp', regex: /aiohttp\//, matches: 0},
		{name: 'Python', regex: /Python\//, matches: 0},
		{name: 'Go http', regex: /[Gg]o-http-client\//, matches: 0},
		{name: 'Java', regex: /Java\//, matches: 0},
		{name: 'axios', regex: /axios\//, matches: 0},
		{name: 'Dart', regex: /Dart\//, matches: 0},
		{name: 'OkHttp', regex: /OkHttp\//, matches: 0},
		{name: 'Uptime Kuma', regex: /Uptime-Kuma\//, matches: 0},
		{name: 'undici', regex: /undici\//, matches: 0},
		{name: 'Lush', regex: /Lush\//, matches: 0},
		{name: 'Zabbix', regex: /Zabbix/, matches: 0},
		{name: 'Guzzle', regex: /GuzzleHttp\//, matches: 0},
		{name: 'Uptime', regex: /Better Uptime/, matches: 0},
		{name: 'GitHub Camo', regex: /github-camo/, matches: 0},
		{name: 'Ruby', regex: /Ruby/, matches: 0},
		{name: 'Node.js', regex: /node/, matches: 0},
		{name: 'Next.js', regex: /Next\.js/, matches: 0},
		{name: 'Vercel Edge Functions', regex: /Vercel Edge Functions/, matches: 0},
		{name: 'OpenAI Image Downloader', regex: /OpenAI Image Downloader/, matches: 0},
		{name: 'OpenAI', regex: /OpenAI/, matches: 0},
		{name: 'Tsunami Security Scanner', regex: /TsunamiSecurityScanner/, matches: 0},
		{name: 'iOS', regex: /iOS\//, matches: 0},
		{name: 'Safari', regex: /Safari\//, matches: 0},
		{name: 'Edge', regex: /Edg\//, matches: 0},
		{name: 'Opera', regex: /(OPR|Opera)\//, matches: 0},
		{name: 'Internet Explorer', regex: /(; MSIE |Trident\/)/, matches: 0},
	]

	function getClient(userAgent: string): string {
		if (userAgent == null) {
			return 'Unknown';
		}

		for (let i = 0; i < clientCandidates.length; i++) {
			const candidate = clientCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
				// Ensure clientCandidates remains sorted by matches desc for future hits
				maintainClientCandidates(i);
				return candidate.name;
			}
		}

		return 'Other';
	}

	function maintainClientCandidates(indexUpdated: number) {
		let j = indexUpdated;
    	while (j > 0 && count > clientCandidates[j - 1].matches) {
        	j--
    	}
    	if (j < indexUpdated) {
        	[clientCandidates[indexUpdated], clientCandidates[j]] = [clientCandidates[j], clientCandidates[indexUpdated]]
    	}
	}

	function getChartData() {
		const clientCount: ValueCount = {};
		const clientGetter = cachedFunction(getClient)
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const client = clientGetter(userAgent);
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
					label: 'Client',
					data: counts,
					backgroundColor: graphColors,
					hoverOffset: 4,
				},
			],
		};
	}

	function genPlot() {
		const data = getChartData();

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
