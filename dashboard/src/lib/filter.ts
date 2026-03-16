import { ColumnIndex } from "./consts";
import { nextDay, toDay } from "./date";
import { statusBad, statusError, statusRedirect, statusSuccess } from "./status";

export type Filter = {
	timespan: [number, number];
	status: {
		success: boolean;
		redirect: boolean;
		client: boolean;
		server: boolean;
	};
	methods: { [method: string]: boolean };
	hostnames: { [hostname: string]: boolean };
	responseTime: [number, number];
	paths: Set<string>;
};

export function defaultFilter(data: RequestsData): Filter {
	const methods: { [method: string]: boolean } = {};
	const hostnames: { [hostname: string]: boolean } = {};
	for (const row of data) {
		methods[row[ColumnIndex.Method]] = true;
		hostnames[row[ColumnIndex.Hostname]] = true;
	}
	return {
		timespan: [
			(data[0][ColumnIndex.CreatedAt] as Date).getTime(),
			(data[data.length - 1][ColumnIndex.CreatedAt] as Date).getTime()
		],
		status: { success: true, redirect: true, client: true, server: true },
		responseTime: [0, Infinity],
		methods,
		hostnames,
		paths: new Set()
	};
}

export function applyFilter(data: RequestsData, filter: Filter): RequestsData {
	const filtered: RequestsData = [];
	const startDate = toDay(new Date(filter.timespan[0]));
	const endDate = toDay(nextDay(new Date(filter.timespan[1])));
	const hasPathFilter = filter.paths.size > 0;

	for (const row of data) {
		const status = row[ColumnIndex.Status] as number;
		if (
			((filter.status.success && statusSuccess(status)) ||
				(filter.status.redirect && statusRedirect(status)) ||
				(filter.status.client && statusBad(status)) ||
				(filter.status.server && statusError(status))) &&
			(row[ColumnIndex.CreatedAt] as Date) >= startDate &&
			(row[ColumnIndex.CreatedAt] as Date) <= endDate &&
			filter.methods[row[ColumnIndex.Method] as number] &&
			filter.hostnames[row[ColumnIndex.Hostname] as string] &&
			(row[ColumnIndex.ResponseTime] as number) >= filter.responseTime[0] &&
			(row[ColumnIndex.ResponseTime] as number) <= filter.responseTime[1] &&
			(!hasPathFilter || filter.paths.has(row[ColumnIndex.Path] as string))
		) {
			filtered.push(row);
		}
	}
	return filtered;
}
