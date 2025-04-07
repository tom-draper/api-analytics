import type { Period } from "./period";


export function periodParamToPeriod(period: string): Period {
    switch (period) {
        case '24-hours':
            return '24 hours';
        case 'week':
            return 'Week';
        case 'month':
            return 'Month';
        case '3-months':
            return '3 months';
        case '6-months':
            return '6 months';
        case 'year':
            return 'Year';
        case 'all-time':
            return 'All time';
        default:
            return 'Week';
    }
}

export function periodToParamString(period: Period) {
    switch (period) {
        case '24 hours':
            return '24-hours';
        case 'Week':
            return 'week';
        case 'Month':
            return 'month';
        case '3 months':
            return '3-months';
        case '6 months':
            return '6-months';
        case 'Year':
            return 'year';
        case 'All time':
            return 'all-time';
        default:
            return 'week';
    }
}