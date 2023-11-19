import { periodToDays } from "./period";
import type { Period } from "./settings";

 export function initFreqMap(period: Period, genValue: () => any) {
    // Populate requestFreq with zeros across date period
    const freq: Map<number, ReturnType<typeof genValue>> = new Map();
    const days = periodToDays(period);
    if (days === null) {
      return freq;
    }

    if (days <= 7) {
      // Freq count for every 5 minute in date range
      for (let i = 0; i < 288; i++) {
        const date = new Date();
        date.setSeconds(0, 0);
        // Round down to multiple of 5
        date.setMinutes(Math.floor(date.getMinutes() / 5) * 5 - i * 5);
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