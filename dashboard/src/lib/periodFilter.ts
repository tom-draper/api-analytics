import { ColumnIndex } from '$lib/consts';
import { periodToDays } from '$lib/period';
import type { DashboardSettings } from '$lib/settings';
import { userTargeted } from '$lib/user';

function lowerBound(data: RequestsData, targetMs: number): number {
	let lo = 0;
	let hi = data.length;
	while (lo < hi) {
		const mid = (lo + hi) >>> 1;
		if ((data[mid][ColumnIndex.CreatedAt] as Date).getTime() < targetMs) {
			lo = mid + 1;
		} else {
			hi = mid;
		}
	}
	return lo;
}

function isHiddenEndpoint(endpoint: string, hiddenEndpoints: Set<string>): boolean {
	const normalized = endpoint.replace(/^\/|\/$/g, '');
	return (
		hiddenEndpoints.has(endpoint) ||
		hiddenEndpoints.has('/' + normalized) ||
		hiddenEndpoints.has(normalized) ||
		wildCardMatch(endpoint, hiddenEndpoints)
	);
}

function wildCardMatch(endpoint: string, hiddenEndpoints: Set<string>): boolean {
	endpoint = endpoint.endsWith('/') ? endpoint : endpoint + '/';
	for (const hidden of hiddenEndpoints) {
		if (!hidden.endsWith('*')) continue;
		let prefix = hidden.slice(0, -1);
		prefix = prefix.startsWith('/') ? prefix : '/' + prefix;
		prefix = prefix.endsWith('/') ? prefix : prefix + '/';
		if (endpoint.startsWith(prefix)) return true;
	}
	return false;
}

export function getPeriodData(
	data: RequestsData,
	settings: DashboardSettings
): { current: RequestsData; previous: RequestsData } {
	const now = Date.now();
	const hasHiddenEndpoints = settings.hiddenEndpoints.size > 0;
	const allTime = settings.period === 'all time';

	const current: RequestsData = [];
	const previous: RequestsData = [];

	let startIdx = 0;
	let currentStartMs = 0;
	let prevStartMs = 0;

	if (!allTime) {
		const days = periodToDays(settings.period)!;
		const periodMs = days * 86400000;
		currentStartMs = now - periodMs;
		prevStartMs = now - periodMs * 2;
		startIdx = lowerBound(data, prevStartMs);
	}

	for (let i = startIdx; i < data.length; i++) {
		const request = data[i];
		const status = request[ColumnIndex.Status];
		const path = request[ColumnIndex.Path];
		const hostname = request[ColumnIndex.Hostname];
		const location = request[ColumnIndex.Location];
		const ipAddress = request[ColumnIndex.IPAddress];
		const customUserID = request[ColumnIndex.UserID];
		const referrer = request[ColumnIndex.Referrer] as string | null | undefined;

		if (
			(settings.targetUser === null ||
				userTargeted(settings.targetUser, ipAddress, customUserID)) &&
			(!settings.disable404 || status !== 404) &&
			(settings.targetEndpoint.path === null || settings.targetEndpoint.path === path) &&
			(settings.targetEndpoint.status === null || settings.targetEndpoint.status === status) &&
			(settings.targetReferrer === null || settings.targetReferrer === referrer) &&
			(settings.targetLocation === null || settings.targetLocation === location) &&
			(!hasHiddenEndpoints || !isHiddenEndpoint(path, settings.hiddenEndpoints)) &&
			(settings.hostname === null || settings.hostname === hostname)
		) {
			if (allTime) {
				current.push(request);
			} else {
				const dateMs = (request[ColumnIndex.CreatedAt] as Date).getTime();
				if (dateMs >= currentStartMs) {
					current.push(request);
				} else if (dateMs >= prevStartMs) {
					previous.push(request);
				}
			}
		}
	}

	return { current, previous };
}

export function getHostnames(data: RequestsData): string[] {
	const freq: ValueCount = {};
	for (let i = 0; i < data.length; i++) {
		const hostname = data[i][ColumnIndex.Hostname];
		if (!hostname) continue;
		freq[hostname] = (freq[hostname] ?? 0) + 1;
	}
	return Object.entries(freq)
		.sort((a, b) => b[1] - a[1])
		.map(([h]) => h);
}
