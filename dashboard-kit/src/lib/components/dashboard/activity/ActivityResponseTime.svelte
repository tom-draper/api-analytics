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
			yaxis: {
				title: { text: 'Response time (ms)' },
				gridcolor: 'gray',
				showgrid: false,
				fixedrange: true,
			},
			xaxis: {
				title: { text: 'Date' },
				showgrid: false,
				fixedrange: true,
				range: [periodAgo, now],
				visible: false,
			},
			dragmode: false,
		};
	}

	function bars(data: RequestsData, period: Period) {
		const responseTimesFreq = initFreqMap(period, () => {
			return { totalResponseTime: 0, count: 0 };
		});

		const days = periodToDays(period);

		for (let i = 0; i < data.length; i++) {
			const date = new Date(data[i][ColumnIndex.CreatedAt]);
			if (days !== null && days === 1) {
				// Round down to multiple of 5
				date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
			} else if (days !== null && days === 7) {
				date.setMinutes(0, 0, 0);
			} else {
				date.setHours(0, 0, 0, 0);
			}
			const time = date.getTime();
			if (responseTimesFreq.has(time)) {
				responseTimesFreq.get(time).totalResponseTime +=
					data[i][ColumnIndex.ResponseTime];
				responseTimesFreq.get(time).count++;
			} else {
				responseTimesFreq.set(time, { totalResponseTime: 1, count: 1 });
			}
		}

		// Combine date and avg response time into (x, y) tuples for sorting
		const responseTimeArr: { date: number; avgResponseTime: number }[] =
			new Array(responseTimesFreq.size);
		let i = 0;
		for (const [time, obj] of responseTimesFreq.entries()) {
			const point = { date: time, avgResponseTime: 0 };
			if (obj.count > 0) {
				point.avgResponseTime = obj.totalResponseTime / obj.count;
			}
			responseTimeArr[i] = point;
			i++;
		}

		// Sort by date
		responseTimeArr.sort((a, b) => {
			//@ts-ignore
			return a.date - b.date;
		});

		// Split into two lists
		const dates: Date[] = new Array(responseTimeArr.length);
		const responseTimes: number[] = new Array(responseTimeArr.length);
		let minAvgResponseTime = Number.POSITIVE_INFINITY;
		for (let i = 0; i < responseTimeArr.length; i++) {
			dates[i] = new Date(responseTimeArr[i].date);
			responseTimes[i] = responseTimeArr[i].avgResponseTime;
			if (responseTimeArr[i].avgResponseTime < minAvgResponseTime) {
				minAvgResponseTime = responseTimeArr[i].avgResponseTime;
			}
		}

		return [
			{
				x: dates,
				y: responseTimes,
				type: 'bar',
				marker: { color: '#707070' },
				hovertemplate: `<b>%{y:.1f}ms average</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
				showlegend: false,
			},
		];
	}

	function getPlotData(data: RequestsData, period: Period) {
		return {
			data: bars(data, period),
			layout: getPlotLayout(period),
			config: {
				responsive: true,
				showSendToCloud: false,
				displayModeBar: false,
			},
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
		Plotly.newPlot(
			plotDiv,
			plotData.data,
			plotData.layout,
			plotData.config,
		);
	}

	function refreshPlot(data: RequestsData, period: Period) {
		Plotly.react(
			plotDiv,
			bars(data, period),
			getPlotLayout(period),
		)
	}

	let plotDiv: HTMLDivElement;

	$: if (plotDiv) {
		generatePlot(data, period);
	}

	export let data: RequestsData, period: Period;
</script>

<div id="plotly">
	<div id="plotDiv" bind:this={plotDiv}>
		<!-- Plotly chart will be drawn inside this DIV -->
	</div>
</div>
