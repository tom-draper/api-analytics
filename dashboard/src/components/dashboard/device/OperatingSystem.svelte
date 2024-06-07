<script lang="ts">
	import { onMount } from 'svelte';
	import { ColumnIndex, graphColors } from '../../../lib/consts';
	import { cachedFunction } from '../../../lib/cache';
	import { type Candidate, maintainCandidates } from '../../../lib/candidates';
	import { Chart } from 'chart.js/auto';

	const osCandidates: Candidate[] = [
    	{name: 'Windows 3.11', regex: /Win16/, matches: 0},
    	{name: 'Windows 95', regex: /(Windows 95)|(Win95)|(Windows_95)/, matches: 0},
    	{name: 'Windows 98', regex: /(Windows 98)|(Win98)/, matches: 0},
    	{name: 'Windows 2000', regex: /(Windows NT 5.0)|(Windows 2000)/, matches: 0},
    	{name: 'Windows XP', regex: /(Windows NT 5.1)|(Windows XP)/, matches: 0},
    	{name: 'Windows Server 2003', regex: /(Windows NT 5.2)/, matches: 0},
    	{name: 'Windows Vista', regex: /(Windows NT 6.0)/, matches: 0},
    	{name: 'Windows 7', regex: /(Windows NT 6.1)/, matches: 0},
    	{name: 'Windows 8', regex: /(Windows NT 6.2)/, matches: 0},
    	{name: 'Windows 10/11', regex: /(Windows NT 10.0)/, matches: 0},
    	{name: 'Windows NT 4.0', regex: /(Windows NT 4.0)|(WinNT4.0)|(WinNT)|(Windows NT)/, matches: 0},
    	{name: 'Windows ME', regex: /Windows ME/, matches: 0},
    	{name: 'OpenBSD', regex: /OpenBSD/, matches: 0},
    	{name: 'SunOS', regex: /SunOS/, matches: 0},
    	{name: 'Android', regex: /Android/, matches: 0},
    	{name: 'Linux', regex: /(Linux)|(X11)/, matches: 0},
    	{name: 'MacOS', regex: /(Mac_PowerPC)|(Macintosh)/, matches: 0},
    	{name: 'QNX', regex: /QNX/, matches: 0},
    	{name: 'iOS', regex: /iPhone OS/, matches: 0},
    	{name: 'BeOS', regex: /BeOS/, matches: 0},
    	{name: 'OS/2', regex: /OS\/2/, matches: 0},
    	{name: 'Search Bot', regex: /(APIs-Google)|(AdsBot)|(nuhk)|(Googlebot)|(Storebot)|(Google-Site-Verification)|(Mediapartners)|(Yammybot)|(Openbot)|(Slurp)|(MSNBot)|(Ask Jeeves\/Teoma)|(ia_archiver)/, matches: 0}
	];

	function getOS(userAgent: string): string {
		if (userAgent === null) {
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

	function getChartData() {
		const osCount: ValueCount = {};
		const osGetter = cachedFunction(getOS)
		for (let i = 0; i < data.length; i++) {
			const userAgent = getUserAgent(data[i][ColumnIndex.UserAgent]);
			const os = osGetter(userAgent);
			if (os in osCount) {
				osCount[os]++;
			} else {
				osCount[os] = 1;
			}
		}

		const dataPoints = Object.entries(osCount).sort(
			(a, b) => b[1] - a[1],
		);

		const oss = new Array(dataPoints.length);
		const counts = new Array(dataPoints.length);
		let i = 0;
		for (const [os, count] of dataPoints) {
			oss[i] = os;
			counts[i] = count;
			i++;
		}

		return {
			labels: oss,
			datasets: [
				{
					label: 'Operating Systems',
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
