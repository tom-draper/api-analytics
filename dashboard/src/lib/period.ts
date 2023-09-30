import type { Period } from './settings'

export default function periodToDays(period: Period): number {
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
