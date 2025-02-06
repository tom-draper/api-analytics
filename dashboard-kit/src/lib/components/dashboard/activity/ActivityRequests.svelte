<script lang="ts">
	import { periodToDays, type Period } from '$lib/period';
	import { initFreqMap } from '$lib/activity';
	import { ColumnIndex } from '$lib/consts';

	function getPlotLayout(period: Period) {
		const days = periodToDays(period);
		let periodAgo: Date | null = null;
		if (days !== null) {
			periodAgo = new Date();
			periodAgo.setDate(periodAgo.getDate() - days);
		}
		const now = new Date();

		return {
			title: false,
			autosize: true,
			margin: { r: 35, l: 70, t: 20, b: 20, pad: 10 },
			hovermode: 'closest',
			plot_bgcolor: 'transparent',
			paper_bgcolor: 'transparent',
			height: 159,
			barmode: 'stack',
			yaxis: {
				title: { text: 'Requests' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true
			},
			xaxis: {
				title: { text: 'Date' },
				fixedrange: true,
				range: [periodAgo, now],
				visible: false
			},
			dragmode: false
		};
	}

	function modifyDate(date: Date, days: number | null) {
		if (days === 1) {
			// Round down to multiple of 5
			date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
		} else if (days === 7) {
			date.setMinutes(0, 0, 0);
		} else {
			date.setHours(0, 0, 0, 0);
		}
		return date;
	}

	function bars(data: RequestsData, period: Period) {
		const requestFreq = initFreqMap(period, () => ({ count: 0 }));
		const userFreq = initFreqMap(period, () => new Set());

		const days = periodToDays(period);

		for (let i = 0; i < data.length; i++) {
			const date = modifyDate(new Date(data[i][ColumnIndex.CreatedAt]), days);
			const ipAddress = data[i][ColumnIndex.IPAddress];
			const time = date.getTime();

			let userSet = userFreq.get(time);
			if (!userSet) {
				userSet = new Set();
				userFreq.set(time, userSet);
			}
			userSet.add(ipAddress);

			// Update request frequency
			let freqObj = requestFreq.get(time);
			if (!freqObj) {
				freqObj = { count: 0 };
				requestFreq.set(time, freqObj);
			}
			freqObj.count++;
		}

		const requestFreqArr = Array.from(requestFreq, ([time, requestsCount]) => ({
			date: time,
			requestCount: requestsCount.count,
			userCount: userFreq.get(time)?.size || 0
		})).sort((a, b) => a.date - b.date);

		const dates = requestFreqArr.map(({ date }) => new Date(date));
		const requests = requestFreqArr.map(({ requestCount, userCount }) => requestCount - userCount);
		const users = requestFreqArr.map(({ userCount }) => userCount);
		const requestsText = requestFreqArr.map(({ requestCount }) => `${requestCount} requests`);
		const usersText = requestFreqArr.map(
			({ requestCount, userCount }) => `${requestCount} requests from ${userCount} users`
		);

		return [
			{
				x: dates,
				y: users,
				text: usersText,
				type: 'bar',
				marker: { color: '#3fcf8e' },
				hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
				showlegend: false
			},
			{
				x: dates,
				y: requests,
				text: requestsText,
				type: 'bar',
				marker: { color: '#228458' },
				hovertemplate: `<b>%{text}</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
				showlegend: false
			}
		];
	}

	function getPlotData(data: RequestsData, period: Period) {
		return {
			data: bars(data, period),
			layout: getPlotLayout(period),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false
			}
		};
	}

	function generatePlot(data: RequestsData, period: Period) {
		if (plotDiv.data) {
			refreshPlot(data, period);
		} else {
			newPlot(data, period);
		}
	}

	async function newPlot(data: RequestsData, period: Period) {
		const plotData = getPlotData(data, period);
		Plotly.newPlot(plotDiv, plotData.data, plotData.layout, plotData.config);
	}

	function refreshPlot(data: RequestsData, period: Period) {
		Plotly.react(plotDiv, bars(data, period), getPlotLayout(period));
	}

	let plotDiv: HTMLDivElement;

	$: if (plotDiv && data) {
		generatePlot(data, period);
	}

	export let data: RequestsData, period: Period;
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
