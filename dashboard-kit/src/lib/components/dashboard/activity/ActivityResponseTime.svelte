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
				fixedrange: true
			},
			xaxis: {
				title: { text: 'Date' },
				showgrid: false,
				fixedrange: true,
				range: [periodAgo, now],
				visible: false
			},
			dragmode: false
		};
	}

	function bars(data: RequestsData, period: Period) {
		const responseTimesFreq = initFreqMap(period, () => ({
			totalResponseTime: 0,
			count: 0
		}));

		const days = periodToDays(period);

		for (const row of data) {
			// const timestamp = row[ColumnIndex.CreatedAt].getTime();

			// // Normalize timestamps efficiently
			// let time: number;
			// if (days === 1) {
			// 	time = Math.floor(timestamp / (5 * 60 * 1000)) * (5 * 60 * 1000); // Round to 5 min
			// } else if (days === 7) {
			// 	time = Math.floor(timestamp / 3600000) * 3600000; // Round to hour
			// } else {
			// 	time = Math.floor(timestamp / 86400000) * 86400000; // Round to day
			// }

			const date = new Date(row[ColumnIndex.CreatedAt]);
			if (days === 1) {
				// Round down to multiple of 5
				date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
			} else if (days === 7) {
				date.setMinutes(0, 0, 0);
			} else {
				date.setHours(0, 0, 0, 0);
			}
			const time = date.getTime();

			const entry = responseTimesFreq.get(time);
			if (entry) {
				entry.totalResponseTime += row[ColumnIndex.ResponseTime];
				entry.count++;
			} else {
				responseTimesFreq.set(time, { totalResponseTime: row[ColumnIndex.ResponseTime], count: 1 });
			}
		}

		// Preallocate array for performance
		const responseTimeArr: { date: number; avgResponseTime: number }[] = new Array(
			responseTimesFreq.size
		);
		let i = 0;
		for (const [time, { totalResponseTime, count }] of responseTimesFreq.entries()) {
			responseTimeArr[i++] = {
				date: time,
				avgResponseTime: count > 0 ? totalResponseTime / count : 0
			};
		}

		// Sort by date (timestamps are numbers, so direct subtraction works)
		responseTimeArr.sort((a, b) => a.date - b.date);

		// Preallocate output arrays
		const len = responseTimeArr.length;
		const dates: Date[] = new Array(len);
		const responseTimes: number[] = new Array(len);
		let minAvgResponseTime = Infinity;

		for (let j = 0; j < len; j++) {
			dates[j] = new Date(responseTimeArr[j].date);
			responseTimes[j] = responseTimeArr[j].avgResponseTime;
			if (responseTimes[j] < minAvgResponseTime) {
				minAvgResponseTime = responseTimes[j];
			}
		}

		return [
			{
				x: dates,
				y: responseTimes,
				type: 'bar',
				marker: { color: '#707070' },
				hovertemplate: `<b>%{y:.1f}ms average</b><br>%{x|%d %b %Y %H:%M}</b><extra></extra>`,
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
