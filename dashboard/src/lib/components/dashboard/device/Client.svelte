<script lang="ts">
	import { graphColors } from '$lib/consts';
	import { cachedFunction } from '$lib/cache';
	import { type Candidate, matchCandidate } from '$lib/candidates';
	import { renderPlot, donutLayout, buildDonutData } from '$lib/plotly';

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
		return matchCandidate(userAgent, clientCandidates);
	}

	const clientGetter = cachedFunction(getClient);

	let { uaIdCount, userAgents }: { uaIdCount: { [id: number]: number }; userAgents: UserAgents } = $props();
	let plotDiv = $state<HTMLDivElement | undefined>(undefined);

	$effect(() => {
		if (plotDiv && uaIdCount) renderPlot(plotDiv, buildDonutData(uaIdCount, userAgents, clientGetter, graphColors), donutLayout(411));
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
