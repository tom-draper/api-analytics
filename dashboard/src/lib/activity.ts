import { periodToDays } from './period';
import type { Period } from './settings';

export function initFreqMap(period: Period, genValue: () => any) {
	// Populate requestFreq with zeros across date period
	const freq: Map<number, ReturnType<typeof genValue>> = new Map();
	const days = periodToDays(period);
	if (days === null) {
		return freq;
	}

	if (days === 1) {
		// Freq count for every 5 minute in date range
		for (let i = 0; i < 4 * 24; i++) {
			const date = new Date();
			date.setSeconds(0, 0);
			// Round down to multiple of 5
			date.setMinutes(Math.floor(date.getMinutes() / 15) * 15 - i * 15);
			freq.set(date.getTime(), genValue());
		}
	} if (days === 7) {
		// Freq count for every 5 minute in date range
		for (let i = 0; i < 24 * 7; i++) {
			const date = new Date();
			date.setSeconds(0, 0);
			// Round down to multiple of 15
			date.setMinutes(Math.floor(date.getMinutes() / 60) * 60 - i * 60);
			freq.set(date.getTime(), genValue());
		}
	} else {
		// Freq count for every day in date range
		for (let i = 0; i < days; i++) {
			const date = new Date();
			date.setHours(0, 0, 0, 0);
			date.setDate(date.getDate() - i);
			freq.set(date.getTime(), genValue());
		}
	}

	return freq;
}
