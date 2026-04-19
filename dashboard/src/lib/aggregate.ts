import { ColumnIndex, methodMap, topListLimit } from '$lib/consts';
import { periodToDays, type Period } from '$lib/period';
import { statusSuccessful } from '$lib/status';
import type { DashboardSettings } from '$lib/settings';

export type ActivityBucket = {
	date: number;
	requestCount: number;
	userCount: number;
	avgResponseTime: number;
	successRate: number;
};

export type LocationBar = {
	location: string;
	frequency: number;
	height: number;
};

export type TopUserData = {
	ipAddress: string;
	customUserID: string;
	lastRequested: Date;
	requests: number;
	locations: { [loc: string]: number };
};

export type ReferrerBar = {
	referrer: string;
	count: number;
	height: number;
};

export type UserIDBar = {
	userID: string;
	count: number;
	height: number;
};

export type AggregatedData = {
	period: Period;
	requestBuckets: number[];
	requestCount: number;
	prevRequestCount: number;
	firstRequestDate: Date | null;
	lastRequestDate: Date | null;

	userBuckets: number[];
	userCount: number;
	prevUserCount: number;

	successRate: number | null;
	successBuckets: number[];

	activityBuckets: ActivityBucket[];

	rtFreqTimes: number[];
	rtFreqCounts: number[];
	rtLQ: number;
	rtMedian: number;
	rtUQ: number;

	versionCount: { [v: string]: number };
	versionHasMultiple: boolean;

	hourlyBuckets: number[];
	weekdayBuckets: number[];

	locationBars: LocationBar[];
	referrerBars: ReferrerBar[];
	referrerAvailable: boolean;
	userIDBars: UserIDBar[];
	userIDAvailable: boolean;

	endpointFreq: { [key: string]: { path: string; status: number; count: number } };

	topUsers: TopUserData[] | null;
	topUserIDActive: boolean;
	topLocationsActive: boolean;

	uaIdCount: { [id: number]: number };
};

const TOP_USERS_LIMIT = 1000;

function buildBars<T>(
	counts: { [key: string]: number },
	mapper: (key: string, count: number, height: number) => T,
	limit?: number
): T[] {
	let max = 0;
	const entries = Object.entries(counts);
	for (const [, count] of entries) {
		if (count > max) max = count;
	}
	const bars = entries
		.sort((a, b) => b[1] - a[1])
		.map(([key, count]) => mapper(key, count, max > 0 ? count / max : 0));
	return limit !== undefined ? bars.slice(0, limit) : bars;
}

function quantileFromFreq(times: number[], counts: number[], total: number, q: number): number {
	if (total === 0) return 0;
	const pos = (total - 1) * q;
	const base = Math.floor(pos);
	const rest = pos - base;
	let cumulative = 0;
	let valAtBase: number | undefined;
	let valAtBaseP1: number | undefined;
	for (let i = 0; i < times.length; i++) {
		cumulative += counts[i];
		if (valAtBase === undefined && cumulative > base) valAtBase = times[i];
		if (valAtBaseP1 === undefined && cumulative > base + 1) {
			valAtBaseP1 = times[i];
			break;
		}
	}
	if (valAtBase === undefined) return 0;
	if (valAtBaseP1 === undefined) return valAtBase;
	return valAtBase + rest * (valAtBaseP1 - valAtBase);
}

function modifyDateForPeriod(date: Date, days: number | null): void {
	if (days === 1) {
		date.setMinutes(Math.floor(date.getMinutes() / 5) * 5, 0, 0);
	} else if (days === 7) {
		date.setMinutes(0, 0, 0);
	} else if (days === 30) {
		date.setMinutes(0, 0, 0);
		date.setHours(Math.floor(date.getHours() / 6) * 6);
	} else {
		date.setHours(0, 0, 0, 0);
	}
}

type ActivityEntry = { requestCount: number; users: Set<string>; totalRT: number; successCount: number };

function initActivityMap(days: number | null): Map<number, ActivityEntry> {
	const map = new Map<number, ActivityEntry>();
	if (days === null) return map;
	const newEntry = (): ActivityEntry => ({ requestCount: 0, users: new Set(), totalRT: 0, successCount: 0 });

	if (days === 1) {
		for (let i = 0; i < 288; i++) {
			const date = new Date();
			date.setSeconds(0, 0);
			date.setMinutes(Math.floor(date.getMinutes() / 5) * 5 - i * 5);
			map.set(date.getTime(), newEntry());
		}
	} else if (days === 7) {
		for (let i = 0; i < 168; i++) {
			const date = new Date();
			date.setSeconds(0, 0);
			date.setMinutes(0);
			date.setHours(date.getHours() - i);
			map.set(date.getTime(), newEntry());
		}
	} else if (days === 30) {
		for (let i = 0; i < 120; i++) {
			const date = new Date();
			date.setSeconds(0, 0);
			date.setMinutes(0);
			date.setHours(Math.floor(date.getHours() / 6) * 6 - i * 6);
			map.set(date.getTime(), newEntry());
		}
	} else {
		for (let i = 0; i < days; i++) {
			const date = new Date();
			date.setHours(0, 0, 0, 0);
			date.setDate(date.getDate() - i);
			map.set(date.getTime(), newEntry());
		}
	}
	return map;
}

function countPrevUsers(previous: RequestsData): number {
	const users = new Set<string>();
	for (let i = 0; i < previous.length; i++) {
		const row = previous[i];
		const uid = row[ColumnIndex.IPAddress] ?? row[ColumnIndex.UserID]?.toString() ?? '';
		if (uid) users.add(uid);
	}
	return users.size;
}

function emptyResult(previous: RequestsData, settings: DashboardSettings): AggregatedData {
	return {
		period: settings.period,
		requestBuckets: [0, 0, 0, 0, 0],
		requestCount: 0,
		prevRequestCount: previous.length,
		firstRequestDate: null,
		lastRequestDate: null,
		userBuckets: [0, 0, 0, 0, 0],
		userCount: 0,
		prevUserCount: countPrevUsers(previous),
		successRate: null,
		successBuckets: [0, 0, 0, 0, 0],
		activityBuckets: [],
		rtFreqTimes: [],
		rtFreqCounts: [],
		rtLQ: 0,
		rtMedian: 0,
		rtUQ: 0,
		versionCount: {},
		versionHasMultiple: false,
		hourlyBuckets: new Array(24).fill(0),
		weekdayBuckets: new Array(7).fill(0),
		locationBars: [],
		referrerBars: [],
		referrerAvailable: false,
		userIDBars: [],
		userIDAvailable: false,
		endpointFreq: {},
		topUsers: null,
		topUserIDActive: false,
		topLocationsActive: false,
		uaIdCount: {},
	};
}

export function aggregate(
	current: RequestsData,
	previous: RequestsData,
	settings: DashboardSettings
): AggregatedData {
	const n = current.length;
	if (n === 0) return emptyResult(previous, settings);

	const days = periodToDays(settings.period);
	const now = Date.now();
	const periodMs = days ? days * 86400000 : (n > 0 ? (current[n - 1][ColumnIndex.CreatedAt] as Date).getTime() - (current[0][ColumnIndex.CreatedAt] as Date).getTime() : 0);
	const startMs = days ? now - periodMs : (n > 0 ? (current[0][ColumnIndex.CreatedAt] as Date).getTime() : 0);
	const bucketInterval = Math.max(periodMs / 5, 1);

	const requestBuckets = [0, 0, 0, 0, 0];
	const userSets: Set<string>[] = [new Set(), new Set(), new Set(), new Set(), new Set()];
	const successBuckets = [0, 0, 0, 0, 0];
	let successCount = 0;

	const activityMap = initActivityMap(days);

	const rtFreq: { [rt: number]: number } = {};

	const versionCount: { [v: string]: number } = {};
	let versionTypes = 0;
	const pathVersionCache = new Map<string, string | null>();

	const hourlyBuckets = new Array(24).fill(0);
	const weekdayBuckets = new Array(7).fill(0);
	const locationCount: { [loc: string]: number } = {};
	const referrerCount: { [ref: string]: number } = {};
	let referrerAvailable = false;
	const userIDCount: { [id: string]: number } = {};
	let userIDAvailable = false;
	const endpointFreq: { [key: string]: { path: string; status: number; count: number } } = {};

	type UserEntry = {
		ipAddress: string;
		customUserID: string;
		lastRequested: Date;
		requests: number;
		locations: { [loc: string]: number };
	};
	const users: { [userID: string]: UserEntry } = {};
	const uaIdCount: { [id: number]: number } = {};

	// Reuse a single Date object for activity bucketing to avoid per-row allocation
	const actDate = new Date(0);

	for (let i = 0; i < n; i++) {
		const row = current[i];
		const createdAt = row[ColumnIndex.CreatedAt] as Date;
		const dateMs = createdAt.getTime();
		const status = row[ColumnIndex.Status];
		const path = row[ColumnIndex.Path];
		const location = row[ColumnIndex.Location];
		const ipAddress = row[ColumnIndex.IPAddress];
		const customUserID = row[ColumnIndex.UserID];
		const responseTime = row[ColumnIndex.ResponseTime];
		const uaId = row[ColumnIndex.UserAgent];
		const method = row[ColumnIndex.Method];
		const userID = ipAddress ?? customUserID?.toString() ?? '';
		const isSuccessful = statusSuccessful(status);

		// 1. Request buckets (time-based)
		const bucketIdx = Math.min(Math.floor((dateMs - startMs) / bucketInterval), 4);
		requestBuckets[bucketIdx]++;

		// 2. User buckets (unique users per time bucket)
		if (userID) userSets[bucketIdx].add(userID);

		// 3. Success buckets (time-based) + total success rate
		if (isSuccessful) {
			successBuckets[bucketIdx]++;
			successCount++;
		}

		// 4. Activity buckets (period-aware time buckets)
		actDate.setTime(dateMs);
		modifyDateForPeriod(actDate, days);
		const actTime = actDate.getTime();
		let actEntry = activityMap.get(actTime);
		if (!actEntry) {
			actEntry = { requestCount: 0, users: new Set(), totalRT: 0, successCount: 0 };
			activityMap.set(actTime, actEntry);
		}
		actEntry.requestCount++;
		if (ipAddress) actEntry.users.add(ipAddress);
		actEntry.totalRT += responseTime;
		if (isSuccessful) actEntry.successCount++;

		// 5. Response time frequency
		const rtRounded = Math.round(responseTime) || 0;
		rtFreq[rtRounded] = (rtFreq[rtRounded] ?? 0) + 1;

		// 6. Version (regex on path, cached per unique path)
		let cachedVersion = pathVersionCache.get(path);
		if (cachedVersion === undefined) {
			const vMatch = path.match(/[^a-z0-9](v\d)[^a-z0-9]/i);
			cachedVersion = vMatch ? vMatch[1] : null;
			pathVersionCache.set(path, cachedVersion);
		}
		if (cachedVersion !== null) {
			if (!(cachedVersion in versionCount)) versionTypes++;
			versionCount[cachedVersion] = (versionCount[cachedVersion] ?? 0) + 1;
		}

		// 7. Usage time (24-hour clock buckets) + day of week
		hourlyBuckets[createdAt.getHours()]++;
		weekdayBuckets[createdAt.getDay()]++;

		// 8. Location
		if (location) {
			locationCount[location] = (locationCount[location] ?? 0) + 1;
		}

		// 8b. Referrer
		const referrer = row[ColumnIndex.Referrer] as string | null | undefined;
		if (referrer) {
			const ref = settings.ignoreParams ? referrer.split('?')[0] : referrer;
			referrerCount[ref] = (referrerCount[ref] ?? 0) + 1;
			referrerAvailable = true;
		}

		// 8c. Custom user ID
		if (customUserID) {
			userIDCount[customUserID] = (userIDCount[customUserID] ?? 0) + 1;
			userIDAvailable = true;
		}

		// 9. Endpoints
		const ePath = settings.ignoreParams ? path.split('?')[0] : path;
		const eKey = `${ePath}${status}`;
		let ep = endpointFreq[eKey];
		if (!ep) {
			ep = { path: `${methodMap[method]}  ${ePath}`, status, count: 0 };
			endpointFreq[eKey] = ep;
		}
		ep.count++;

		// 10. Top users (data is sorted ascending so last seen is always most recent)
		if (userID) {
			let user = users[userID];
			if (user) {
				user.requests++;
				if (location) user.locations[location] = (user.locations[location] || 0) + 1;
				user.lastRequested = createdAt;
			} else {
				users[userID] = {
					ipAddress: ipAddress ?? '',
					customUserID: customUserID ?? '',
					lastRequested: createdAt,
					requests: 1,
					locations: location ? { [location]: 1 } : {},
				};
			}
		}

		// 11. UA ID count
		if (uaId != null) {
			uaIdCount[uaId] = (uaIdCount[uaId] ?? 0) + 1;
		}
	}

	// Build sparse response time frequency histogram and compute quartiles from it
	const rtFreqTimes = Object.keys(rtFreq).map(Number).sort((a, b) => a - b);
	const rtFreqCounts = rtFreqTimes.map((t) => rtFreq[t]);
	const rtLQ = quantileFromFreq(rtFreqTimes, rtFreqCounts, n, 0.25);
	const rtMedian = quantileFromFreq(rtFreqTimes, rtFreqCounts, n, 0.5);
	const rtUQ = quantileFromFreq(rtFreqTimes, rtFreqCounts, n, 0.75);

	// Build activity buckets (sorted ascending by date)
	const activityBuckets: ActivityBucket[] = Array.from(activityMap, ([date, entry]) => ({
		date,
		requestCount: entry.requestCount,
		userCount: entry.users.size,
		avgResponseTime: entry.requestCount > 0 ? entry.totalRT / entry.requestCount : 0,
		successRate: entry.requestCount > 0 ? entry.successCount / entry.requestCount : 0,
	})).sort((a, b) => a.date - b.date);

	// Build location, referrer, and user ID bars (sorted by count, heights normalized)
	const locationBars: LocationBar[] = buildBars(locationCount, (location, frequency, height) => ({ location, frequency, height }));
	const referrerBars: ReferrerBar[] = buildBars(referrerCount, (referrer, count, height) => ({ referrer, count, height }), topListLimit);
	const userIDBars: UserIDBar[] = buildBars(userIDCount, (userID, count, height) => ({ userID, count, height }), topListLimit);

	// Build top users
	const userValues = Object.values(users);
	const totalUsers = userValues.length;
	let topUsers: TopUserData[] | null = null;
	let topUserIDActive = false;
	let topLocationsActive = false;

	if (totalUsers >= 10 || settings.targetUser !== null) {
		const sorted = userValues.sort((a, b) => b.requests - a.requests);
		const topUserRequestsCount = sorted.length > 0 ? sorted[0].requests : 0;
		if (topUserRequestsCount > 1 || settings.targetUser !== null) {
			topUsers = sorted.slice(0, TOP_USERS_LIMIT);
			topUserIDActive = topUsers.some((u) => u.customUserID !== '' && u.customUserID !== null);
			topLocationsActive = topUsers.some((u) => Object.keys(u.locations).length > 0);
		}
	}

	const prevUserCount = countPrevUsers(previous);

	return {
		period: settings.period,
		requestBuckets,
		requestCount: n,
		prevRequestCount: previous.length,
		firstRequestDate: current[0][ColumnIndex.CreatedAt] as Date,
		lastRequestDate: current[n - 1][ColumnIndex.CreatedAt] as Date,

		userBuckets: userSets.map((s) => s.size),
		userCount: totalUsers,
		prevUserCount,

		successRate: (successCount / n) * 100,
		successBuckets,

		activityBuckets,

		rtFreqTimes,
		rtFreqCounts,
		rtLQ,
		rtMedian,
		rtUQ,

		versionCount,
		versionHasMultiple: versionTypes > 1,

		hourlyBuckets,
		weekdayBuckets,
		locationBars,
		referrerBars,
		referrerAvailable,
		userIDBars,
		userIDAvailable,
		endpointFreq,

		topUsers,
		topUserIDActive,
		topLocationsActive,

		uaIdCount,
	};
}
