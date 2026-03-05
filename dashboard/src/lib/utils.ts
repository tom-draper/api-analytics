import { periodToDays, type Period } from '$lib/period';

export function getPercentageChange(count: number, prevCount: number): number | null {
	if (prevCount === 0) return null;
	return (count / prevCount) * 100 - 100;
}

export function countPerHour(count: number, period: Period, firstDate: Date | null, lastDate: Date | null): number {
	let days = periodToDays(period);
	if (days === null && firstDate && lastDate) {
		days = Math.floor((lastDate.getTime() - firstDate.getTime()) / 86_400_000);
	}
	if (!days) return count;
	return count / (24 * days);
}
