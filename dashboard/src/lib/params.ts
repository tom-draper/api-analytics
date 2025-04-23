import type { Period } from "./period";


export function periodParamToPeriod(period: string): Period {
    switch (period) {
        case '24-hours':
            return '24 hours';
        case 'week':
            return 'week';
        case 'month':
            return 'month';
        case '3-months':
            return '3 months';
        case '6-months':
            return '6 months';
        case 'year':
            return 'year';
        case 'all-time':
            return 'all time';
        default:
            return 'week';
    }
}

export function periodToParamString(period: Period) {
    switch (period) {
        case '24 hours':
            return '24-hours';
        case 'week':
            return 'week';
        case 'month':
            return 'month';
        case '3 months':
            return '3-months';
        case '6 months':
            return '6-months';
        case 'year':
            return 'year';
        case 'all time':
            return 'all-time';
        default:
            return 'week';
    }
}