<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { ColumnIndex } from '$lib/consts';
	import { cachedFunction } from '$lib/cache';
	import { type Candidate, maintainCandidates } from '$lib/candidates';

	const clientCandidates: Candidate[] = [
		{ name: 'Curl', regex: /curl\//, matches: 0 },
		{ name: 'Postman', regex: /PostmanRuntime\//, matches: 0 },
		{ name: 'Insomnia', regex: /insomnia\//, matches: 0 },
		{ name: 'Python requests', regex: /python-requests\//, matches: 0 },
		{ name: 'Nodejs fetch', regex: /node-fetch\//, matches: 0 },
		{ name: 'Seamonkey', regex: /Seamonkey\//, matches: 0 },
		{ name: 'Firefox', regex: /Firefox\//, matches: 0 },
		{ name: 'Chrome', regex: /Chrome\//, matches: 0 },
		{ name: 'Chromium', regex: /Chromium\//, matches: 0 },
		{ name: 'aiohttp', regex: /aiohttp\//, matches: 0 },
		{ name: 'Python', regex: /Python\//, matches: 0 },
		{ name: 'Go http', regex: /[Gg]o-http-client\//, matches: 0 },
		{ name: 'Java', regex: /Java\//, matches: 0 },
		{ name: 'axios', regex: /axios\//, matches: 0 },
		{ name: 'Dart', regex: /Dart\//, matches: 0 },
		{ name: 'OkHttp', regex: /OkHttp\//, matches: 0 },
		{ name: 'Uptime Kuma', regex: /Uptime-Kuma\//, matches: 0 },
		{ name: 'undici', regex: /undici\//, matches: 0 },
		{ name: 'Lush', regex: /Lush\//, matches: 0 },
		{ name: 'Zabbix', regex: /Zabbix/, matches: 0 },
		{ name: 'Guzzle', regex: /GuzzleHttp\//, matches: 0 },
		{ name: 'Uptime', regex: /Better Uptime/, matches: 0 },
		{ name: 'GitHub Camo', regex: /github-camo/, matches: 0 },
		{ name: 'Ruby', regex: /Ruby/, matches: 0 },
		{ name: 'Node.js', regex: /node/, matches: 0 },
		{ name: 'Next.js', regex: /Next\.js/, matches: 0 },
		{
			name: 'Vercel Edge Functions',
			regex: /Vercel Edge Functions/,
			matches: 0
		},
		{
			name: 'OpenAI Image Downloader',
			regex: /OpenAI Image Downloader/,
			matches: 0
		},
		{ name: 'OpenAI', regex: /OpenAI/, matches: 0 },
		{
			name: 'Tsunami Security Scanner',
			regex: /TsunamiSecurityScanner/,
			matches: 0
		},
		{ name: 'iOS', regex: /iOS\//, matches: 0 },
		{ name: 'Safari', regex: /Safari\//, matches: 0 },
		{ name: 'Edge', regex: /Edg\//, matches: 0 },
		{ name: 'Opera', regex: /(OPR|Opera)\//, matches: 0 },
		{ name: 'Internet Explorer', regex: /(; MSIE |Trident\/)/, matches: 0 }
	];

	function getClient(userAgent: string | null): string {
		if (!userAgent) {
			return 'Unknown';
		}

		for (let i = 0; i < clientCandidates.length; i++) {
			const candidate = clientCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
				// Ensure clientCandidates remains sorted by matches desc for future hits
				maintainCandidates(i, clientCandidates);
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
			margin: { r: 30, l: 30, t: 10, b: 25, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 196,
			width: 411,
			yaxis: {
				title: { text: 'Requests' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true
			},
			xaxis: {
				visible: false
			},
			dragmode: false
		};
	}

	function donut(data: RequestsData) {
		const clientCount: ValueCount = {};
		const clientGetter = cachedFunction(getClient);
		for (let i = 0; i < data.length; i++) {
			const userAgent = userAgents[data[i][ColumnIndex.UserAgent]] || '';
			const client = clientGetter(userAgent);
			if (client in clientCount) {
				clientCount[client]++;
			} else {
				clientCount[client] = 1;
			}
		}

		const dataPoints = Object.entries(clientCount).sort((a, b) => b[1] - a[1]);

		const clients = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [client, count] of dataPoints) {
			clients[i] = client;
			counts[i] = count;
			i++;
		}

		return [
			{
				values: counts,
				labels: clients,
				type: 'pie',
				hole: 0.6,
				marker: {
					colors: graphColors
				}
			}
		];
	}

	function getPlotData(data: RequestsData) {
		return {
			data: donut(data),
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
