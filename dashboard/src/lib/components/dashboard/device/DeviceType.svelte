<script lang="ts">
	import { cachedFunction } from '$lib/cache';
	import { type Candidate, maintainCandidates } from '$lib/candidates';

	const deviceCandidates: Candidate[] = [
		{ name: 'iPhone', regex: /iPhone/, matches: 0 },
		{ name: 'Android', regex: /Android/, matches: 0 },
		{ name: 'Samsung', regex: /Tizen\//, matches: 0 },
		{ name: 'Mac', regex: /Macintosh/, matches: 0 },
		{ name: 'Windows', regex: /Windows/, matches: 0 }
	];

	function getDevice(userAgent: string | null): string {
		if (!userAgent) return 'Unknown';
		for (let i = 0; i < deviceCandidates.length; i++) {
			const candidate = deviceCandidates[i];
			if (userAgent.match(candidate.regex)) {
				candidate.matches++;
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

	const colors = ['#3FCF8E', '#E46161', '#EBEB81'];

	function donut(uaIdCount: { [id: number]: number }, userAgents: UserAgents) {
		const deviceCount: ValueCount = {};
		const deviceGetter = cachedFunction(getDevice);
		for (const [uaId, count] of Object.entries(uaIdCount)) {
			const userAgent = userAgents[uaId as unknown as number] || '';
			const device = deviceGetter(userAgent);
			deviceCount[device] = (deviceCount[device] ?? 0) + count;
		}

		const dataPoints = Object.entries(deviceCount).sort((a, b) => b[1] - a[1]);
		const devices = dataPoints.map(([d]) => d);
		const counts = dataPoints.map(([, n]) => n);

		return [{ values: counts, labels: devices, type: 'pie', hole: 0.6, marker: { colors } }];
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
