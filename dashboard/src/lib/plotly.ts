declare const Plotly: any;
import { periodToDays, type Period } from '$lib/period';

const config = { responsive: true, showSendToCloud: false, displayModeBar: false };

/** Calls react() if the chart exists, newPlot() otherwise. Shared by all chart components. */
export function renderPlot(plotDiv: HTMLElement & { data?: unknown }, data: object[], layout: object): void {
	if (plotDiv.data) {
		Plotly.react(plotDiv, data, layout);
	} else {
		Plotly.newPlot(plotDiv, data, layout, config);
	}
}

/** Layout for Requests / Users / SuccessRate sparklines */
export function sparklineLayout() {
	return {
		title: false,
		autosize: true,
		margin: { r: 0, l: 0, t: 0, b: 0, pad: 0 },
		hovermode: false,
		plot_bgcolor: 'transparent',
		paper_bgcolor: 'transparent',
		height: 60,
		yaxis: { gridcolor: 'gray', showgrid: false, fixedrange: true, dragmode: false },
		xaxis: { visible: false, dragmode: false },
		dragmode: false,
	};
}

/** Data trace for Requests / Users / SuccessRate sparklines */
export function sparklineData(buckets: number[]) {
	return [{
		x: [...Array(buckets.length).keys()],
		y: buckets,
		type: 'lines',
		marker: { color: 'transparent' },
		showlegend: false,
		line: { shape: 'spline', smoothing: 1, color: '#3FCF8E30' },
		fill: 'tozeroy',
		fillcolor: '#3fcf8e15',
	}];
}

/** Layout for Client / DeviceType / OperatingSystem / Version donut charts */
export function donutLayout(width?: number) {
	return {
		title: false,
		autosize: true,
		margin: { r: 30, l: 30, t: 25, b: 25, pad: 0 },
		hovermode: 'closest',
		plot_bgcolor: 'transparent',
		paper_bgcolor: 'transparent',
		height: 196,
		...(width !== undefined ? { width } : {}),
		yaxis: { title: { text: 'Requests' }, gridcolor: 'gray', showgrid: false, fixedrange: true },
		xaxis: { visible: false },
		dragmode: false,
	};
}

/** Data trace for donut/pie charts */
export function donutData(labels: string[], values: number[], colors: string[]) {
	return [{ values, labels, type: 'pie', hole: 0.6, marker: { colors } }];
}

/** Layout for ActivityRequests / ActivityResponseTime bar charts */
export function activityLayout(period: Period, yAxisTitle: string, barmode?: string) {
	const days = periodToDays(period);
	const periodAgo = days !== null ? new Date(Date.now() - days * 86_400_000) : null;
	return {
		title: false,
		autosize: true,
		margin: { r: 35, l: 70, t: 20, b: 20, pad: 10 },
		hovermode: 'closest',
		plot_bgcolor: 'transparent',
		paper_bgcolor: 'transparent',
		height: 159,
		...(barmode ? { barmode } : {}),
		yaxis: { title: { text: yAxisTitle }, gridcolor: 'gray', showgrid: false, fixedrange: true },
		xaxis: { title: { text: 'Date' }, showgrid: false, fixedrange: true, range: [periodAgo, new Date()], visible: false },
		dragmode: false,
	};
}
