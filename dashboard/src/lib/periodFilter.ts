import { ColumnIndex } from '$lib/consts';
import { periodToDays } from '$lib/period';
import type { DashboardSettings } from '$lib/settings';
import { userTargeted } from '$lib/user';
import { clientCandidates, deviceCandidates, osCandidates, matchLabel } from '$lib/device';

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

const botPattern = /bot|crawl|spider|scraper|slurp|curl|wget|python|java(?!script)|ruby|php|perl|go-http|okhttp|libwww|httpclient|axios|requests|puppeteer|playwright|selenium|headless/i;

export function getPeriodData(
	data: RequestsData,
	settings: DashboardSettings,
	userAgents: UserAgents = {}
): { current: RequestsData; previous: RequestsData } {
	const now = Date.now();
	const hasHiddenEndpoints = settings.hiddenEndpoints.size > 0;
	const allTime = settings.period === 'all time';
	const pathVersionCache = new Map<string, string | null>();

	let botUAIds: Set<number> | null = null;
	if (settings.ignoreBots) {
		botUAIds = new Set();
		for (const [id, ua] of Object.entries(userAgents)) {
			if (botPattern.test(ua)) botUAIds.add(Number(id));
		}
	}

	function buildUASet(label: string, candidates: typeof clientCandidates): Set<number> {
		const set = new Set<number>();
		for (const [id, ua] of Object.entries(userAgents)) {
			if (matchLabel(ua, candidates) === label) set.add(Number(id));
		}
		return set;
	}

	const clientUAIds = settings.targetClient !== null ? buildUASet(settings.targetClient, clientCandidates) : null;
	const deviceUAIds = settings.targetDeviceType !== null ? buildUASet(settings.targetDeviceType, deviceCandidates) : null;
	const osUAIds = settings.targetOS !== null ? buildUASet(settings.targetOS, osCandidates) : null;

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
		const createdAt = request[ColumnIndex.CreatedAt] as Date;
		const weekday = createdAt.getDay();
		const hour = createdAt.getHours();

		let version: string | null | undefined;
		if (settings.targetVersion !== null) {
			const vPath = settings.ignoreParams ? path.split('?')[0] : path;
			version = pathVersionCache.get(vPath);
			if (version === undefined) {
				const m = vPath.match(/[^a-z0-9](v\d)[^a-z0-9]/i);
				version = m ? m[1] : null;
				pathVersionCache.set(vPath, version);
			}
		}

		if (
			(settings.targetUser === null ||
				userTargeted(settings.targetUser, ipAddress, customUserID)) &&
			(!settings.disable404 || status !== 404) &&
			(!botUAIds || !botUAIds.has(request[ColumnIndex.UserAgent] as number)) &&
			(!clientUAIds || clientUAIds.has(request[ColumnIndex.UserAgent] as number)) &&
			(!deviceUAIds || deviceUAIds.has(request[ColumnIndex.UserAgent] as number)) &&
			(!osUAIds || osUAIds.has(request[ColumnIndex.UserAgent] as number)) &&
			(settings.targetEndpoint.path === null || settings.targetEndpoint.path === path) &&
			(settings.targetEndpoint.status === null || settings.targetEndpoint.status === status) &&
			(settings.targetReferrer === null || settings.targetReferrer === referrer) &&
			(settings.targetLocation === null || settings.targetLocation === location) &&
			(settings.targetWeekday === null || settings.targetWeekday === weekday) &&
			(settings.targetHour === null || settings.targetHour === hour) &&
			(settings.targetVersion === null || settings.targetVersion === version) &&
			(!hasHiddenEndpoints || !isHiddenEndpoint(path, settings.hiddenEndpoints)) &&
			(settings.hostname === null || settings.hostname === hostname)
		) {
			if (allTime) {
				current.push(request);
			} else {
				const dateMs = createdAt.getTime();
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
