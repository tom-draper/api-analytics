import type { Period } from './settings'

export function periodToDays(period: Period) {
  switch (period) {
    case '24 hours':
      return 1
    case 'Week':
      return 7
    case 'Month':
      return 30
    case '3 months':
      return 90
    case '6 months':
      return 30 * 7
    case 'Year':
      return 365
    default:
      return null
  }
}

export function periodToMarkers(period: string) {
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