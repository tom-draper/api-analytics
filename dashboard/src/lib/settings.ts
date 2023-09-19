export type Period = "24 hours" | "Week" | "Month" | "3 months" | "6 months" | "Year" | "All time"

export type DashboardSettings = {
    disable404: boolean
    hostname: string
    period: Period
    endpoint: string
}

export function initSettings(): DashboardSettings {
    return {
        disable404: false,
        hostname: null,
        period: "Month",
        endpoint: null,
    }
}