import { defaultPeriod, isPeriod, type Period } from "./period";

export type DashboardSettings = {
	disable404: boolean;
	ignoreBots: boolean;
	hostname: string | null;
	period: Period;
	targetEndpoint: {
		path: string | null;
		status: number | null;
	};
	targetReferrer: string | null;
	targetLocation: string | null;
	targetWeekday: number | null;
	targetHour: number | null;
	targetVersion: string | null;
	targetClient: string | null;
	targetDeviceType: string | null;
	targetOS: string | null;
	targetUser: {
		ipAddress: string,
		userID: string,
		composite: boolean
	} | null;
	hiddenEndpoints: Set<string>;
	ignoreParams: boolean;
};

export function parseSettingsFromURL(searchParams: URLSearchParams): DashboardSettings {
	const settings = initSettings();

	const period = searchParams.get('period');
	if (period && isPeriod(period)) settings.period = period;

	const hostname = searchParams.get('hostname');
	if (hostname) settings.hostname = hostname;

	const location = searchParams.get('location');
	if (location) settings.targetLocation = location;

	const path = searchParams.get('path');
	if (path) settings.targetEndpoint.path = path;

	const status = searchParams.get('status');
	if (status) settings.targetEndpoint.status = parseInt(status);

	const userID = searchParams.get('userID');
	if (userID) {
		if (settings.targetUser === null) {
			settings.targetUser = { ipAddress: '', userID: '', composite: false };
		}
		settings.targetUser.userID = userID;
	}

	const ipAddress = searchParams.get('ipAddress');
	if (ipAddress) {
		if (settings.targetUser === null) {
			settings.targetUser = { ipAddress: '', userID: '', composite: false };
		}
		settings.targetUser.ipAddress = ipAddress;
	}

	const weekday = searchParams.get('weekday');
	if (weekday !== null) {
		const parsed = parseInt(weekday);
		if (parsed >= 0 && parsed <= 6) settings.targetWeekday = parsed;
	}

	const hour = searchParams.get('hour');
	if (hour !== null) {
		const parsed = parseInt(hour);
		if (parsed >= 0 && parsed <= 23) settings.targetHour = parsed;
	}

	const version = searchParams.get('version');
	if (version) settings.targetVersion = version;

	return settings;
}

export function initSettings(): DashboardSettings {
	return {
		disable404: false,
		ignoreBots: false,
		hostname: null,
		period: defaultPeriod,
		targetEndpoint: {
			path: null,
			status: null,
		},
		targetReferrer: null,
		targetLocation: null,
		targetWeekday: null,
		targetHour: null,
		targetVersion: null,
		targetClient: null,
		targetDeviceType: null,
		targetOS: null,
		targetUser: null,
		hiddenEndpoints: new Set(),
		ignoreParams: true,
	};
}
