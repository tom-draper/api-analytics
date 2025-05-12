export type Period =
	| '24 hours'
	| 'week'
	| 'month'
	| '6 months'
	| 'year'
	| 'all time';

export type MonitorPeriod = '24h' | '7d' | '30d' | '60d'

export const defaultPeriod: Period = 'week';

export function periodToDays(period: Period): number | null {
	switch (period) {
		case '24 hours':
			return 1;
		case 'week':
			return 7;
		case 'month':
			return 30;
		case '6 months':
			return 30 * 7;
		case 'year':
			return 365;
		default:
			return null;
	}
}

export function periodToMarkers(period: MonitorPeriod): number | null {
	switch (period) {
		case '24h':
			return 38;
		case '7d':
			return 84;
		case '30d':
		case '60d':
			return 120;
		default:
			return null;
	}
}

export function dateInPeriod(date: Date, period: Period) {
	if (period === 'all time') {
		return true;
	}
	const days = periodToDays(period);
	if (days === null) {
		return true;
	}
	const periodAgo = new Date();
	periodAgo.setDate(periodAgo.getDate() - days);
	return date > periodAgo;
}

export function dateInPrevPeriod(date: Date, period: Period) {
	if (period === 'all time') {
		return true;
	}
	const days = periodToDays(period);
	if (days === null) {
		return true;
	}
	const startPeriodAgo = new Date();
	startPeriodAgo.setDate(startPeriodAgo.getDate() - days * 2);
	const endPeriodAgo = new Date();
	endPeriodAgo.setDate(endPeriodAgo.getDate() - days);
	return startPeriodAgo < date && date < endPeriodAgo;
}