import { defaultPeriod, type Period } from "./period";

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
	targetVersion: string | null;
	targetUser: {
		ipAddress: string,
		userID: string,
		composite: boolean
	} | null;
	hiddenEndpoints: Set<string>;
	ignoreParams: boolean;
};

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
		targetVersion: null,
		targetUser: null,
		hiddenEndpoints: new Set(),
		ignoreParams: true,
	};
}
