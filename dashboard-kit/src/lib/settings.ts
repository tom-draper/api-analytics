import { defaultPeriod, type Period } from "./period";

export type DashboardSettings = {
	disable404: boolean;
	hostname: string | null;
	period: Period;
	targetEndpoint: {
		path: string | null;
		status: number | null;
	};
	targetLocation: string | null;
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
		hostname: null,
		period: defaultPeriod,
		targetEndpoint: {
			path: null,
			status: null,
		},
		targetLocation: null,
		targetUser: null,
		hiddenEndpoints: new Set(),
		ignoreParams: true,
	};
}

