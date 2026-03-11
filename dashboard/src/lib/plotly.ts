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

/** Build donut chart data by aggregating UA IDs through a getter function */
export function buildDonutData(
	uaIdCount: { [id: number]: number },
	userAgents: UserAgents,
	getter: (ua: string) => string,
	colors: string[],
	selectedLabel?: string | null
): object[] {
	const count: { [key: string]: number } = {};
	for (const [uaId, c] of Object.entries(uaIdCount)) {
		const ua = userAgents[uaId as unknown as number] || '';
		const key = getter(ua);
		count[key] = (count[key] ?? 0) + c;
	}
	const dataPoints = Object.entries(count).sort((a, b) => b[1] - a[1]);
	const labels = dataPoints.map(([k]) => k);
	const values = dataPoints.map(([, v]) => v);
	const pull = selectedLabel != null ? labels.map((l) => (l === selectedLabel ? 0.08 : 0)) : undefined;
	return [{ values, labels, type: 'pie', hole: 0.6, marker: { colors }, ...(pull ? { pull } : {}) }];
}

/** Format a bucket start date as a time range label based on period bucket size */
export function bucketRange(date: Date, period: Period): string {
	const pad = (n: number) => String(n).padStart(2, '0');
	if (period === '24 hours') {
		const end = new Date(date.getTime() + 5 * 60 * 1000);
		return `${pad(date.getHours())}:${pad(date.getMinutes())}-${pad(end.getHours())}:${pad(end.getMinutes())}`;
	} else if (period === 'week') {
		const end = new Date(date.getTime() + 60 * 60 * 1000);
		const day = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][date.getDay()];
		return `${day} ${pad(date.getHours())}:00-${pad(end.getHours())}:00`;
	} else {
		return date.toLocaleDateString([], { day: 'numeric', month: 'short', year: 'numeric' });
	}
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
		xaxis: { title: { text: 'Date' }, showgrid: false, fixedrange: true, ...(periodAgo !== null ? { range: [periodAgo, new Date()] } : {}), visible: false },
		dragmode: false,
	};
}
