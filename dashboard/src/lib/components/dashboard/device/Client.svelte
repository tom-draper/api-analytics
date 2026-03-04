<script lang="ts">
	import { graphColors } from '$lib/consts';
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
		{ name: 'Vercel Edge Functions', regex: /Vercel Edge Functions/, matches: 0 },
		{ name: 'OpenAI Image Downloader', regex: /OpenAI Image Downloader/, matches: 0 },
		{ name: 'OpenAI', regex: /OpenAI/, matches: 0 },
		{ name: 'Tsunami Security Scanner', regex: /TsunamiSecurityScanner/, matches: 0 },
		{ name: 'iOS', regex: /iOS\//, matches: 0 },
		{ name: 'Safari', regex: /Safari\//, matches: 0 },
		{ name: 'Edge', regex: /Edg\//, matches: 0 },
		{ name: 'Opera', regex: /(OPR|Opera)\//, matches: 0 },
		{ name: 'Internet Explorer', regex: /(; MSIE |Trident\/)/, matches: 0 }
	];

	function getClient(userAgent: string | null): string {
		if (!userAgent) return 'Unknown';
		for (let i = 0; i < clientCandidates.length; i++) {
			const candidate = clientCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
				maintainCandidates(i, clientCandidates);
				return candidate.name;
			}
		}
		return 'Other';
	}

	function getPlotLayout() {
		return {
			title: false,
			autosize: true,
			margin: { r: 30, l: 30, t: 25, b: 25, pad: 0 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 196,
			width: 411,
			yaxis: { title: { text: 'Requests' }, gridcolor: 'gray', showgrid: false, fixedrange: true },
			xaxis: { visible: false },
			dragmode: false
		};
	}

	function donut(uaIdCount: { [id: number]: number }, userAgents: UserAgents) {
		const clientCount: ValueCount = {};
		const clientGetter = cachedFunction(getClient);
		for (const [uaId, count] of Object.entries(uaIdCount)) {
			const userAgent = userAgents[uaId as unknown as number] || '';
			const client = clientGetter(userAgent);
			clientCount[client] = (clientCount[client] ?? 0) + count;
		}

		const dataPoints = Object.entries(clientCount).sort((a, b) => b[1] - a[1]);
		const clients = dataPoints.map(([c]) => c);
		const counts = dataPoints.map(([, n]) => n);

		return [{ values: counts, labels: clients, type: 'pie', hole: 0.6, marker: { colors: graphColors } }];
	}

	function generatePlot(uaIdCount: { [id: number]: number }, userAgents: UserAgents) {
		const d = donut(uaIdCount, userAgents);
		const layout = getPlotLayout();
		const config = { responsive: true, showSendToCloud: false, displayModeBar: false };
		if (plotDiv.data) {
			Plotly.react(plotDiv, d, layout);
		} else {
			Plotly.newPlot(plotDiv, d, layout, config);
		}
	}

	let { uaIdCount, userAgents }: { uaIdCount: { [id: number]: number }; userAgents: UserAgents } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && uaIdCount) generatePlot(uaIdCount, userAgents);
	});
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
